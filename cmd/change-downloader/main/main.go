package main

import (
	"encoding/json"
	"github.com/markuszm/npm-analysis/downloader"
	"log"
	"net/http"
	"strings"

	"flag"
	"fmt"
	"github.com/markuszm/npm-analysis/model"
	"github.com/markuszm/npm-analysis/storage"
	"io/ioutil"
	"os"
	"strconv"
)

const s3BucketName = "455877074454-npm-packages"

const lastSeqFile = "/tmp/lastseq"

func main() {
	since := flag.Int("since", 5506500, "since which sequence to track changes")
	flag.Parse()

	s3Client := storage.NewS3Client("us-east-1")

	url := fmt.Sprintf("https://replicate.npmjs.com/_changes?include_docs=true&feed=continuous&since=%v&heartbeat=10000", *since)

	log.Printf("Using replicate url: %v", url)

	resp, err := http.Get(url)
	if err != nil {
		log.Fatal(err)
	}

	defer resp.Body.Close()

	decoder := json.NewDecoder(resp.Body)

	lastSeq := *since

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
				continue
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
		defer file.Close()
		if err != nil {
			log.Printf("ERROR: %v", err)
			continue
		}

		err = s3Client.PutObject(s3BucketName, packageFilePath, file)
		if err != nil {
			log.Printf("ERROR: %v", err)
		}

		log.Printf("Uploaded: Seq: %v, Id: %v, Version: %v", value.Seq, value.Name, latest)

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
