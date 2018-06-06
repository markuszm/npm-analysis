package main

import (
	"encoding/json"
	"github.com/markuszm/npm-analysis/database"
	"github.com/markuszm/npm-analysis/evolution"
	"github.com/markuszm/npm-analysis/model"
	"github.com/markuszm/npm-analysis/util"
	"io/ioutil"
	"log"
	"os"
	"sync"
	"time"
)

const MONGOURL = "mongodb://npm:npm123@localhost:27017"

var latestVersionsSync = sync.Map{}

const workerNumber = 75

func main() {
	mongoDB := database.NewMongoDB(MONGOURL, "npm", "packages")

	mongoDB.Connect()
	defer mongoDB.Disconnect()

	startTime := time.Now()

	allDocs, err := mongoDB.FindAll()
	if err != nil {
		log.Fatalf("ERROR: %v", err)
	}

	workerWait := sync.WaitGroup{}

	jobs := make(chan database.Document, 100)

	for w := 1; w <= workerNumber; w++ {
		workerWait.Add(1)
		go worker(w, jobs, &workerWait)
	}

	for _, doc := range allDocs {
		jobs <- doc
	}

	close(jobs)

	workerWait.Wait()

	endTime := time.Now()

	log.Printf("Took %v minutes to process all Documents from MongoDB", endTime.Sub(startTime).Minutes())

	latestVersions := make(map[string]string, 0)
	latestVersionsSync.Range(func(key, value interface{}) bool {
		latestVersions[key.(string)] = value.(string)
		return true
	})

	bytes, _ := json.MarshalIndent(latestVersions, "", "\t")
	ioutil.WriteFile("/home/markus/npm-analysis/latestVersionList.json", bytes, os.ModePerm)
}

func worker(id int, jobs chan database.Document, workerWait *sync.WaitGroup) {
	for j := range jobs {
		processDocument(j)
	}
	workerWait.Done()
}

func processDocument(doc database.Document) {
	val, err := util.Decompress(doc.Value)
	if err != nil {
		log.Fatalf("ERROR: Decompressing: %v", err)
	}

	if val == "" {
		log.Printf("WARNING: empty metadata in package %v", doc.Key)
		return
	}

	metadata := model.Metadata{}

	err = json.Unmarshal([]byte(val), &metadata)
	if err != nil {
		ioutil.WriteFile("/home/markus/npm-analysis/error.json", []byte(val), os.ModePerm)
		log.Fatalf("ERROR: Unmarshalling: %v", err)
	}

	packageData := evolution.GetPackageMetadataForEachMonth(metadata)

	latestVersionsSync.Store(doc.Key, packageData)
}
