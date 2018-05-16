package main

import (
	"flag"
	"fmt"
	"log"
	"github.com/markuszm/npm-analysis/database"
	"github.com/markuszm/npm-analysis/database/graph"
	"github.com/markuszm/npm-analysis/database/model"
	"sync"
)

const NEO4J_URL = "bolt://neo4j:npm@localhost:7687"

const MYSQL_USER = "root"

const MYSQL_PW = "npm-analysis"

const workerNumber = 100

var insertType string

func main() {
	flag.StringVar(&insertType, "insert", "author", "type to insert")

	flag.Parse()

	mysqlInitializer := &database.Mysql{}
	mysql, databaseInitErr := mysqlInitializer.InitDB(fmt.Sprintf("%s:%s@/npm?charset=utf8mb4&collation=utf8mb4_bin", MYSQL_USER, MYSQL_PW))
	if databaseInitErr != nil {
		log.Fatal(databaseInitErr)
	}

	InitGraphDatabase()

	count := 0

	workerWait := sync.WaitGroup{}

	jobs := make(chan interface{}, 100)

	for w := 1; w <= workerNumber; w++ {
		workerWait.Add(1)
		go worker(w, jobs, &workerWait)
	}

	var jobItems []interface{}
	var retrieveErr error

	switch insertType {
	case "author":
		var persons []database.Person
		persons, retrieveErr = database.GetAuthors(mysql)
		jobItems = transformPersonToInterfaceSlice(persons)
	case "maintainers":
		var persons []database.Person
		persons, retrieveErr = database.GetMaintainers(mysql)
		jobItems = transformPersonToInterfaceSlice(persons)
	case "package":
		var packages []string
		packages, retrieveErr = database.GetPackages(mysql)
		jobItems = transformStringToInterfaceSlice(packages)
	}
	if retrieveErr != nil {
		log.Fatal(retrieveErr)
	}

	for _, pkg := range jobItems {
		jobs <- pkg
	}

	close(jobs)

	log.Println(count)

	workerWait.Wait()
}

func transformPersonToInterfaceSlice(persons []database.Person) []interface{} {
	var result []interface{}
	for _, p := range persons {
		result = append(result, p)
	}
	return result
}

func transformStringToInterfaceSlice(strings []string) []interface{} {
	var result []interface{}
	for _, p := range strings {
		result = append(result, p)
	}
	return result
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

func worker(workerId int, jobs chan interface{}, workerWait *sync.WaitGroup) {
	neo4JDatabase := graph.NewNeo4JDatabase()
	initErr := neo4JDatabase.InitDB(NEO4J_URL)
	if initErr != nil {
		log.Fatal(initErr)
	}

	for j := range jobs {
		var insertErr error
		switch insertType {
		case "author":
			author := j.(database.Person)
			person := model.Person{
				Name:  author.Name,
				Email: author.Email,
				Url:   author.Url,
			}
			packageName := author.PackageName

			insertErr = graph.InsertAuthorRelation(neo4JDatabase, person, packageName)
		case "maintainers":
			maintainer := j.(database.Person)
			person := model.Person{
				Name:  maintainer.Name,
				Email: maintainer.Email,
				Url:   maintainer.Url,
			}
			packageName := maintainer.PackageName

			insertErr = graph.InsertMaintainerRelation(neo4JDatabase, person, packageName)
		case "package":
			packageName := j.(string)
			insertErr = graph.InsertPackage(neo4JDatabase, packageName)
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
