package main

import (
	"flag"
	"fmt"
	"github.com/markuszm/npm-analysis/database"
	"github.com/markuszm/npm-analysis/database/graph"
	"github.com/markuszm/npm-analysis/database/model"
	"log"
	"sync"
)

const NEO4J_URL = "bolt://neo4j:npm@localhost:7687"

const MYSQL_USER = "root"

const MYSQL_PW = "npm-analysis"

const workerNumber = 100

var depType string

func main() {
	depTypeFlag := flag.String("type", "dependencies", "specify which type of dependency to insert")
	flag.Parse()

	depType = *depTypeFlag

	mysqlInitializer := &database.Mysql{}
	mysql, databaseInitErr := mysqlInitializer.InitDB(fmt.Sprintf("%s:%s@/npm?charset=utf8mb4&collation=utf8mb4_bin", MYSQL_USER, MYSQL_PW))
	if databaseInitErr != nil {
		log.Fatal(databaseInitErr)
	}

	InitGraphDatabase()

	count := 0

	workerWait := sync.WaitGroup{}

	jobs := make(chan model.Dependency, 100)

	for w := 1; w <= workerNumber; w++ {
		workerWait.Add(1)
		go worker(w, jobs, &workerWait)
	}

	dependencies, retrieveErr := database.GetDependencies(mysql, depType)

	if retrieveErr != nil {
		log.Fatal(retrieveErr)
	}

	for _, dep := range dependencies {
		jobs <- dep
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

func worker(workerId int, jobs chan model.Dependency, workerWait *sync.WaitGroup) {
	neo4JDatabase := graph.NewNeo4JDatabase()
	initErr := neo4JDatabase.InitDB(NEO4J_URL)
	if initErr != nil {
		log.Fatal(initErr)
	}

	for j := range jobs {
		var insertErr error
		switch depType {
		case "dependencies":
			insertErr = graph.InsertDependency(neo4JDatabase, j)
		case "bundledDependencies":
			insertErr = graph.InsertBundledDependency(neo4JDatabase, j)
		case "devDependencies":
			insertErr = graph.InsertDevDependency(neo4JDatabase, j)
		case "optionalDependencies":
			insertErr = graph.InsertOptionalDependency(neo4JDatabase, j)
		case "peerDependencies":
			insertErr = graph.InsertPeerDependency(neo4JDatabase, j)
		}
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
