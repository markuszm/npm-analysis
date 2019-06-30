package cmd

import (
	"context"
	"encoding/json"
	"github.com/markuszm/npm-analysis/database"
	"github.com/markuszm/npm-analysis/evolution"
	"github.com/markuszm/npm-analysis/model"
	"github.com/markuszm/npm-analysis/util"
	"github.com/mongodb/mongo-go-driver/bson"
	"github.com/spf13/cobra"
	"io/ioutil"
	"log"
	"os"
	"sync"
	"time"
)

var maintainerTimelineMongoUrl string

var maintainerTimelineWorkers int

// Extracts maintainers and dependencies for every month and grouped by package from evolution data and stores it into mongo collection called "timeline"
var maintainerTimelineCmd = &cobra.Command{
	Use:   "maintainerTimeline",
	Short: "Create maintainer timeline aggregation",
	Long:  `Extracts maintainers and dependencies for every month and grouped by package from evolution data and stores it into mongo collection called "timeline`,
	Run: func(cmd *cobra.Command, args []string) {
		mongoDB := database.NewMongoDB(maintainerTimelineMongoUrl, "npm", "packages")

		mongoDB.Connect()
		defer mongoDB.Disconnect()

		startTime := time.Now()

		workerWait := sync.WaitGroup{}

		jobs := make(chan database.Document, 100)

		for w := 1; w <= maintainerTimelineWorkers; w++ {
			workerWait.Add(1)
			go maintainerTimelineWorker(w, jobs, "timelineNew", &workerWait)
		}

		cursor, err := mongoDB.ActiveCollection.Find(context.Background(), bson.D{})
		if err != nil {
			log.Fatal(err)
		}
		for cursor.Next(context.Background()) {
			doc, err := mongoDB.DecodeValue(cursor)
			if err != nil {
				log.Fatal(err)
			}
			jobs <- doc
		}

		close(jobs)

		workerWait.Wait()

		endTime := time.Now()

		log.Printf("Took %v minutes to process all Documents from MongoDB", endTime.Sub(startTime).Minutes())
	},
}

func init() {
	rootCmd.AddCommand(maintainerTimelineCmd)

	maintainerTimelineCmd.Flags().StringVar(&maintainerTimelineMongoUrl, "mongoUrl", "mongodb://npm:npm123@localhost:27017", "url to mongo db")
	maintainerTimelineCmd.Flags().IntVar(&maintainerTimelineWorkers, "workers", 75, "number of workers")
}

func maintainerTimelineWorker(id int, jobs chan database.Document, collectionName string, workerWait *sync.WaitGroup) {
	mongoDB := database.NewMongoDB(maintainerTimelineMongoUrl, "npm", collectionName)
	mongoDB.Connect()
	defer mongoDB.Disconnect()
	log.Printf("logged in mongo - workerId %v", id)

	ensureIndex(mongoDB)
	for j := range jobs {
		maintainerTimelineProcessDocument(j, mongoDB)
	}
	workerWait.Done()
}

func maintainerTimelineProcessDocument(doc database.Document, mongoDB *database.MongoDB) {
	if val, err := mongoDB.FindOneSimple("key", doc.Key); val != "" && err == nil {
		log.Printf("Package %v already exists", doc.Key)

		val, err := util.Decompress(val)
		if err != nil {
			log.Fatalf("ERROR: Decompressing: %v", err)
		}

		if val == "" {
			err := mongoDB.RemoveWithKey(doc.Key)
			if err != nil {
				log.Fatalf("ERROR: could not remove already existing but wrong data for package %v", doc.Key)
			}
		} else {
			return
		}
	}

	val, err := util.Decompress(doc.Value)
	if err != nil {
		log.Fatalf("ERROR: Decompressing: %v", err)
	}

	if val == "" {
		log.Printf("WARNING: empty metadata in package %v", doc.Key)
		return
	}

	metadata := model.Metadata{}

	err = json.Unmarshal([]byte(val), &metadata)
	if err != nil {
		ioutil.WriteFile("./output/error.json", []byte(val), os.ModePerm)
		log.Fatalf("ERROR: Unmarshalling: %v", err)
	}

	packageData := evolution.GetPackageMetadataForEachMonth(metadata)

	err = mongoDB.InsertPackageTimeline(doc.Key, packageData)
	if err != nil {
		log.Fatalf("ERROR: could not insert package timeline with error: %v", err)
	}
}
