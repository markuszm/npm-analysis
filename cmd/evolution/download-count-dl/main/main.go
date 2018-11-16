package main

import (
	"fmt"
	"log"
	"sync"

	"encoding/json"
	"flag"
	"github.com/markuszm/npm-analysis/database"
	"github.com/markuszm/npm-analysis/evolution"
	"strings"
)

const MYSQL_USER = "root"

const MYSQL_PW = "npm-analysis"

const MONGOURL = "mongodb://npm:npm123@localhost:27017"

// NPM is rate-limiting so don't go over 8 workers here
const workerNumber = 3

func main() {
	bulk := flag.Bool("bulk", true, "use bulk queries?")
	flag.Parse()

	mysqlInitializer := &database.Mysql{}
	mysql, databaseInitErr := mysqlInitializer.InitDB(fmt.Sprintf("%s:%s@/npm?charset=utf8mb4&collation=utf8mb4_bin", MYSQL_USER, MYSQL_PW))
	if databaseInitErr != nil {
		log.Fatal(databaseInitErr)
	}
	defer mysql.Close()

	workerWait := sync.WaitGroup{}

	jobs := make(chan string, 100)

	if *bulk {
		log.Printf("Using bulk queries")
	} else {
		log.Printf("Using single package queries")
	}

	for w := 1; w <= workerNumber; w++ {
		workerWait.Add(1)
		if *bulk {
			go workerBulk(w, jobs, &workerWait)
		} else {
			go worker(w, jobs, &workerWait)
		}
	}

	log.Println("Loading packages from database")

	packages, err := database.GetPackages(mysql)
	if err != nil {
		log.Fatal("cannot load packages from mysql")
	}

	for _, pkg := range packages {
		jobs <- pkg
	}

	close(jobs)

	workerWait.Wait()
}

func worker(workerId int, jobs chan string, workerWait *sync.WaitGroup) {
	mongoDB := database.NewMongoDB(MONGOURL, "npm", "downloads")
	mongoDB.Connect()
	defer mongoDB.Disconnect()
	log.Printf("logged in mongo - workerId %v", workerId)

	ensureIndex(mongoDB)

	for pkg := range jobs {
		if val, err := mongoDB.FindOneSimple("key", pkg); val != "" && err == nil {
			downloadCount := evolution.DownloadCountResponse{}
			err := json.Unmarshal([]byte(val), &downloadCount)
			if err != nil {
				log.Fatalf("ERROR: Cannot unmarshal DownloadCountResponse from db with error %v", err)
			}
			if len(downloadCount.DownloadCounts) == 1190 {
				//log.Printf("Package %v already exists", pkg)
				continue
			} else {
				err := mongoDB.RemoveWithKey(pkg)
				if err != nil {
					log.Fatalf("ERROR: could not remove already existing but wrong data for package %v", pkg)
				}
			}
		}

		doc, err := evolution.GetDownloadCountsForPackage(pkg)
		if err != nil {
			log.Printf("ERROR: %v", err)
			jobs <- pkg
		}

		mongoDB.InsertOneSimple(pkg, doc)
		if err != nil {
			log.Fatalf("ERROR: inserting %v into mongo with %s", pkg, err)
		}

		log.Printf("Inserted download counts of %v worker: %v", pkg, workerId)
	}

	workerWait.Done()
	log.Println("send finished worker ", workerId)
}

func workerBulk(workerId int, jobs chan string, workerWait *sync.WaitGroup) {
	mongoDB := database.NewMongoDB(MONGOURL, "npm", "downloads")
	mongoDB.Connect()
	defer mongoDB.Disconnect()
	log.Printf("logged in mongo - workerId %v", workerId)

	ensureIndex(mongoDB)

	var bulkPackages []string

	for pkg := range jobs {
		if val, err := mongoDB.FindOneSimple("key", pkg); val != "" && err == nil {
			downloadCount := evolution.DownloadCountResponse{}
			err := json.Unmarshal([]byte(val), &downloadCount)
			if err != nil {
				log.Fatalf("ERROR: Cannot unmarshal DownloadCountResponse from db with error %v", err)
			}
			if len(downloadCount.DownloadCounts) == 1190 {
				//log.Printf("Package %v already exists", pkg)
				continue
			} else {
				err := mongoDB.RemoveWithKey(pkg)
				if err != nil {
					log.Fatalf("ERROR: could not remove already existing but wrong data for package %v", pkg)
				}
			}
		}

		if strings.HasPrefix(pkg, "@") {
			log.Printf("WARNING: Package %v unsupported for bulk download", pkg)
			continue
		}

		bulkPackages = append(bulkPackages, pkg)

		if len(bulkPackages) == 128 {
			processBulkPackages(bulkPackages, mongoDB, workerId)

			bulkPackages = make([]string, 0)
		}
	}

	if len(bulkPackages) > 0 {
		processBulkPackages(bulkPackages, mongoDB, workerId)
	}

	workerWait.Done()
	log.Println("send finished worker ", workerId)
}

func processBulkPackages(bulkPackages []string, mongoDB *database.MongoDB, workerId int) {
	bulk, err := evolution.GetDownloadCountsBulk(bulkPackages)
	if err != nil {
		log.Fatalf("ERROR: %v", err)
	}
	for p, d := range bulk {
		mongoDB.InsertOneSimple(p, d)
		if err != nil {
			log.Fatalf("ERROR: inserting %v into mongo with %s", p, err)
		}

		log.Printf("Inserted download counts of %v worker: %v", p, workerId)
	}
}

func ensureIndex(mongoDB *database.MongoDB) {
	indexResp, err := mongoDB.EnsureSingleIndex("key")
	if err != nil {
		log.Fatalf("Index cannot be created with ERROR: %v", err)
	}
	log.Printf("Index created %v", indexResp)
}
