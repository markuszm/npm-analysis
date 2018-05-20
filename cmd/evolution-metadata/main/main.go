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

var db *sql.DB

func main() {
	flag.BoolVar(&DEBUG, "debug", false, "DEBUG output")
	flag.Parse()

	mysqlInitializer := &database.Mysql{}
	mysql, databaseInitErr := mysqlInitializer.InitDB(fmt.Sprintf("%s:%s@/npm?charset=utf8mb4&collation=utf8mb4_bin", MYSQL_USER, MYSQL_PW))
	if databaseInitErr != nil {
		log.Fatal(databaseInitErr)
	}
	err := database.CreateLicenseTable(mysql)
	if err != nil {
		log.Fatal(err)
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

	var versions []insert.License

	for version, data := range metadata.Versions {
		license := evolution.ProcessLicense(data)
		if license == "" {
			license = evolution.ProcessLicenses(data)
		}
		timeForVersion := evolution.ParseTime(metadata, data.Version)
		versions = append(versions, insert.License{PkgName: data.Name, License: license, Version: version, Time: timeForVersion})
	}

	err = insert.StoreLicenseWithVersion(db, versions)
	if err != nil {
		log.Fatal(err)
	}

	return len(metadata.Versions)
}

func createTypeMapping(metadata model.Metadata) {
	for k, val := range metadata.Time {
		t := reflect.TypeOf(val)
		if t.String() != "string" {
			log.Print(k, val)
		}
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
