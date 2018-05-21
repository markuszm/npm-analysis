package main

import (
	"database/sql"
	"encoding/json"
	"flag"
	"fmt"
	"github.com/markuszm/npm-analysis/database"
	"github.com/markuszm/npm-analysis/database/evolution"
	"github.com/markuszm/npm-analysis/database/insert"
	"github.com/markuszm/npm-analysis/database/model"
	"github.com/markuszm/npm-analysis/util"
	"io/ioutil"
	"log"
	"os"
	"reflect"
	"sync"
	"time"
)

const MONGOURL = "mongodb://npm:npm123@localhost:27017"

const workerNumber = 100

const MYSQL_USER = "root"

const MYSQL_PW = "npm-analysis"

var typeMapping = sync.Map{}

var DEBUG bool

var insertType string

var db *sql.DB

func main() {
	flag.BoolVar(&DEBUG, "debug", false, "DEBUG output")
	flag.StringVar(&insertType, "insert", "license", "type to insert")
	flag.Parse()

	mysqlInitializer := &database.Mysql{}
	mysql, databaseInitErr := mysqlInitializer.InitDB(fmt.Sprintf("%s:%s@/npm?charset=utf8mb4&collation=utf8mb4_bin", MYSQL_USER, MYSQL_PW))
	if databaseInitErr != nil {
		log.Fatal(databaseInitErr)
	}

	var createError error
	switch insertType {
	case "license":
		createError = database.CreateLicenseTable(mysql)
	case "maintainers":
		createError = database.CreateMaintainersTable(mysql)
	}
	if createError != nil {
		log.Fatal(createError)
	}

	db = mysql

	mongoDB := evolution.NewMongoDB(MONGOURL, "npm", "packages")

	mongoDB.Connect()
	defer mongoDB.Disconnect()

	startTime := time.Now()

	allDocs, err := mongoDB.FindAll()
	if err != nil {
		log.Fatalf("ERROR: %v", err)
	}

	endTime := time.Now()

	log.Printf("Took %v seconds to get all Documents from MongoDB", endTime.Sub(startTime).Seconds())

	sumVersions := 0

	workerWait := sync.WaitGroup{}

	jobs := make(chan evolution.Document, 100)

	results := make(chan int, workerNumber)

	for w := 1; w <= workerNumber; w++ {
		workerWait.Add(1)
		go worker(w, jobs, results, &workerWait)
	}
	startTime = time.Now()

	for _, doc := range allDocs {
		jobs <- doc
	}

	close(jobs)

	workerWait.Wait()
	endTime = time.Now()
	log.Printf("Took %v seconds to parse all documents", endTime.Sub(startTime).Seconds())

	if DEBUG {
		printTypeMapping()
	}

	for w := 1; w <= workerNumber; w++ {
		result := <-results
		sumVersions += result
	}
	log.Printf("%v Versions", sumVersions)
}

func worker(id int, jobs chan evolution.Document, resultChan chan int, workerWait *sync.WaitGroup) {
	versions := 0
	for j := range jobs {
		versions += processDocument(j)
	}
	resultChan <- versions
	workerWait.Done()
}

func processDocument(doc evolution.Document) int {
	val, err := util.Decompress(doc.Value)
	if err != nil {
		log.Fatalf("ERROR: Decompressing: %v", err)
	}

	if val == "" {
		log.Printf("WARNING: empty metadata in package %v", doc.Key)
		return 0
	}

	metadata := model.Metadata{}

	err = json.Unmarshal([]byte(val), &metadata)
	if err != nil {
		ioutil.WriteFile("/home/markus/npm-analysis/error.json", []byte(val), os.ModePerm)
		log.Fatalf("ERROR: Unmarshalling: %v", err)
	}

	if DEBUG {
		createTypeMapping(metadata)
	}

	defer func() {
		if r := recover(); r != nil {
			log.Fatalf("document process error, %v with error %v", metadata, r)
		}
	}()

	var insertError error
	switch insertType {
	case "license":
		insertError = insertLicenses(metadata)
	case "maintainers":
		insertError = insertMaintainers(metadata)
	}

	if insertError != nil {
		log.Fatalf("ERROR: inserting into database with %v", insertError)
	}

	return len(metadata.Versions)
}

func insertLicenses(metadata model.Metadata) error {
	var licenses []insert.License
	for version, data := range metadata.Versions {
		license := evolution.ProcessLicense(data)
		if license == "" {
			license = evolution.ProcessLicenses(data)
		}
		timeForVersion := evolution.ParseTime(metadata, data.Version)
		licenses = append(licenses, insert.License{PkgName: data.Name, License: license, Version: version, Time: timeForVersion})
	}
	err := insert.StoreLicenseWithVersion(db, licenses)
	return err
}

func insertMaintainers(metadata model.Metadata) error {
	_, err := evolution.ProcessMaintainers(metadata)
	if err != nil {
		log.Fatalf("ERROR: Processing maintainers package: %v with error: %v", metadata.Name, err)
	}
	// TODO: insert change list into database
	return nil
}

func createTypeMapping(metadata model.Metadata) {
	for _, val := range metadata.Versions {
		t := reflect.TypeOf(val.Maintainers)
		if val, ok := typeMapping.Load(t); !ok {
			typeMapping.Store(t, 1)
		} else {
			typeMapping.Store(t, val.(int)+1)
		}
	}
}

func printTypeMapping() {
	log.Print("Type Mapping:")
	typeMapping.Range(func(key, value interface{}) bool {
		log.Println(key, "count: ", value)
		return true
	})
}
