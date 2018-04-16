package main

import (
	"io/ioutil"
	"github.com/buger/jsonparser"
	"log"
	"fmt"
	"npm-analysis/downloader"
)

const PATH_TO_NPM_JSON = "/home/markus/npm-analysis/npm_download.json"

const DOWNLOAD_PATH = "/media/markus/Seagate Expansion Drive/NPM"

const workerNumber = 10

func main() {
	data, readErr := ioutil.ReadFile(PATH_TO_NPM_JSON)

	if readErr != nil {
		panic("Read error")
	}

	finishedWorker := make(chan bool)

	jobs := make(chan string, 10000)

	for w := 1; w <= workerNumber; w++ {
		go worker(w, jobs, finishedWorker)
	}

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

	fmt.Println("Finished Downloading")
}

func worker(id int, jobs chan string, finished chan bool) {
	for j := range jobs {
		downloader.DownloadPackage(DOWNLOAD_PATH, j)
		fmt.Println("worker", id, "finished job", j)
	}
	finished <- true
}

