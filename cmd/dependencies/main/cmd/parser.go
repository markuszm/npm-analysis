package cmd

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"github.com/buger/jsonparser"
	"github.com/markuszm/npm-analysis/database"
	"github.com/markuszm/npm-analysis/database/insert"
	"github.com/markuszm/npm-analysis/evolution"
	"github.com/markuszm/npm-analysis/model"
	"github.com/pkg/errors"
	"github.com/spf13/cobra"
	"io"
	"io/ioutil"
	"log"
	"os"
	"reflect"
	"strings"
	"sync"
)

var parserPathToNpmJson = "./output/npm-all.json"

var parserErrorPath string

var parserWorkerNumber int

var parserMysqlUser string

var parserMysqlPassword string

var db *sql.DB

var errorStr strings.Builder

var parserInsertType string

var typeMapping = sync.Map{}

var parserIsDebug bool
var parserIsCreate bool

func init() {
	rootCmd.AddCommand(parserCmd)

	parserCmd.Flags().BoolVar(&parserIsCreate, "create", false, "create db scheme")
	parserCmd.Flags().BoolVar(&parserIsDebug, "debug", false, "type mapping debug")
	parserCmd.Flags().StringVar(&parserInsertType, "insert", "package", "what value to insert")
	parserCmd.Flags().StringVar(&parserErrorPath, "error", "./output/errors.txt", "path to error file")
	parserCmd.Flags().StringVar(&parserMysqlUser, "mysqlUser", "root", "mysql user")
	parserCmd.Flags().StringVar(&parserMysqlPassword, "mysqlPassword", "npm-analysis", "mysql password")
	parserCmd.Flags().IntVar(&parserWorkerNumber, "workers", 100, "number of workers")
}

var parserCmd = &cobra.Command{
	Use:   "parser",
	Short: "Parses npm metadata and stores it in a database",
	Long:  `...`,
	Run: func(cmd *cobra.Command, args []string) {
		db = initializeDB(&database.Mysql{}, fmt.Sprintf("%s:%s@/npm?charset=utf8mb4&collation=utf8mb4_bin", parserMysqlUser, parserMysqlPassword))

		defer db.Close()

		if parserIsCreate {
			dbErr := createSchema(db)
			if dbErr != nil {
				log.Fatal(dbErr)
			}
		}

		data, readErr := ioutil.ReadFile(parserPathToNpmJson)

		if readErr != nil {
			log.Fatal(errors.Wrap(readErr, "Read error"))
		}

		errorStr = strings.Builder{}

		workerWait := sync.WaitGroup{}

		jobs := make(chan []byte, 100)

		for w := 1; w <= parserWorkerNumber; w++ {
			workerWait.Add(1)
			go parserWorker(w, jobs, &errorStr, &workerWait)
		}

		count := 0

		jsonparser.ArrayEach(data, func(value []byte, dataType jsonparser.ValueType, offset int, err error) {
			count += 1
			jobs <- value
		}, "rows")

		close(jobs)

		workerWait.Wait()

		log.Println(count)

		if parserIsDebug {
			typeMapping.Range(func(key, value interface{}) bool {
				log.Println(key, "count: ", value)
				return true
			})
		}

		errFile, _ := os.Create(parserErrorPath)
		defer errFile.Close()
		io.Copy(errFile, strings.NewReader(errorStr.String()))
	},
}

func storePackageValues(value []byte, db *sql.DB) (string, error) {
	pkgVal, _, _, _ := jsonparser.Get(value, "value")
	var pkg model.Package
	jsonErr := json.Unmarshal(pkgVal, &pkg)
	if jsonErr != nil {
		log.Print(jsonErr)
	}

	if parserIsDebug {
		t := reflect.TypeOf(pkg.Licenses)
		if val, ok := typeMapping.Load(t); !ok {
			typeMapping.Store(t, 1)
		} else {
			typeMapping.Store(t, val.(int)+1)
		}
	}

	var storeErr error

	switch parserInsertType {
	case "package":
		storeErr = insert.StorePackage(db, pkg)
	case "dependencies":
		storeErr = insert.StoreDependencies(db, pkg)
	case "authors":
		storeErr = insert.StoreAuthor(db, pkg)
	case "maintainers":
		storeErr = insert.StoreMaintainers(db, pkg)
	case "license":
		if pkg.License != nil {
			license := evolution.ProcessLicenseInternal(pkg.License)
			if license != "" {
				insert.StoreLicense(db, insert.License{PkgName: pkg.Name, License: license})
			}
		}
		if pkg.Licenses != nil {
			license := evolution.ProcessLicensesInternal(pkg.Licenses)
			if license != "" {
				insert.StoreLicense(db, insert.License{PkgName: pkg.Name, License: license})
			}
		}
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

func parserWorker(id int, jobs chan []byte, errorsStr *strings.Builder, workerWait *sync.WaitGroup) {
	for j := range jobs {
		name, err := storePackageValues(j, db)
		log.Println("worker", id, "finished job", name)
		if err != nil {
			errorsStr.WriteString(err.Error())
		}
	}
	workerWait.Done()
}
