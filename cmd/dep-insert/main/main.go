package main

import (
	"fmt"
	"github.com/pkg/errors"
	"log"
	"npm-analysis/database"
	"npm-analysis/database/graph"
	"sync"
)

const NEO4J_URL = "bolt://neo4j:npm@localhost:7687"

const MYSQL_USER = "root"

const MYSQL_PW = "npm-analysis"

type Dependency struct {
	id                     int
	name, version, pkgName string
}

const workerNumber = 100

func main() {
	mysqlInitializer := &database.Mysql{}
	mysql, databaseInitErr := mysqlInitializer.InitDB(fmt.Sprintf("%s:%s@/npm?charset=utf8mb4&collation=utf8mb4_bin", MYSQL_USER, MYSQL_PW))
	if databaseInitErr != nil {
		log.Fatal(databaseInitErr)
	}

	count := 0

	workerWait := sync.WaitGroup{}

	jobs := make(chan Dependency, 100)

	for w := 1; w <= workerNumber; w++ {
		workerWait.Add(1)
		go worker(w, jobs, &workerWait)
	}

	rows, err := mysql.Query("select * from dependencies")
	if err != nil {
		log.Fatal(errors.Wrap(err, "Failed to query dependencies"))
	}
	defer rows.Close()
	for rows.Next() {
		var (
			id                     int
			name, version, pkgName string
		)
		err := rows.Scan(&id, &name, &version, &pkgName)
		count++
		if err != nil {
			log.Fatal(errors.Wrap(err, "Could not get info from row"))
		}

		// write dependency to neo4j
		dep := Dependency{id: id, pkgName: pkgName, name: name, version: version}
		jobs <- dep
	}
	err = rows.Err()
	if err != nil {
		log.Fatal(err)
	}

	close(jobs)

	log.Println(count)

	workerWait.Wait()

}

func worker(workerId int, jobs chan Dependency, workerWait *sync.WaitGroup) {
	neo4JDatabase := graph.NewNeo4JDatabase()
	initErr := neo4JDatabase.InitDB(NEO4J_URL)
	if initErr != nil {
		log.Fatal(initErr)
	}

	for j := range jobs {
		insertErr := insertDependency(neo4JDatabase, j)
		if insertErr != nil {
			log.Println(insertErr)
			jobs <- j
		}
		log.Println("worker", workerId, "finished job", j)
	}
	workerWait.Done()
	log.Println("send finished worker ", workerId)

	neo4JDatabase.Close()
}

func insertDependency(neo4JDatabase graph.Database, dep Dependency) error {
	_, insertErr := neo4JDatabase.Exec(`
					MERGE (p1:Package {name: {p1}}) 
					MERGE (p2:Package {name: {p2}})
					MERGE (p1)-[:DEPEND {version: {version}}]->(p2)`,
		map[string]interface{}{"p1": dep.pkgName, "p2": dep.name, "version": dep.version})
	if insertErr != nil {
		return insertErr
	}
	return nil
}
