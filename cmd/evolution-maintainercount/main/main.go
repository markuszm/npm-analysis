package main

import (
	"database/sql"
	"fmt"
	"github.com/markuszm/npm-analysis/database"
	"github.com/markuszm/npm-analysis/database/insert"
	"github.com/markuszm/npm-analysis/evolution"
	"log"
	"sync"
)

const MYSQL_USER = "root"

const MYSQL_PW = "npm-analysis"

const workerNumber = 100

var db *sql.DB

func main() {
	mysqlInitializer := &database.Mysql{}
	mysql, databaseInitErr := mysqlInitializer.InitDB(fmt.Sprintf("%s:%s@/npm?charset=utf8mb4&collation=utf8mb4_bin", MYSQL_USER, MYSQL_PW))
	if databaseInitErr != nil {
		log.Fatal(databaseInitErr)
	}

	db = mysql

	changes, err := database.GetMaintainerChanges(mysql)
	if err != nil {
		log.Fatalf("ERROR: loading changes from mysql with %v", err)
	}

	log.Print("Finished retrieving changes from db")

	err = database.CreateMaintainerCount(db)
	if err != nil {
		log.Fatal(err)
	}

	countMap := evolution.CalculateMaintainerCounts(changes)

	workerWait := sync.WaitGroup{}

	jobs := make(chan evolution.MaintainerCount, 100)

	for w := 1; w <= workerNumber; w++ {
		workerWait.Add(1)
		go worker(w, jobs, &workerWait)
	}

	for _, maintainerCount := range countMap {
		jobs <- maintainerCount
	}

	close(jobs)
	workerWait.Wait()
}

func worker(id int, jobs chan evolution.MaintainerCount, workerWait *sync.WaitGroup) {
	for m := range jobs {
		err := insert.StoreMaintainerCount(db, m)
		if err != nil {
			log.Fatalf("ERROR: writing to database with %v", err)
		}
	}
	workerWait.Done()
}
