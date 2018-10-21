package main

import (
	"encoding/csv"
	"github.com/markuszm/npm-analysis/database/graph"
	"log"
	"os"
)

const NEO4J_URL = "bolt://neo4j:npm@localhost:7687"

const outputFile = "/home/markus/npm-analysis/packages_with_dependents"

func main() {
	packageChan := make(chan []string, 0)

	go queryPackagesWithDependents(packageChan)

	file, err := os.Create(outputFile)
	if err != nil {
		log.Fatal(err)
	}

	writer := csv.NewWriter(file)

	for p := range packageChan {
		log.Print(p)
		err := writer.Write(p)
		if err != nil {
			log.Fatal(err)
		}
	}

	writer.Flush()
}

func queryPackagesWithDependents(packageChan chan []string) {
	database := graph.NewNeo4JDatabase()
	defer database.Close()
	err := database.InitDB(NEO4J_URL)
	if err != nil {
		log.Fatal(err)
	}
	resultChan := make(chan []interface{}, 0)
	go database.QueryStream("MATCH (p:Package)<-[:DEPEND|:DEPEND_DEV|:DEPEND_PEER|:DEPEND_OPTIONAL]-(d:Package) RETURN DISTINCT p.name, p.version", map[string]interface{}{}, resultChan)
	for r := range resultChan {
		// TODO: very unsafe - should check if result is valid
		if r[1] == nil {
			continue
		}
		packageChan <- []string{r[0].(string), r[1].(string)}
	}
	close(packageChan)
}
