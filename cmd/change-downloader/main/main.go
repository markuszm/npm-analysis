package main

import (
	"encoding/json"
	"github.com/markuszm/npm-analysis/downloader"
	"log"
	"net/http"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3"
	"os"
	"strings"
)

const url = "https://replicate.npmjs.com/_changes?include_docs=true&feed=continuous&since=5000000"

const s3BucketName = "455877074454-npm-packages"

func main() {
	sess := session.Must(session.NewSession(&aws.Config{
		Region: aws.String("us-east-1"),
	}))

	svc := s3.New(sess)

	resp, err := http.Get(url)
	if err != nil {
		log.Fatal(err)
	}

	defer resp.Body.Close()

	decoder := json.NewDecoder(resp.Body)

	for {
		value := Value{}
		err := decoder.Decode(&value)
		if err != nil {
			log.Fatal(err)
		}
		latest := value.Document.DistTag["latest"]
		tarball := value.Document.Versions[latest].Dist.Tarball
		checksum := value.Document.Versions[latest].Dist.Checksum

		downloadPath := "/tmp/"
		filePath, err := downloader.DownloadPackageAndVerify(downloadPath, tarball, checksum)
		if err != nil {
			if err.Error() == "Not Found" {
				log.Printf("WARNING: did not find package %v", value.Name)
				continue
			}
			log.Printf("ERROR: %v", err)
			continue
		}

		file, err := os.Open(filePath)
		if err != nil {
			log.Fatal(err)
		}

		headObjectInput := s3.HeadObjectInput{
			Bucket: aws.String(s3BucketName),
			Key:    aws.String(strings.Replace(filePath, downloadPath, "", 1)),
		}

		_, err = svc.HeadObject(&headObjectInput)
		if err == nil {
			log.Printf("Already retrieved package %v", value.Name)
			continue
		}

		putObjectInput := s3.PutObjectInput{
			Body:   file,
			Bucket: aws.String(s3BucketName),
			Key:    aws.String(strings.Replace(filePath, downloadPath, "", 1)),
		}
		_, err = svc.PutObject(&putObjectInput)
		if err != nil {
			log.Fatal(err)
		}

		file.Close()

		log.Printf("Uploaded: Seq: %v, Id: %v, Version: %v", value.Seq, value.Name, latest)

		os.Remove(filePath)
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
