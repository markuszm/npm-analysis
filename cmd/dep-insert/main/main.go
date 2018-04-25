package main

import (
	"fmt"
	"github.com/pkg/errors"
	"log"
	"npm-analysis/database"
	"npm-analysis/database/graph"
)

const NEO4J_URL = "bolt://neo4j:npm@localhost:7687"

const MYSQL_USER = "root"

const MYSQL_PW = "npm-analysis"

func main() {
	mysqlInitializer := &database.Mysql{}
	mysql, databaseInitErr := mysqlInitializer.InitDB(fmt.Sprintf("%s:%s@/npm?charset=utf8mb4&collation=utf8mb4_bin", MYSQL_USER, MYSQL_PW))
	if databaseInitErr != nil {
		log.Fatal(databaseInitErr)
	}

	neo4JDatabase := graph.NewNeo4JDatabase()
	initErr := neo4JDatabase.InitDB(NEO4J_URL)
	if initErr != nil {
		log.Fatal(initErr)
	}

	count := 0

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

		rowsAffected, insertErr := neo4JDatabase.Exec(`
			MERGE (p1:Package {name: {p1}}) 
			MERGE (p2:Package {name: {p2}})
			MERGE (p1)-[:DEPEND {version: {version}}]->(p2)`,
			map[string]interface{}{"p1": pkgName, "p2": name, "version": version})

		if insertErr != nil {
			log.Fatal(errors.Wrap(insertErr, "Failed to insert into neo4j"))
		}
		log.Println("Inserted:", id, name, version, pkgName, "with affected rows", rowsAffected)
	}
	err = rows.Err()
	if err != nil {
		log.Fatal(err)
	}

	log.Println(count)

	neo4JDatabase.Close()
}
