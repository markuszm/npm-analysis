package main

import (
	"database/sql"
	"fmt"
	"github.com/markuszm/npm-analysis/database"
	"github.com/markuszm/npm-analysis/plots"
	"log"
	"sync"
)

const MYSQL_USER = "root"

const MYSQL_PW = "npm-analysis"

var db *sql.DB

var workerNumber = 100

func main() {
	mysqlInitializer := &database.Mysql{}
	mysql, err := mysqlInitializer.InitDB(fmt.Sprintf("%s:%s@/npm?charset=utf8mb4&collation=utf8mb4_bin", MYSQL_USER, MYSQL_PW))
	if err != nil {
		log.Fatal(err)
	}
	defer mysql.Close()

	db = mysql

	maintainerNames, err := database.GetMaintainerNames(mysql)
	if err != nil {
		log.Fatal(err)
	}

	log.Print("Retrieved maintainer names from database")

	workerWait := sync.WaitGroup{}

	jobs := make(chan string, 100)

	for w := 1; w <= workerNumber; w++ {
		workerWait.Add(1)
		go worker(w, jobs, &workerWait)
	}

	for _, m := range maintainerNames {
		jobs <- m
	}

	close(jobs)

	workerWait.Wait()
}

func worker(id int, jobs chan string, workerWait *sync.WaitGroup) {
	for j := range jobs {
		plots.CreateLinePlotForMaintainerPackageCount(j, db)
	}
	workerWait.Done()
}
