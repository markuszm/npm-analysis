package main

import (
	"context"
	"database/sql"
	"encoding/json"
	"flag"
	"fmt"
	"github.com/markuszm/npm-analysis/database"
	"github.com/markuszm/npm-analysis/database/insert"
	"github.com/markuszm/npm-analysis/evolution"
	"github.com/markuszm/npm-analysis/model"
	"github.com/markuszm/npm-analysis/plots"
	"github.com/mongodb/mongo-go-driver/bson"
	"log"
	"sync"
)

const MONGOURL = "mongodb://npm:npm123@localhost:27017"

const workerNumber = 100

const MYSQL_USER = "root"

const MYSQL_PW = "npm-analysis"

var resultPath string

var storeDatabase bool
var isAverage bool

var db *sql.DB

func main() {
	flag.BoolVar(&storeDatabase, "store", false, "whether it should store yearly popularity to mysql")
	// TODO: at the moment result path is not used
	flag.StringVar(&resultPath, "resultPath", "/home/markus/npm-analysis/popularity", "result path for monthly popularity")
	flag.BoolVar(&isAverage, "average", true, "whether to calculate average or just first day of month")
	flag.Parse()

	if storeDatabase {
		mysqlInitializer := &database.Mysql{}
		mysql, err := mysqlInitializer.InitDB(fmt.Sprintf("%s:%s@/npm?charset=utf8mb4&collation=utf8mb4_bin", MYSQL_USER, MYSQL_PW))
		if err != nil {
			log.Fatal(err)
		}
		defer mysql.Close()

		err = database.CreatePopularity(mysql)
		if err != nil {
			log.Fatal(err)
		}

		db = mysql
	}

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
	if err != nil {
		log.Fatal(err)
	}
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
	if doc.Key == "" {
		return
	}

	val := doc.Value
	if val == "" {
		log.Printf("WARNING: empty document for %v", doc.Key)
	}
	downloadCount := evolution.DownloadCountResponse{}

	err := json.Unmarshal([]byte(val), &downloadCount)
	if err != nil {
		log.Fatalf("ERROR: Unmarshalling: %v", err)
	}

	if storeDatabase {
		popularity := evolution.CalculateAveragePopularityByYear(doc.Key, downloadCount)

		err = insert.StorePopularity(popularity, db)
		if err != nil {
			log.Fatalf("ERROR: inserting popularity of package %v \n with error: %v \n popularity: %v", doc.Key, err, popularity)
		}
	}

	var popularityMonthly model.PopularityMonthly
	if isAverage {
		popularityMonthly = evolution.CalculateAveragePopularityByMonth(doc.Key, downloadCount)
	} else {
		popularityMonthly = evolution.CalculatePopularityByMonth(doc.Key, downloadCount)
	}

	bytes, err := json.Marshal(popularityMonthly.Popularity)
	if err != nil {
		log.Fatal(err)
	}

	folderName := "popularity"
	if isAverage {
		folderName = "popularity-average"
	}
	plots.SaveValues(popularityMonthly.PackageName, folderName, bytes)

}