package cmd

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"github.com/markuszm/npm-analysis/database"
	"github.com/markuszm/npm-analysis/database/insert"
	"github.com/markuszm/npm-analysis/evolution"
	"github.com/markuszm/npm-analysis/model"
	"github.com/markuszm/npm-analysis/plots"
	"github.com/mongodb/mongo-go-driver/bson"
	"github.com/spf13/cobra"
	"log"
	"sync"
)

const downloadcountMongourl = "mongodb://npm:npm123@localhost:27017"

const downloadCountWorkerNumber = 100

const downloadCountMysqlUser = "root"

const downloadCountMysqlPassword = "npm-analysis"

var downloadCountResultPath string

var downloadCountStoreDatabase bool
var downloadCountIsAverage bool

var downloadCountDB *sql.DB

var downloadCountCmd = &cobra.Command{
	Use:   "downloadCount",
	Short: "Processes download count data",
	Long:  `...`,
	Run: func(cmd *cobra.Command, args []string) {
		if downloadCountStoreDatabase {
			mysqlInitializer := &database.Mysql{}
			mysql, err := mysqlInitializer.InitDB(fmt.Sprintf("%s:%s@/npm?charset=utf8mb4&collation=utf8mb4_bin", downloadCountMysqlUser, downloadCountMysqlPassword))
			if err != nil {
				log.Fatal(err)
			}
			defer mysql.Close()

			err = database.CreatePopularity(mysql)
			if err != nil {
				log.Fatal(err)
			}

			downloadCountDB = mysql
		}

		mongoDB := database.NewMongoDB(downloadcountMongourl, "npm", "downloads")

		mongoDB.Connect()
		defer mongoDB.Disconnect()

		workerWait := sync.WaitGroup{}

		jobs := make(chan database.Document, 100)

		for w := 1; w <= downloadCountWorkerNumber; w++ {
			workerWait.Add(1)
			go downloadCountWorker(w, jobs, &workerWait)
		}

		cursor, err := mongoDB.ActiveCollection.Find(context.Background(), bson.D{})
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
	},
}

func init() {
	rootCmd.AddCommand(downloadCountCmd)

	downloadCountCmd.Flags().BoolVar(&downloadCountStoreDatabase, "store", false, "whether it should store yearly popularity to mysql")
	downloadCountCmd.Flags().BoolVar(&downloadCountIsAverage, "average", true, "whether to calculate average or just first day of month")

	// TODO: at the moment result path is not used
	downloadCountCmd.Flags().StringVar(&downloadCountResultPath, "resultPath", "/home/markus/npm-analysis/popularity", "result path for monthly popularity")
}

func downloadCountWorker(id int, jobs chan database.Document, workerWait *sync.WaitGroup) {
	for j := range jobs {
		downloadCountProcessDocument(j)

	}
	workerWait.Done()
}

func downloadCountProcessDocument(doc database.Document) {
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

	if downloadCountStoreDatabase {
		popularity := evolution.CalculateAveragePopularityByYear(doc.Key, downloadCount)

		err = insert.StorePopularity(popularity, downloadCountDB)
		if err != nil {
			log.Fatalf("ERROR: inserting popularity of package %v \n with error: %v \n popularity: %v", doc.Key, err, popularity)
		}
	}

	var popularityMonthly model.PopularityMonthly
	if downloadCountIsAverage {
		popularityMonthly = evolution.CalculateAveragePopularityByMonth(doc.Key, downloadCount)
	} else {
		popularityMonthly = evolution.CalculatePopularityByMonth(doc.Key, downloadCount)
	}

	bytes, err := json.Marshal(popularityMonthly.Popularity)
	if err != nil {
		log.Fatal(err)
	}

	folderName := "popularity"
	if downloadCountIsAverage {
		folderName = "popularity-average"
	}
	plots.SaveValues(popularityMonthly.PackageName, folderName, bytes)

}
