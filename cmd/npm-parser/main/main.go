package main

import (
	"database/sql"
	"encoding/json"
	"flag"
	"fmt"
	"github.com/buger/jsonparser"
	"github.com/pkg/errors"
	"io"
	"io/ioutil"
	"log"
	"npm-analysis/database"
	"npm-analysis/database/model"
	"os"
	"strings"
)

const PATH_TO_NPM_JSON = "/home/markus/npm-analysis/npm-all.json"

const DATABASE_PATH = "/home/markus/npm-analysis/npm.db"

const ERROR_PATH = "/home/markus/npm-analysis/errors.txt"

const workerNumber = 100

const MYSQL_USER = "root"

const MYSQL_PW = "npm-analysis"

var db *sql.DB

var errorStr strings.Builder

func main() {
	dbFlag := flag.String("db", "", "name of db to use")

	flag.Parse()

	if *dbFlag == "mysql" {
		db = initializeDB(
			&database.Mysql{},
			fmt.Sprintf("%s:%s@/npm?charset=utf8mb4&collation=utf8mb4_unicode_ci", MYSQL_USER, MYSQL_PW))
	}

	if *dbFlag == "sqlite" {
		db = initializeDB(&database.Sqlite{}, DATABASE_PATH)
	}

	data, readErr := ioutil.ReadFile(PATH_TO_NPM_JSON)

	if readErr != nil {
		log.Fatal(errors.Wrap(readErr, "Read error"))
	}

	errorStr = strings.Builder{}

	finishedWorker := make(chan bool)

	jobs := make(chan []byte, 10000)

	for w := 1; w <= workerNumber; w++ {
		go worker(w, jobs, finishedWorker)
	}

	jsonparser.ArrayEach(data, func(value []byte, dataType jsonparser.ValueType, offset int, err error) {
		jobs <- value
	}, "rows")

	close(jobs)

	for r := 1; r <= workerNumber; r++ {
		<-finishedWorker
	}

	errFile, _ := os.Create(ERROR_PATH)
	defer errFile.Close()
	io.Copy(errFile, strings.NewReader(errorStr.String()))
}

func storePackageValue(value []byte, db *sql.DB) (string, error) {
	pkgVal, _, _, _ := jsonparser.Get(value, "value")
	var pkg model.Package
	jsonErr := json.Unmarshal(pkgVal, &pkg)

	storeErr := database.StorePackage(db, pkg)
	if storeErr != nil {
		log.Fatal(pkg.Name, " ", storeErr, string(value))
	}

	return pkg.Name, jsonErr
}

func initializeDB(databaseInitializer database.Database, settings string) *sql.DB {
	db, databaseInitErr := databaseInitializer.InitDB(settings)
	if databaseInitErr != nil {
		log.Fatal(databaseInitErr)
	}
	tableCreationErr := database.CreateTables(db)
	if tableCreationErr != nil {
		log.Fatal(tableCreationErr)
	}

	return db
}

func worker(id int, jobs chan []byte, finished chan bool) {
	for j := range jobs {
		name, _ := storePackageValue(j, db)
		log.Println("worker", id, "finished job", name)
	}
	finished <- true
}
