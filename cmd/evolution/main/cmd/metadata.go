package cmd

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"github.com/markuszm/npm-analysis/database"
	"github.com/markuszm/npm-analysis/database/insert"
	"github.com/markuszm/npm-analysis/evolution"
	"github.com/markuszm/npm-analysis/model"
	"github.com/markuszm/npm-analysis/util"
	"github.com/spf13/cobra"
	"io/ioutil"
	"log"
	"os"
	"reflect"
	"runtime/debug"
	"sync"
	"time"
)

const metadataMongoUrl = "mongodb://npm:npm123@localhost:27017"

const metadataWorkerNumber = 100

const metadataMysqlUser = "root"

const metadataMysqlPassword = "npm-analysis"

// time cutoff because other dependency data was downloaded at different time
// other data was downloaded at Fr 13 Apr 2018 13∶38
var timeCutoff = time.Unix(1523626680, 0)

var typeMapping = sync.Map{}

var metadataDebug bool

var metadataInsertType string

var metadataDB *sql.DB

var metadataCmd = &cobra.Command{
	Use:   "metadata",
	Short: "Process and insert evolution metadata into database",
	Long:  `...`,
	Run: func(cmd *cobra.Command, args []string) {
		mysqlInitializer := &database.Mysql{}
		mysql, databaseInitErr := mysqlInitializer.InitDB(fmt.Sprintf("%s:%s@/npm?charset=utf8mb4&collation=utf8mb4_bin", metadataMysqlUser, metadataMysqlPassword))
		if databaseInitErr != nil {
			log.Fatal(databaseInitErr)
		}
		defer mysql.Close()

		if metadataInsertType == "" {
			log.Print("WARNING: No insert type selected")
			log.Print("Options are: license, licenseChange, maintainers, dependencies, version")
		}

		var createError error
		switch metadataInsertType {
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
		case "versionNormalized":
			createError = database.CreateVersionChangeNormalizedTable(mysql)
		default:
			log.Print("WARNING: Wrong insert type - no changes")
		}

		if createError != nil {
			log.Fatal(createError)
		}

		metadataDB = mysql

		mongoDB := database.NewMongoDB(metadataMongoUrl, "npm", "packages")

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

		results := make(chan int, metadataWorkerNumber)

		for w := 1; w <= metadataWorkerNumber; w++ {
			workerWait.Add(1)
			go metadataWorker(w, jobs, results, &workerWait)
		}
		startTime = time.Now()

		for _, doc := range allDocs {
			jobs <- doc
		}

		close(jobs)

		workerWait.Wait()
		endTime = time.Now()
		log.Printf("Took %v seconds to parse all documents", endTime.Sub(startTime).Seconds())

		if metadataDebug {
			printTypeMapping()
		}

		for w := 1; w <= metadataWorkerNumber; w++ {
			result := <-results
			sumVersions += result
		}
		log.Printf("%v Versions", sumVersions)
	},
}

func init() {
	rootCmd.AddCommand(metadataCmd)

	metadataCmd.Flags().BoolVar(&metadataDebug, "debug", false, "DEBUG output")
	metadataCmd.Flags().StringVar(&metadataInsertType, "insert", "", "type to insert")
}

func metadataWorker(id int, jobs chan database.Document, resultChan chan int, workerWait *sync.WaitGroup) {
	versions := 0
	for j := range jobs {
		versions += metadataProcessDocument(j)
	}
	resultChan <- versions
	workerWait.Done()
}

func metadataProcessDocument(doc database.Document) int {
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
		ioutil.WriteFile("./output/error.json", []byte(val), os.ModePerm)
		log.Fatalf("ERROR: Unmarshalling: %v", err)
	}

	if metadataDebug {
		createTypeMapping(metadata)
	}

	defer func() {
		if r := recover(); r != nil {
			log.Fatalf("document process error, %v \n with error %v \n Stack: %v", val, r, string(debug.Stack()))
		}
	}()

	var insertError error
	switch metadataInsertType {
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
	case "versionNormalized":
		insertError = insertVersionChangesNormalized(metadata)
	}

	if insertError != nil {
		log.Fatalf("ERROR: inserting into database with %v", insertError)
	}

	return len(metadata.Versions)
}

func insertLicenses(metadata model.Metadata) error {
	var licenses []insert.LicenseVersion
	for version, data := range metadata.Versions {
		releaseTime := evolution.GetTimeForVersion(metadata, data.Version)
		if releaseTime.After(timeCutoff) {
			continue
		}

		license := evolution.ProcessLicense(data)
		if license == "" {
			license = evolution.ProcessLicenses(data)
		}

		licenses = append(licenses, insert.LicenseVersion{PkgName: data.Name, License: license, Version: version, Time: releaseTime})
	}
	err := insert.StoreLicenseWithVersion(metadataDB, licenses)
	return err
}

func insertLicenseChanges(metadata model.Metadata) error {
	licenseChanges, err := evolution.ProcessLicenseChanges(metadata, timeCutoff)
	if err != nil {
		log.Fatalf("ERROR: Processing licences in package: %v with error: %v", metadata.Name, err)
	}
	err = insert.StoreLicenceChanges(metadataDB, licenseChanges)
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
	err = insert.StoreMaintainerChange(metadataDB, maintainerChanges)
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
	err = insert.StoreDependencyChanges(metadataDB, dependencyChanges)
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
	err = insert.StoreVersionChanges(metadataDB, versionChanges)
	if err != nil {
		log.Fatalf("ERROR: inserting version changes of package %v with error: %v", metadata.Name, err)
	}
	return nil
}

func insertVersionChangesNormalized(metadata model.Metadata) error {
	versionChanges, err := evolution.ProcessVersionsNormalized(metadata, timeCutoff)
	if err != nil {
		log.Fatalf("ERROR: Processing versions in package: %v with error: %v", metadata.Name, err)
	}
	err = insert.StoreVersionChangesNormalized(metadataDB, versionChanges)
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
