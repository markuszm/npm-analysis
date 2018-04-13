package main

import (
	"io/ioutil"
	"github.com/buger/jsonparser"
	"net/http"
	"log"
	"fmt"
)

const PATH_TO_NPM_JSON = "/home/markus/npm-analysis/npm_download.json"

var downloadSize int64 = 0

const workerNumber = 100

func main() {
	data, readErr := ioutil.ReadFile(PATH_TO_NPM_JSON)

	if readErr != nil {
		panic("Read error")
	}

	contentLengthCh := make(chan int64, 10000)

	finishedSum := make(chan bool)
	finishedWorker := make(chan bool)

	jobs := make(chan string, 10000)

	for w := 1; w <= workerNumber; w++ {
		go worker(w, jobs, contentLengthCh, finishedWorker)
	}

	go addContentSize(contentLengthCh, finishedSum)

	jsonparser.ArrayEach(data, func(value []byte, dataType jsonparser.ValueType, offset int, err error) {
		tarballValue, _, _, parseErr := jsonparser.Get(value, "tarball")
		if parseErr != nil {
			log.Fatal(parseErr)
		}

		url := string(tarballValue)
		jobs <- url
	})

	close(jobs)

	for r := 1; r <= workerNumber; r++ {
		<-finishedWorker
	}

	close(contentLengthCh)

	<-finishedSum

	println(downloadSize)
}

func worker(id int, jobs chan string, contentLengthCh chan int64, finished chan bool) {
	for j := range jobs {
		contentLength := getContentLength(j)
		contentLengthCh <- contentLength
		fmt.Println("worker", id, "finished job", j)
	}
	finished <- true
}

func getContentLength(url string) int64 {
	resp, requestErr := http.Head(url)
	if requestErr != nil {
		log.Fatal(requestErr)
	}
	contentLength := resp.ContentLength
	return contentLength
}

func addContentSize(contentSizeCh chan int64, finished chan bool) {
	for contentSize := range contentSizeCh {
		downloadSize += contentSize
	}
	finished <- true
}