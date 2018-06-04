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

	packages, err := database.GetPackages(mysql)
	if err != nil {
		log.Fatalf("ERROR: loading packages from mysql with %v", err)
	}

	log.Print("Finished retrieving packages from db")

	err = database.CreateVersionCount(mysql)
	if err != nil {
		log.Fatalf("ERROR: creating table with %v", err)
	}

	workerWait := sync.WaitGroup{}

	jobs := make(chan string, 100)

	for w := 1; w <= workerNumber; w++ {
		workerWait.Add(1)
		go worker(w, jobs, &workerWait)
	}

	for i, p := range packages {
		if i%10000 == 0 {
			log.Printf("Finished %v packages", i)
		}
		jobs <- p
	}

	close(jobs)
	workerWait.Wait()
}

func worker(id int, jobs chan string, workerWait *sync.WaitGroup) {
	for p := range jobs {
		versionChanges, err := database.GetVersionChangesForPackage(p, db)
		if err != nil {
			log.Fatalf("ERROR: retrieving version changes for package %v with %v", p, err)
		}

		evolution.SortVersionChange(versionChanges)

		versionCount := evolution.CountVersions(versionChanges)

		err = insert.StoreVersionCount(db, p, versionCount)
		if err != nil {
			log.Fatalf("ERROR: writing to database with %v", err)
		}
	}
	workerWait.Done()
}
