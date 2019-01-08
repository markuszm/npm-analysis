package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"github.com/markuszm/npm-analysis/downloader"
	"github.com/markuszm/npm-analysis/model"
	"github.com/markuszm/npm-analysis/storage"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"path"
	"strconv"
	"strings"
	"sync"
)

const s3BucketName = "455877074454-npm-packages"

const lastSeqFile = "/tmp/lastseq"

const registryUrl = "http://registry.npmjs.org"

var workerNumber = 20

func main() {
	since := flag.Int("since", 6087177, "since which sequence to track changes")
	pruneFlag := flag.Bool("prune", false, "whether to prune s3 bucket")
	workerFlag := flag.Int("worker", 20, "number of workers")
	flag.Parse()

	workerNumber = *workerFlag

	if *pruneFlag {
		prune()
	} else {
		download(*since)
	}
}

func prune() {
	s3Client := storage.NewS3Client("us-east-1")

	keys := make(chan string, 100)

	workerWait := sync.WaitGroup{}

	// start workers
	go s3Client.StreamBucketObjects(s3BucketName, keys)

	for w := 1; w <= workerNumber; w++ {
		workerWait.Add(1)
		go worker(w, keys, &workerWait, s3Client)
	}

	// wait for workers to finish
	workerWait.Wait()
}

func worker(id int, jobs chan string, workerWait *sync.WaitGroup, s3Client *storage.S3Client) {
	for j := range jobs {
		deleteIfExists(j, s3Client)
	}
	workerWait.Done()
	log.Println("send finished worker ", id)
}

func deleteIfExists(key string, s3Client *storage.S3Client) {
	fileName := path.Base(key)
	sep := strings.LastIndex(fileName, "-")
	extSep := strings.Index(fileName, ".tgz")
	packageName := fileName[0:sep]
	version := fileName[sep+1 : extSep]
	downloadUrl := downloader.GenerateDownloadUrl(model.PackageVersionPair{
		Name:    packageName,
		Version: version,
	}, registryUrl)
	if packageExists(downloadUrl) {
		err := s3Client.DeleteObject(s3BucketName, key)
		if err != nil {
			log.Printf("error deleting key: %s with err: %s", key, err.Error())
		}
		log.Printf("deleted key: %s", key)
	}
}

func packageExists(url string) bool {
	resp, requestErr := http.Head(url)
	if requestErr != nil || resp.StatusCode == http.StatusNotFound {
		return false
	}
	return true
}

func download(since int) {
	lastSeqFileBytes, err := ioutil.ReadFile(lastSeqFile)
	if err == nil {
		toInt, err := strconv.Atoi(string(lastSeqFileBytes))
		if err == nil {
			since = int(toInt)
		}
	}

	s3Client := storage.NewS3Client("us-east-1")

	url := fmt.Sprintf("https://replicate.npmjs.com/_changes?include_docs=true&feed=continuous&since=%v&heartbeat=10000", since)

	log.Printf("Using replicate url: %v", url)

	resp, err := http.Get(url)
	if err != nil {
		log.Fatal(err)
	}

	defer resp.Body.Close()

	decoder := json.NewDecoder(resp.Body)

	lastSeq := since

	for {
		value := Value{}
		err := decoder.Decode(&value)
		if err != nil {
			if strings.Contains(err.Error(), "EOF") {
				err := ioutil.WriteFile(lastSeqFile, []byte(strconv.Itoa(lastSeq)), os.ModePerm)
				log.Fatalf("EOF restarting with latest sequence stored")
				if err != nil {
					log.Print("ERROR writing lastSeq")
				}
			} else {
				log.Printf("ERROR: parsing metadata for id after last id %v with error %v", lastSeq, err)
				err = ioutil.WriteFile(lastSeqFile, []byte(strconv.Itoa(lastSeq+1)), os.ModePerm)
				log.Fatalf("EOF restarting with latest sequence stored")
			}
		}

		latest := value.Document.DistTag["latest"]
		tarball := value.Document.Versions[latest].Dist.Url
		checksum := value.Document.Versions[latest].Dist.Shasum

		packageFilePath, err := downloader.GeneratePackageFullPath(tarball)
		if err != nil {
			log.Printf("ERROR: Cannot get package file name for Seq: %v, Package %v", value.Seq, value.Name)
			continue
		}

		_, err = s3Client.HeadObject(s3BucketName, packageFilePath)
		if err == nil {
			log.Printf("Seq: %v, Already retrieved package %v", value.Seq, value.Name)
			continue
		}

		downloadPath := "/tmp/"
		filePath, err := downloader.DownloadPackageAndVerify(downloadPath, tarball, checksum)
		if err != nil {
			if err.Error() == "Not Found" {
				log.Printf("WARNING: Seq: %v, did not find package %v", value.Seq, value.Name)
				lastSeq = int(value.Seq)
				continue
			}
			log.Printf("ERROR: %v", err)
			continue
		}

		file, err := os.Open(filePath)
		if err != nil {
			log.Printf("ERROR: %v", err)
			continue
		}

		err = s3Client.PutObject(s3BucketName, packageFilePath, file)
		if err != nil {
			log.Printf("ERROR: %v", err)
		}

		log.Printf("Uploaded: Seq: %v, Id: %v, Version: %v", value.Seq, value.Name, latest)

		file.Close()

		os.Remove(filePath)

		lastSeq = int(value.Seq)
	}
}

type Version struct {
	Dist      model.Dist `json:"dist"`
	DistWrong string     `json:"DIST"`
}

type Document struct {
	DistTag  map[string]string  `json:"dist-tags"`
	Versions map[string]Version `json:"versions"`
}

type Value struct {
	Seq      int64    `json:"seq"`
	Name     string   `json:"id"`
	Document Document `json:"doc"`
}
