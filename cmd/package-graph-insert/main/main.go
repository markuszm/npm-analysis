package main

import (
	"fmt"
	"log"
	"npm-analysis/database"
	"npm-analysis/database/graph"
	"sync"
)

const NEO4J_URL = "bolt://neo4j:npm@localhost:7687"

const MYSQL_USER = "root"

const MYSQL_PW = "npm-analysis"

const workerNumber = 100

func main() {
	mysqlInitializer := &database.Mysql{}
	mysql, databaseInitErr := mysqlInitializer.InitDB(fmt.Sprintf("%s:%s@/npm?charset=utf8mb4&collation=utf8mb4_bin", MYSQL_USER, MYSQL_PW))
	if databaseInitErr != nil {
		log.Fatal(databaseInitErr)
	}

	InitGraphDatabase()

	count := 0

	workerWait := sync.WaitGroup{}

	jobs := make(chan string, 100)

	for w := 1; w <= workerNumber; w++ {
		workerWait.Add(1)
		go worker(w, jobs, &workerWait)
	}

	packages, retrieveErr := database.GetPackages(mysql)
	if retrieveErr != nil {
		log.Fatal(retrieveErr)
	}

	for _, pkg := range packages {
		jobs <- pkg
	}

	close(jobs)

	log.Println(count)

	workerWait.Wait()
}

func InitGraphDatabase() {
	neo4JDatabase := graph.NewNeo4JDatabase()
	initErr := neo4JDatabase.InitDB(NEO4J_URL)
	if initErr != nil {
		log.Fatal(initErr)
	}

	graph.Init(neo4JDatabase)

	neo4JDatabase.Close()
}

func worker(workerId int, jobs chan string, workerWait *sync.WaitGroup) {
	neo4JDatabase := graph.NewNeo4JDatabase()
	initErr := neo4JDatabase.InitDB(NEO4J_URL)
	if initErr != nil {
		log.Fatal(initErr)
	}

	for j := range jobs {
		insertErr := graph.InsertPackage(neo4JDatabase, j)
		if insertErr != nil {
			log.Println("ERROR:", insertErr, "with job", j)
			// could access failure code from neo4j with .(messages.FailureMessage).Metadata["code"]
			jobs <- j
		}
		log.Println("worker", workerId, "finished job", j)
	}
	workerWait.Done()
	log.Println("send finished worker ", workerId)

	neo4JDatabase.Close()
}
