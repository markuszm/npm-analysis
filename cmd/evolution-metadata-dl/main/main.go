package main

import (
	"bytes"
	"compress/gzip"
	"context"
	"fmt"
	"github.com/mongodb/mongo-go-driver/bson"
	"github.com/mongodb/mongo-go-driver/mongo"
	"io/ioutil"
	"log"
	"net/http"
	"npm-analysis/database"
	"strings"
	"sync"
)

const npmUrl = "https://registry.npmjs.com/"

const MYSQL_USER = "root"

const MYSQL_PW = "npm-analysis"

const workerNumber = 25

func main() {
	mysqlInitializer := &database.Mysql{}
	mysql, databaseInitErr := mysqlInitializer.InitDB(fmt.Sprintf("%s:%s@/npm?charset=utf8mb4&collation=utf8mb4_bin", MYSQL_USER, MYSQL_PW))
	if databaseInitErr != nil {
		log.Fatal(databaseInitErr)
	}

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
	mongodb, err := mongo.NewClient("mongodb://npm:npm123@localhost:27017")
	mongodb.Connect(context.Background())
	if err != nil {
		log.Fatal(err)
	}
	defer mongodb.Disconnect(context.Background())

	log.Printf("logged in mongo workerId %v", workerId)

	metadataDB := mongodb.Database("npm").Collection("packages")

	for pkg := range jobs {
		pkgName := pkg
		if strings.Contains(pkg, "/") {
			pkgName = strings.Replace(pkg, "/", "%2f", -1)
		}

		result := metadataDB.FindOne(context.Background(), bson.NewDocument(
			bson.EC.String("key", pkg),
		))

		err = result.Decode(nil)
		if err == nil {
			log.Printf("Package %v already exists", pkg)
			continue
		}

		url := npmUrl + pkgName

		doc, err := getMetadataFromNpm(url)
		if err != nil {
			log.Printf("ERROR: %v", err)
			jobs <- pkg
		}

		var b bytes.Buffer
		gz := gzip.NewWriter(&b)
		if _, err := gz.Write([]byte(doc)); err != nil {
			log.Fatal(err)
		}
		if err := gz.Flush(); err != nil {
			log.Fatal(err)
		}
		if err := gz.Close(); err != nil {
			log.Fatal(err)
		}

		data := b.String()

		_, err = metadataDB.InsertOne(context.Background(), bson.NewDocument(
			bson.EC.String("key", pkg),
			bson.EC.String("value", data),
		))
		if err != nil {
			log.Fatalf("ERROR: inserting %v into mongo with %s", pkgName, err)
		}

		log.Printf("Inserted package metadata of %v worker: %v", pkg, workerId)

	}

	workerWait.Done()
	log.Println("send finished worker ", workerId)
}

func getMetadataFromNpm(url string) (string, error) {
	resp, err := http.Get(url)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()
	bytes, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}
	doc := string(bytes)
	return doc, err
}
