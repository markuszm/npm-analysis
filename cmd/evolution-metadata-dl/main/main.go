package main

import (
	"flag"
	"fmt"
	"github.com/markuszm/npm-analysis/database"
	"github.com/markuszm/npm-analysis/evolution"
	"github.com/markuszm/npm-analysis/util"
	"log"
	"sync"
)

const MYSQL_USER = "root"

const MYSQL_PW = "npm-analysis"

const MONGOURL = "mongodb://npm:npm123@localhost:27017"

const workerNumber = 25

var command string

func main() {
	flag.StringVar(&command, "command", "check", "either download or check metadata")

	mysqlInitializer := &database.Mysql{}
	mysql, databaseInitErr := mysqlInitializer.InitDB(fmt.Sprintf("%s:%s@/npm?charset=utf8mb4&collation=utf8mb4_bin", MYSQL_USER, MYSQL_PW))
	if databaseInitErr != nil {
		log.Fatal(databaseInitErr)
	}
	defer mysql.Close()

	count := 0

	workerWait := sync.WaitGroup{}

	jobs := make(chan string, 100)

	for w := 1; w <= workerNumber; w++ {
		workerWait.Add(1)
		go worker(w, jobs, &workerWait)
	}

	log.Println("Loading packages from database")

	packages, err := database.GetPackages(mysql)
	if err != nil {
		log.Fatal("cannot load packages from mysql")
	}

	for _, pkg := range packages {
		jobs <- pkg
	}

	close(jobs)

	log.Println(count)

	workerWait.Wait()
}

func worker(workerId int, jobs chan string, workerWait *sync.WaitGroup) {
	var mongoDB *database.MongoDB

	if command == "download" {
		mongoDB := database.NewMongoDB(MONGOURL, "npm", "packages")
		mongoDB.Connect()
		defer mongoDB.Disconnect()
		log.Printf("logged in mongo - workerId %v", workerId)

		ensureIndex(mongoDB)
	}

	for pkg := range jobs {
		if command == "download" {
			err := downloadMetadata(mongoDB, pkg)
			if err != nil {
				log.Printf("ERROR: %v", err)
				jobs <- pkg
				log.Printf("Processed package %v worker: %v", pkg, workerId)
			}
		}

		if command == "check" {
			exists, err := evolution.PackageStillExists(pkg)
			if !exists || err != nil {
				log.Printf("Package %v does not exist anymore", pkg)
			}
		}

	}

	workerWait.Done()
	log.Println("send finished worker ", workerId)
}

func downloadMetadata(db *database.MongoDB, pkg string) error {
	if val, err := db.FindOneSimple("key", pkg); val != "" && err == nil {
		log.Printf("Package %v already exists", pkg)

		val, err := util.Decompress(val)
		if err != nil {
			log.Fatalf("ERROR: Decompressing: %v", err)
		}

		if val == "" {
			err := db.RemoveWithKey(pkg)
			if err != nil {
				log.Fatalf("ERROR: could not remove already existing but wrong data for package %v", pkg)
			}
		} else {
			return nil
		}
	}

	doc, err := evolution.GetMetadataFromNpm(pkg)
	if err != nil {
		return err
	}

	data, err := util.Compress(doc)
	if err != nil {
		log.Fatalf(err.Error())
	}

	db.InsertOneSimple(pkg, data)
	if err != nil {
		log.Fatalf("ERROR: inserting %v into mongo with %s", pkg, err)
	}
	return nil
}

func ensureIndex(mongoDB *database.MongoDB) {
	indexResp, err := mongoDB.EnsureSingleIndex("key")
	if err != nil {
		log.Fatalf("Index cannot be created with ERROR: %v", err)
	}
	log.Printf("Index created %v", indexResp)
}
