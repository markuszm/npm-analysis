package main

import (
	"log"
	"npm-analysis/database/graph"
)

const NEO4J_URL = "bolt://neo4j:npm@localhost:7687"

func main() {
	db := graph.NewNeo4JDatabase()
	initErr := db.InitDB(NEO4J_URL)
	if initErr != nil {
		log.Fatal(initErr)
	}

	rows, execErr := db.Exec("MATCH (n) DELETE (n)", nil)
	if execErr != nil {
		log.Fatal(execErr)
	}
	log.Println("Rows affected ", rows)

	db.Close()
}
