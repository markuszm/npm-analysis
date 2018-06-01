package main

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"github.com/markuszm/npm-analysis/database"
	"github.com/markuszm/npm-analysis/database/insert"
	"github.com/markuszm/npm-analysis/evolution"
	"github.com/mongodb/mongo-go-driver/bson"
	"log"
	"sync"
)

const MONGOURL = "mongodb://npm:npm123@localhost:27017"

const workerNumber = 100

const MYSQL_USER = "root"

const MYSQL_PW = "npm-analysis"

var db *sql.DB

func main() {
	mysqlInitializer := &database.Mysql{}
	mysql, databaseInitErr := mysqlInitializer.InitDB(fmt.Sprintf("%s:%s@/npm?charset=utf8mb4&collation=utf8mb4_bin", MYSQL_USER, MYSQL_PW))
	if databaseInitErr != nil {
		log.Fatal(databaseInitErr)
	}

	err := database.CreatePopularity(mysql)

	if err != nil {
		log.Fatal(err)
	}

	db = mysql

	mongoDB := database.NewMongoDB(MONGOURL, "npm", "downloads")

	mongoDB.Connect()
	defer mongoDB.Disconnect()

	workerWait := sync.WaitGroup{}

	jobs := make(chan database.Document, 100)

	for w := 1; w <= workerNumber; w++ {
		workerWait.Add(1)
		go worker(w, jobs, &workerWait)
	}

	cursor, err := mongoDB.ActiveCollection.Find(context.Background(), bson.NewDocument())
	for cursor.Next(context.Background()) {
		val, err := mongoDB.DecodeValue(cursor)
		if err != nil {
			log.Fatalf("ERROR: Decoding value from mongodb")
		}
		jobs <- val
	}

	close(jobs)

	workerWait.Wait()
}

func worker(id int, jobs chan database.Document, workerWait *sync.WaitGroup) {
	for j := range jobs {
		processDocument(j)

	}
	workerWait.Done()
}

func processDocument(doc database.Document) {
	val := doc.Value
	if val == "" {
		log.Printf("WARNING: empty document for %v", doc.Key)
	}
	downloadCount := evolution.DownloadCountResponse{}

	err := json.Unmarshal([]byte(val), &downloadCount)
	if err != nil {
		log.Fatalf("ERROR: Unmarshalling: %v", err)
	}

	popularity := evolution.CalculatePopularity(doc.Key, downloadCount)

	err = insert.StorePopularity(popularity, db)
	if err != nil {
		log.Fatalf("ERROR: inserting popularity of package %v \n with error: %v \n popularity: %v", doc.Key, err, popularity)
	}

}
