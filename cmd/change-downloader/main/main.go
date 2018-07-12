package main

import (
	"encoding/json"
	"github.com/markuszm/npm-analysis/downloader"
	"log"
	"net/http"

	"flag"
	"fmt"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3"
	"io/ioutil"
	"os"
	"strconv"
	"strings"
)

const s3BucketName = "455877074454-npm-packages"

const lastSeqFile = "/tmp/lastseq"

func main() {
	since := flag.Int("since", 5506500, "since which sequence to track changes")
	flag.Parse()

	if _, err := os.Stat(lastSeqFile); err == nil {
		bytes, err := ioutil.ReadFile(lastSeqFile)
		if err == nil {
			lastSeq, err := strconv.Atoi(strings.Trim(string(bytes), "\n"))
			if err == nil {
				since = &lastSeq
			}
		}
	}

	sess := session.Must(session.NewSession(&aws.Config{
		Region: aws.String("us-east-1"),
	}))

	svc := s3.New(sess)

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
			ioutil.WriteFile(lastSeqFile, []byte(strconv.Itoa(lastSeq)), os.ModePerm)
			log.Fatal("EOF restarting with latest sequence stored")
		}
		latest := value.Document.DistTag["latest"]
		tarball := value.Document.Versions[latest].Dist.Tarball
		checksum := value.Document.Versions[latest].Dist.Checksum

		packageFilePath, err := downloader.GeneratePackageFullPath(tarball)
		if err != nil {
			log.Printf("ERROR: Cannot get package file name for Seq: %v, Package %v", value.Seq, value.Name)
			continue
		}

		headObjectInput := s3.HeadObjectInput{
			Bucket: aws.String(s3BucketName),
			Key:    aws.String(packageFilePath),
		}

		_, err = svc.HeadObject(&headObjectInput)
		if err == nil {
			log.Printf("Seq: %v, Already retrieved package %v", value.Seq, value.Name)
			lastSeq = int(value.Seq)
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

		putObjectInput := s3.PutObjectInput{
			Body:   file,
			Bucket: aws.String(s3BucketName),
			Key:    aws.String(strings.Replace(filePath, downloadPath, "", 1)),
		}
		_, err = svc.PutObject(&putObjectInput)
		if err != nil {
			log.Printf("ERROR: %v", err)
		}

		log.Printf("Uploaded: Seq: %v, Id: %v, Version: %v", value.Seq, value.Name, latest)

		os.Remove(filePath)

		lastSeq = int(value.Seq)
	}
}

type Dist struct {
	Tarball  string `json:"tarball"`
	Checksum string `json:"shasum"`
}

type Version struct {
	Dist Dist `json:"dist"`
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
