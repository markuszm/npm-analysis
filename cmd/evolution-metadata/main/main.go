package main

import (
	"encoding/json"
	"github.com/markuszm/npm-analysis/database/evolution"
	"github.com/markuszm/npm-analysis/database/model"
	"github.com/markuszm/npm-analysis/util"
	"io/ioutil"
	"log"
	"os"
	"time"
)

const MONGOURL = "mongodb://npm:npm123@localhost:27017"

func main() {
	mongoDB := evolution.NewMongoDB(MONGOURL, "npm", "packages")

	mongoDB.Connect()
	defer mongoDB.Disconnect()

	startTime := time.Now()

	allDocs, err := mongoDB.FindAll()
	if err != nil {
		log.Fatalf("ERROR: %v", err)
	}

	endTime := time.Now()

	log.Printf("Took %v seconds to get all Documents from MongoDB", endTime.Sub(startTime).Seconds())

	sumVersions := 0

	startTime = time.Now()

	for _, doc := range allDocs {
		val, err := util.Decompress(doc.Value)
		if err != nil {
			log.Fatalf("ERROR Decompressing: %v", err)
		}

		if val == "" {
			log.Printf("WARNING: empty metadata in package %v", doc.Key)
			continue
		}

		metadata := Metadata{}

		err = json.Unmarshal([]byte(val), &metadata)
		if err != nil {
			ioutil.WriteFile("/home/markus/npm-analysis/error.json", []byte(val), os.ModePerm)
			log.Fatalf("ERROR Unmarshalling: %v", err)
		}

		sumVersions += len(metadata.Versions)
	}

	endTime = time.Now()

	log.Printf("%v Versions", sumVersions)
	log.Printf("Took %v seconds to parse all documents", endTime.Sub(startTime).Seconds())

}

type Metadata struct {
	Name     string                         `json:"name"`
	Versions map[string]model.PackageLegacy `json:"versions"`
	Time     map[string]interface{}         `json:"time"`
}
