package main

import (
	"io/ioutil"
	"log"
	"npm-analysis/downloader"
	"os"
	"os/signal"
	"sync"
	"syscall"

	"github.com/buger/jsonparser"
)

const PATH_TO_NPM_JSON = "/home/markus/npm-analysis/npm_download.json"

const DOWNLOAD_PATH = "/media/markus/Seagate Expansion Drive/NPM"

const workerNumber = 10

func main() {
	files, readErr := ioutil.ReadFile(PATH_TO_NPM_JSON)

	if readErr != nil {
		panic("Read error")
	}

	workerWait := sync.WaitGroup{}

	stop := make(chan bool, 1)

	jobs := make(chan string)

	// gracefully stop downloads
	var gracefulStop = make(chan os.Signal)
	signal.Notify(gracefulStop, syscall.SIGTERM)
	signal.Notify(gracefulStop, syscall.SIGINT)
	go func() {
		<-gracefulStop
		log.Println("sigtem received")
		stop <- true
	}()

	// start workers
	for w := 1; w <= workerNumber; w++ {
		workerWait.Add(1)
		go worker(w, jobs, &workerWait)
	}

	go extractTarballs(files, jobs, stop)

	// wait for workers to finish
	workerWait.Wait()

	log.Println("Finished Downloading")
}

func extractTarballs(data []byte, jobs chan string, stop chan bool) {
	stopReceived := false
	jsonparser.ArrayEach(data, func(value []byte, dataType jsonparser.ValueType, offset int, err error) {
		select {
		case <-stop:
			stopReceived = true
			return
		default:
			if stopReceived {
				return
			}
			tarballValue, _, _, parseErr := jsonparser.Get(value, "tarball")
			if parseErr != nil {
				log.Fatal(parseErr)
			}

			url := string(tarballValue)
			jobs <- url
		}
	})
	close(jobs)
	log.Println("closed jobs")
}

func worker(id int, jobs chan string, workerWait *sync.WaitGroup) {
	for j := range jobs {
		err := downloader.DownloadPackage(DOWNLOAD_PATH, j)
		if err != nil {
			log.Println(err)
			jobs <- j
			continue
		}
		log.Println("worker", id, "finished job", j)
	}
	workerWait.Done()
	log.Println("send finished worker ", id)
}
