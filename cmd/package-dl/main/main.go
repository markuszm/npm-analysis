package main

import (
	"io/ioutil"
	"log"
	"github.com/markuszm/npm-analysis/downloader"
	"os"
	"os/signal"
	"sync"
	"syscall"

	"fmt"
	"github.com/buger/jsonparser"
	"io"
	"github.com/markuszm/npm-analysis/database/model"
	"path"
	"strings"
)

const PATH_TO_NPM_JSON = "/home/markus/npm-analysis/npm_download_shasum.json"

const DOWNLOAD_PATH = "/media/markus/NPM/NPM"

const workerNumber = 10

var notFoundPackages strings.Builder

func main() {
	files, readErr := ioutil.ReadFile(PATH_TO_NPM_JSON)

	if readErr != nil {
		panic("Read error")
	}

	workerWait := sync.WaitGroup{}

	stop := make(chan bool, 1)

	jobs := make(chan model.Dist)

	notFoundPackages = strings.Builder{}

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

	errFile, _ := os.Create(path.Join(DOWNLOAD_PATH, "notFound.txt"))
	defer errFile.Close()
	io.Copy(errFile, strings.NewReader(notFoundPackages.String()))

	log.Println("Finished Downloading")
}

func extractTarballs(data []byte, jobs chan model.Dist, stop chan bool) {
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

			checksumValue, _, _, parseErr := jsonparser.Get(value, "shasum")
			if parseErr != nil {
				log.Fatal(parseErr)
			}

			url := string(tarballValue)
			checksum := string(checksumValue)
			pkg := model.Dist{Url: url, Shasum: checksum}
			jobs <- pkg
		}
	})
	close(jobs)
	log.Println("closed jobs")
}

func worker(id int, jobs chan model.Dist, workerWait *sync.WaitGroup) {
	for j := range jobs {
		err := downloader.DownloadPackage(DOWNLOAD_PATH, j)
		if err != nil {
			if err.Error() == "Not Found" {
				notFoundPackages.WriteString(fmt.Sprintf("%s \n", j))
				continue
			}
			log.Println(err)
			// this only works when there are more than 1 worker
			jobs <- j
			continue
		}
		log.Println("worker", id, "finished job", j)
	}
	workerWait.Done()
	log.Println("send finished worker ", id)
}
