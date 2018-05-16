package main

import (
	"fmt"
	"github.com/markuszm/npm-analysis/database"
	"github.com/markuszm/npm-analysis/database/evolution"
	"log"
	"sync"
)

const MYSQL_USER = "root"

const MYSQL_PW = "npm-analysis"

const MONGOURL = "mongodb://npm:npm123@localhost:27017"

const workerNumber = 25

func main() {
	mysqlInitializer := &database.Mysql{}
	mysql, databaseInitErr := mysqlInitializer.InitDB(fmt.Sprintf("%s:%s@/npm?charset=utf8mb4&collation=utf8mb4_bin", MYSQL_USER, MYSQL_PW))
	if databaseInitErr != nil {
		log.Fatal(databaseInitErr)
	}

	count := 0

	workerWait := sync.WaitGroup{}

	jobs := make(chan string, 100)

	for w := 1; w <= workerNumber; w++ {
		workerWait.Add(1)
		go worker(w, jobs, &workerWait)
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

	log.Println(count)

	workerWait.Wait()
}

func worker(workerId int, jobs chan string, workerWait *sync.WaitGroup) {
	mongoDB := evolution.NewMongoDB(MONGOURL, "npm", "packages")
	mongoDB.Connect()
	defer mongoDB.Disconnect()

	log.Printf("logged in mongo - workerId %v", workerId)

	for pkg := range jobs {
		if val, err := mongoDB.FindOneSimple("key", pkg); val != "" && err == nil {
			log.Printf("Package %v already exists", pkg)
			continue
		}

		doc, err := evolution.GetMetadataFromNpm(pkg)
		if err != nil {
			log.Printf("ERROR: %v", err)
			jobs <- pkg
		}

		data, err := evolution.Compress(doc)
		if err != nil {
			log.Fatalf(err.Error())
		}

		mongoDB.InsertOneSimple(pkg, data)
		if err != nil {
			log.Fatalf("ERROR: inserting %v into mongo with %s", pkg, err)
		}

		log.Printf("Inserted package metadata of %v worker: %v", pkg, workerId)
	}

	workerWait.Done()
	log.Println("send finished worker ", workerId)
}
