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
	"github.com/markuszm/npm-analysis/database"
	"github.com/markuszm/npm-analysis/database/insert"
	"github.com/markuszm/npm-analysis/database/model"
	"os"
	"strings"
	"sync"
)

const PATH_TO_NPM_JSON = "/home/markus/npm-analysis/npm-all.json"

const DATABASE_PATH = "/home/markus/npm-analysis/npm.db"

const ERROR_PATH = "/home/markus/npm-analysis/errors.txt"

const workerNumber = 100

const MYSQL_USER = "root"

const MYSQL_PW = "npm-analysis"

var db *sql.DB

var errorStr strings.Builder

var insertType string

var typeMapping = sync.Map{}

func main() {
	dbFlag := flag.String("db", "mysql", "name of db to use")
	createFlag := flag.Bool("create", false, "create db scheme")
	flag.StringVar(&insertType, "insert", "package", "what value to insert")

	flag.Parse()

	if *dbFlag == "mysql" {
		db = initializeDB(
			&database.Mysql{},
			fmt.Sprintf("%s:%s@/npm?charset=utf8mb4&collation=utf8mb4_bin", MYSQL_USER, MYSQL_PW))
	}

	if *dbFlag == "sqlite" {
		db = initializeDB(&database.Sqlite{}, DATABASE_PATH)
	}

	if *createFlag {
		dbErr := createSchema(db)
		if dbErr != nil {
			log.Fatal(dbErr)
		}
	}

	data, readErr := ioutil.ReadFile(PATH_TO_NPM_JSON)

	if readErr != nil {
		log.Fatal(errors.Wrap(readErr, "Read error"))
	}

	errorStr = strings.Builder{}

	workerWait := sync.WaitGroup{}

	jobs := make(chan []byte, 100)

	for w := 1; w <= workerNumber; w++ {
		workerWait.Add(1)
		go worker(w, jobs, &errorStr, &workerWait)
	}

	count := 0

	jsonparser.ArrayEach(data, func(value []byte, dataType jsonparser.ValueType, offset int, err error) {
		count += 1
		jobs <- value
	}, "rows")

	close(jobs)

	workerWait.Wait()

	log.Println(count)

	//typeMapping.Range(func(key, value interface{}) bool {
	//	log.Println(key, "count: ", value)
	//	return true
	//})

	errFile, _ := os.Create(ERROR_PATH)
	defer errFile.Close()
	io.Copy(errFile, strings.NewReader(errorStr.String()))
}

func storePackageValues(value []byte, db *sql.DB) (string, error) {
	pkgVal, _, _, _ := jsonparser.Get(value, "value")
	var pkg model.Package
	jsonErr := json.Unmarshal(pkgVal, &pkg)

	//t := reflect.TypeOf(pkg.Maintainers)
	//if val, ok := typeMapping.Load(t); !ok {
	//	typeMapping.Store(t, 1)
	//} else {
	//	typeMapping.Store(t, val.(int)+1)
	//}

	var storeErr error

	switch insertType {
	case "package":
		storeErr = insert.StorePackage(db, pkg)
	case "dependencies":
		storeErr = insert.StoreDependencies(db, pkg)
	case "authors":
		storeErr = insert.StoreAuthor(db, pkg)
	case "maintainers":
		storeErr = insert.StoreMaintainers(db, pkg)
	}

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
	return db
}

func createSchema(db *sql.DB) error {
	return database.CreateTables(db)
}

func worker(id int, jobs chan []byte, errorsStr *strings.Builder, workerWait *sync.WaitGroup) {
	for j := range jobs {
		name, err := storePackageValues(j, db)
		log.Println("worker", id, "finished job", name)
		if err != nil {
			errorsStr.WriteString(err.Error())
		}
	}
	workerWait.Done()
}
