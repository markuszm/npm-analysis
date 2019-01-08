package cmd

import (
	"fmt"
	"github.com/buger/jsonparser"
	"github.com/markuszm/npm-analysis/downloader"
	"github.com/markuszm/npm-analysis/model"
	"github.com/spf13/cobra"
	"io"
	"io/ioutil"
	"log"
	"os"
	"os/signal"
	"path"
	"strings"
	"sync"
	"syscall"
)

var packageDownloadPathToNpmJson string
var packageDownloadPath string

var packageDownloadWorkerNumber int

var notFoundPackages strings.Builder

func init() {
	rootCmd.AddCommand(packageDownloadCmd)

	packageDownloadCmd.Flags().StringVar(&packageDownloadPath, "path", "/media/markus/NPM/NPM", "destination path to download packages")
	packageDownloadCmd.Flags().StringVar(&packageDownloadPathToNpmJson, "source", "/home/markus/npm-analysis/npm_download_shasum.json", "path to JSON containing tar url and checksum for all packages (downloaded from registry)")
	packageDownloadCmd.Flags().IntVar(&packageDownloadWorkerNumber, "workers", 10, "number of workers")
}

var packageDownloadCmd = &cobra.Command{
	Use:   "download",
	Short: "Download packages",
	Long:  `...`,
	Run: func(cmd *cobra.Command, args []string) {
		files, readErr := ioutil.ReadFile(packageDownloadPathToNpmJson)

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
		for w := 1; w <= packageDownloadWorkerNumber; w++ {
			workerWait.Add(1)
			go packageDownloadWorker(w, jobs, &workerWait)
		}

		go extractTarballs(files, jobs, stop)

		// wait for workers to finish
		workerWait.Wait()

		errFile, _ := os.Create(path.Join(packageDownloadPath, "notFound.txt"))
		defer errFile.Close()
		io.Copy(errFile, strings.NewReader(notFoundPackages.String()))

		log.Println("Finished Downloading")
	},
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

func packageDownloadWorker(id int, jobs chan model.Dist, workerWait *sync.WaitGroup) {
	for j := range jobs {
		_, err := downloader.DownloadPackageAndVerify(packageDownloadPath, j.Url, j.Shasum)
		if err != nil {
			if err.Error() == "Not Found" {
				notFoundPackages.WriteString(fmt.Sprintf("%s \n", j))
				continue
			}
			log.Println(err)
			// this only works when there is more than 1 worker
			jobs <- j
			continue
		}
		log.Println("worker", id, "finished job", j)
	}
	workerWait.Done()
	log.Println("send finished worker ", id)
}
