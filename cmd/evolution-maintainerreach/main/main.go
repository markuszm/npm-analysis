package main

import (
	"context"
	"encoding/json"
	"github.com/markuszm/npm-analysis/database"
	"github.com/mongodb/mongo-go-driver/bson"
	"github.com/remeh/sizedwaitgroup"
	"io/ioutil"
	"log"
	"os"
	"sync"
	"time"
)

const MONGOURL = "mongodb://npm:npm123@localhost:27017"

const JSONPATH = "./db-data/dependenciesTimeline.json"

const maxGoroutines = 100

func main() {
	calculatePackageReach()
}

func calculatePackageReach() {
	log.Print("Loading json")

	bytes, err := ioutil.ReadFile(JSONPATH)
	if err != nil {
		log.Fatal(err)
	}

	var dependenciesTimeline map[time.Time]map[string]map[string]bool

	err = json.Unmarshal(bytes, &dependenciesTimeline)
	if err != nil {
		log.Fatal(err)
	}

	log.Print("Finished loading json")

	someDate := time.Date(2018, time.Month(4), 1, 0, 0, 0, 0, time.UTC)

	packageMap := dependenciesTimeline[someDate]

	packages := sync.Map{}

	workerWait := sizedwaitgroup.New(maxGoroutines)

	workerWait.Add()
	PackageReach("lodash", packageMap, &packages, &workerWait)

	workerWait.Wait()

	length := 0

	packages.Range(func(key, value interface{}) bool {
		log.Print(key)
		length++
		return true
	})
	log.Print(length)
}

func PackageReach(pkg string, packageMap map[string]map[string]bool, packages *sync.Map, workerWait *sizedwaitgroup.SizedWaitGroup) {
	for p, deps := range packageMap {
		if deps[pkg] {
			if _, ok := packages.Load(p); !ok {
				packages.Store(p, true)
				workerWait.Add()
				go PackageReach(p, packageMap, packages, workerWait)
			}
		}
	}
	workerWait.Done()
}

func generateTimeToLatestVersion() {
	dependenciesTimeline := make(map[time.Time]map[string]map[string]bool, 0)

	mongoDB := database.NewMongoDB(MONGOURL, "npm", "timeline")

	mongoDB.Connect()
	defer mongoDB.Disconnect()

	startTime := time.Now()

	cursor, err := mongoDB.ActiveCollection.Find(context.Background(), bson.NewDocument())
	if err != nil {
		log.Fatal(err)
	}
	i := 0
	for cursor.Next(context.Background()) {
		doc, err := mongoDB.DecodeValue(cursor)
		if err != nil {
			log.Fatal(err)
		}

		var timeMap map[time.Time]SlimPackageData

		err = json.Unmarshal([]byte(doc.Value), &timeMap)
		if err != nil {
			log.Fatal(err)
		}

		for t, pkg := range timeMap {
			if dependenciesTimeline[t] == nil {
				dependenciesTimeline[t] = make(map[string]map[string]bool, 0)
			}
			if len(pkg.Dependencies) > 0 {
				dependencies := make(map[string]bool, 0)
				for _, dep := range pkg.Dependencies {
					dependencies[dep] = true
				}
				dependenciesTimeline[t][doc.Key] = dependencies
			}
		}

		if i%10000 == 0 {
			log.Printf("Finished %v packages", i)
		}
		i++
	}
	cursor.Close(context.Background())
	endTime := time.Now()
	log.Printf("Took %v minutes to process all Documents from MongoDB", endTime.Sub(startTime).Minutes())
	bytes, err := json.Marshal(dependenciesTimeline)
	if err != nil {
		log.Fatal(err)
	}
	log.Print("Finished transforming to JSON")
	ioutil.WriteFile("./db-data/dependenciesTimeline.json", bytes, os.ModePerm)
}

type SlimPackageData struct {
	Dependencies []string `json:"dependencies"`
}
