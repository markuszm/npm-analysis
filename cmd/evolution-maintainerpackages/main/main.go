package main

import (
	"encoding/json"
	"fmt"
	"github.com/markuszm/npm-analysis/database"
	"github.com/markuszm/npm-analysis/evolution"
	"github.com/markuszm/npm-analysis/util"
	"log"
	"sync"
	"time"
)

const MONGOURL = "mongodb://npm:npm123@localhost:27017"

const workerNumber = 75

const MYSQL_USER = "root"

const MYSQL_PW = "npm-analysis"

// Collects all packages that are maintained by a maintainer for a specific time and stores into mongo in collection "maintainerPackages"

func main() {
	mysqlInitializer := &database.Mysql{}
	mysql, databaseInitErr := mysqlInitializer.InitDB(fmt.Sprintf("%s:%s@/npm?charset=utf8mb4&collation=utf8mb4_bin", MYSQL_USER, MYSQL_PW))
	if databaseInitErr != nil {
		log.Fatal(databaseInitErr)
	}
	defer mysql.Close()

	changes, err := database.GetMaintainerChanges(mysql)
	if err != nil {
		log.Fatalf("ERROR: loading changes from mysql with %v", err)
	}

	log.Print("Finished retrieving changes from db")

	maintainedPackages := evolution.CalculateMaintainerPackages(changes)

	startTime := time.Now()

	workerWait := sync.WaitGroup{}

	jobs := make(chan StoreMaintainedPackages, 100)

	for w := 1; w <= workerNumber; w++ {
		workerWait.Add(1)
		go worker(w, jobs, &workerWait)
	}

	for _, v := range maintainedPackages {
		packageTimeline := make(map[time.Time][]string)

		for year, monthMap := range v.Packages {
			for month, packages := range monthMap {
				var keys []string
				for p, ok := range packages {
					if ok {
						keys = append(keys, p)
					}
				}
				packageTimeline[time.Date(year, time.Month(month), 1, 0, 0, 0, 0, time.UTC)] = keys
			}
		}
		s := StoreMaintainedPackages{
			Name:             v.Name,
			PackagesTimeline: packageTimeline,
		}

		jobs <- s
	}

	close(jobs)

	workerWait.Wait()

	endTime := time.Now()

	log.Printf("Took %v minutes to process all Documents from MongoDB", endTime.Sub(startTime).Minutes())
}

type StoreMaintainedPackages struct {
	Name             string                 `json:"name"`
	PackagesTimeline map[time.Time][]string `json:"packages"`
}

func worker(id int, jobs chan StoreMaintainedPackages, workerWait *sync.WaitGroup) {
	mongoDB := database.NewMongoDB(MONGOURL, "npm", "maintainerPackages")
	mongoDB.Connect()
	defer mongoDB.Disconnect()
	log.Printf("logged in mongo - workerId %v", id)

	ensureIndex(mongoDB)
	for j := range jobs {
		processDocument(j, mongoDB)
	}
	workerWait.Done()
}

func processDocument(pkgs StoreMaintainedPackages, mongoDB *database.MongoDB) {
	if val, err := mongoDB.FindOneSimple("key", pkgs.Name); val != "" && err == nil {
		log.Printf("Package %v already exists", pkgs.Name)

		val, err := util.Decompress(val)
		if err != nil {
			log.Fatalf("ERROR: Decompressing: %v", err)
		}

		if val == "" {
			err := mongoDB.RemoveWithKey(pkgs.Name)
			if err != nil {
				log.Fatalf("ERROR: could not remove already existing but wrong data for package %v", pkgs.Name)
			}
		} else {
			return
		}
	}

	bytes, err := json.Marshal(pkgs)
	if err != nil {
		log.Fatalf("ERROR: marshalling package data for %v with %v", pkgs.Name, err)
	}

	err = mongoDB.InsertOneSimple(pkgs.Name, string(bytes))
	if err != nil {
		log.Fatalf("ERROR: inserting package %v into mongo with %v", pkgs.Name, err)
	}
}

func ensureIndex(mongoDB *database.MongoDB) {
	indexResp, err := mongoDB.EnsureSingleIndex("key")
	if err != nil {
		log.Fatalf("Index cannot be created with ERROR: %v", err)
	}
	log.Printf("Index created %v", indexResp)
}
