package main

import (
	"database/sql"
	"encoding/json"
	"flag"
	"fmt"
	"github.com/markuszm/npm-analysis/database"
	"github.com/markuszm/npm-analysis/database/insert"
	"github.com/markuszm/npm-analysis/evolution"
	"github.com/markuszm/npm-analysis/model"
	"github.com/markuszm/npm-analysis/util"
	"io/ioutil"
	"log"
	"os"
	"reflect"
	"runtime/debug"
	"sync"
	"time"
)

const MONGOURL = "mongodb://npm:npm123@localhost:27017"

const workerNumber = 100

const MYSQL_USER = "root"

const MYSQL_PW = "npm-analysis"

// time cutoff because other dependency data was downloaded at different time
// other data was downloaded at Fr 13 Apr 2018 13∶38
var timeCutoff = time.Unix(1523626680, 0)

var typeMapping = sync.Map{}

var DEBUG bool

var insertType string

var db *sql.DB

func main() {
	flag.BoolVar(&DEBUG, "debug", false, "DEBUG output")
	flag.StringVar(&insertType, "insert", "", "type to insert")
	flag.Parse()

	mysqlInitializer := &database.Mysql{}
	mysql, databaseInitErr := mysqlInitializer.InitDB(fmt.Sprintf("%s:%s@/npm?charset=utf8mb4&collation=utf8mb4_bin", MYSQL_USER, MYSQL_PW))
	if databaseInitErr != nil {
		log.Fatal(databaseInitErr)
	}

	if insertType == "" {
		log.Print("WARNING: No insert type selected")
		log.Print("Options are: license, licenseChange, maintainers, dependencies, version")
	}

	var createError error
	switch insertType {
	case "license":
		createError = database.CreateLicenseTable(mysql)
	case "licenseChange":
		createError = database.CreateLicenseChangeTable(mysql)
	case "maintainers":
		createError = database.CreateMaintainerChangeTable(mysql)
	case "dependencies":
		createError = database.CreateDependencyChangeTable(mysql)
	case "version":
		createError = database.CreateVersionChangeTable(mysql)
	default:
		log.Print("WARNING: Wrong insert type - no changes")
	}

	if createError != nil {
		log.Fatal(createError)
	}

	db = mysql

	mongoDB := database.NewMongoDB(MONGOURL, "npm", "packages")

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

	jobs := make(chan database.Document, 100)

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

func worker(id int, jobs chan database.Document, resultChan chan int, workerWait *sync.WaitGroup) {
	versions := 0
	for j := range jobs {
		versions += processDocument(j)
	}
	resultChan <- versions
	workerWait.Done()
}

func processDocument(doc database.Document) int {
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
			log.Fatalf("document process error, %v \n with error %v \n Stack: %v", val, r, string(debug.Stack()))
		}
	}()

	var insertError error
	switch insertType {
	case "license":
		insertError = insertLicenses(metadata)
	case "licenseChange":
		insertError = insertLicenseChanges(metadata)
	case "maintainers":
		insertError = insertMaintainersChanges(metadata)
	case "dependencies":
		insertError = insertDependencyChanges(metadata)
	case "version":
		insertError = insertVersionChanges(metadata)
	}

	if insertError != nil {
		log.Fatalf("ERROR: inserting into database with %v", insertError)
	}

	return len(metadata.Versions)
}

func insertLicenses(metadata model.Metadata) error {
	var licenses []insert.License
	for version, data := range metadata.Versions {
		releaseTime := evolution.GetTimeForVersion(metadata, data.Version)
		if releaseTime.After(timeCutoff) {
			continue
		}

		license := evolution.ProcessLicense(data)
		if license == "" {
			license = evolution.ProcessLicenses(data)
		}

		licenses = append(licenses, insert.License{PkgName: data.Name, License: license, Version: version, Time: releaseTime})
	}
	err := insert.StoreLicenseWithVersion(db, licenses)
	return err
}

func insertLicenseChanges(metadata model.Metadata) error {
	licenseChanges, err := evolution.ProcessLicenseChanges(metadata, timeCutoff)
	if err != nil {
		log.Fatalf("ERROR: Processing licences in package: %v with error: %v", metadata.Name, err)
	}
	err = insert.StoreLicenceChanges(db, licenseChanges)
	if err != nil {
		log.Fatalf("ERROR: inserting licence changes of package %v with error: %v", metadata.Name, err)
	}
	return nil
}

func insertMaintainersChanges(metadata model.Metadata) error {
	maintainerChanges, err := evolution.ProcessMaintainersTimeSorted(metadata, timeCutoff)
	if err != nil {
		log.Fatalf("ERROR: Processing maintainers in package: %v with error: %v", metadata.Name, err)
	}
	err = insert.StoreMaintainerChange(db, maintainerChanges)
	if err != nil {
		log.Fatalf("ERROR: inserting maintainer changes of package %v with error: %v", metadata.Name, err)
	}
	return nil
}

func insertDependencyChanges(metadata model.Metadata) error {
	dependencyChanges, err := evolution.ProcessDependencies(metadata, timeCutoff)
	if err != nil {
		log.Fatalf("ERROR: Processing dependencies in package: %v with error: %v", metadata.Name, err)
	}
	err = insert.StoreDependencyChanges(db, dependencyChanges)
	if err != nil {
		log.Fatalf("ERROR: inserting dependency changes of package %v with error: %v", metadata.Name, err)
	}
	return nil
}

func insertVersionChanges(metadata model.Metadata) error {
	versionChanges, err := evolution.ProcessVersions(metadata, timeCutoff)
	if err != nil {
		log.Fatalf("ERROR: Processing versions in package: %v with error: %v", metadata.Name, err)
	}
	err = insert.StoreVersionChanges(db, versionChanges)
	if err != nil {
		log.Fatalf("ERROR: inserting version changes of package %v with error: %v", metadata.Name, err)
	}
	return nil
}

func createTypeMapping(metadata model.Metadata) {
	for _, val := range metadata.Versions {
		for _, v := range val.Dependencies {
			t := reflect.TypeOf(v)
			if val, ok := typeMapping.Load(t); !ok {
				typeMapping.Store(t, 1)
			} else {
				typeMapping.Store(t, val.(int)+1)
			}
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
