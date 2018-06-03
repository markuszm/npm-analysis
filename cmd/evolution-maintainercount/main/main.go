package main

import (
	"database/sql"
	"fmt"
	"github.com/markuszm/npm-analysis/database"
	"log"
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

	log.Print(changes) // remove
}
