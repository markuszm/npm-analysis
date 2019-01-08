package cmd

import (
	"encoding/json"
	"fmt"
	"github.com/markuszm/npm-analysis/database"
	"github.com/markuszm/npm-analysis/evolution"
	"github.com/spf13/cobra"
	"log"
	"strings"
	"sync"
)

const downloadCountDlMysqlUser = "root"

const downloadCountDlMysqlPassword = "npm-analysis"

const downloadCountDlMongoUrl = "mongodb://npm:npm123@localhost:27017"

// NPM is rate-limiting so don't go over 8 workers here
const downloadCountDlWorkerNumber = 3

var downloadCountDlIsBulk bool

var downloadCountDlCmd = &cobra.Command{
	Use:   "downloadCountDl",
	Short: "Retrieve download numbers",
	Long:  `...`,
	Run: func(cmd *cobra.Command, args []string) {
		mysqlInitializer := &database.Mysql{}
		mysql, databaseInitErr := mysqlInitializer.InitDB(fmt.Sprintf("%s:%s@/npm?charset=utf8mb4&collation=utf8mb4_bin", downloadCountDlMysqlUser, downloadCountDlMysqlPassword))
		if databaseInitErr != nil {
			log.Fatal(databaseInitErr)
		}
		defer mysql.Close()

		workerWait := sync.WaitGroup{}

		jobs := make(chan string, 100)

		if downloadCountDlIsBulk {
			log.Printf("Using bulk queries")
		} else {
			log.Printf("Using single package queries")
		}

		for w := 1; w <= downloadCountDlWorkerNumber; w++ {
			workerWait.Add(1)
			if downloadCountDlIsBulk {
				go downloadCountDlWorkerBulk(w, jobs, &workerWait)
			} else {
				go downloadCountDlWorker(w, jobs, &workerWait)
			}
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

		workerWait.Wait()
	},
}

func init() {
	rootCmd.AddCommand(downloadCountDlCmd)

	downloadCountDlCmd.Flags().BoolVar(&downloadCountDlIsBulk, "bulk", true, "use bulk queries?")
}

func downloadCountDlWorker(workerId int, jobs chan string, workerWait *sync.WaitGroup) {
	mongoDB := database.NewMongoDB(downloadCountDlMongoUrl, "npm", "downloads")
	mongoDB.Connect()
	defer mongoDB.Disconnect()
	log.Printf("logged in mongo - workerId %v", workerId)

	ensureIndex(mongoDB)

	for pkg := range jobs {
		if val, err := mongoDB.FindOneSimple("key", pkg); val != "" && err == nil {
			downloadCount := evolution.DownloadCountResponse{}
			err := json.Unmarshal([]byte(val), &downloadCount)
			if err != nil {
				log.Fatalf("ERROR: Cannot unmarshal DownloadCountResponse from db with error %v", err)
			}
			if len(downloadCount.DownloadCounts) == 1190 {
				//log.Printf("Package %v already exists", pkg)
				continue
			} else {
				err := mongoDB.RemoveWithKey(pkg)
				if err != nil {
					log.Fatalf("ERROR: could not remove already existing but wrong data for package %v", pkg)
				}
			}
		}

		doc, err := evolution.GetDownloadCountsForPackage(pkg)
		if err != nil {
			log.Printf("ERROR: %v", err)
			jobs <- pkg
		}

		mongoDB.InsertOneSimple(pkg, doc)
		if err != nil {
			log.Fatalf("ERROR: inserting %v into mongo with %s", pkg, err)
		}

		log.Printf("Inserted download counts of %v downloadCountDlWorker: %v", pkg, workerId)
	}

	workerWait.Done()
	log.Println("send finished downloadCountDlWorker ", workerId)
}

func downloadCountDlWorkerBulk(workerId int, jobs chan string, workerWait *sync.WaitGroup) {
	mongoDB := database.NewMongoDB(downloadCountDlMongoUrl, "npm", "downloads")
	mongoDB.Connect()
	defer mongoDB.Disconnect()
	log.Printf("logged in mongo - workerId %v", workerId)

	ensureIndex(mongoDB)

	var bulkPackages []string

	for pkg := range jobs {
		if val, err := mongoDB.FindOneSimple("key", pkg); val != "" && err == nil {
			downloadCount := evolution.DownloadCountResponse{}
			err := json.Unmarshal([]byte(val), &downloadCount)
			if err != nil {
				log.Fatalf("ERROR: Cannot unmarshal DownloadCountResponse from db with error %v", err)
			}
			if len(downloadCount.DownloadCounts) == 1190 {
				//log.Printf("Package %v already exists", pkg)
				continue
			} else {
				err := mongoDB.RemoveWithKey(pkg)
				if err != nil {
					log.Fatalf("ERROR: could not remove already existing but wrong data for package %v", pkg)
				}
			}
		}

		if strings.HasPrefix(pkg, "@") {
			log.Printf("WARNING: Package %v unsupported for bulk download", pkg)
			continue
		}

		bulkPackages = append(bulkPackages, pkg)

		if len(bulkPackages) == 128 {
			downloadCountDlProcessBulk(bulkPackages, mongoDB, workerId)

			bulkPackages = make([]string, 0)
		}
	}

	if len(bulkPackages) > 0 {
		downloadCountDlProcessBulk(bulkPackages, mongoDB, workerId)
	}

	workerWait.Done()
	log.Println("send finished downloadCountDlWorker ", workerId)
}

func downloadCountDlProcessBulk(bulkPackages []string, mongoDB *database.MongoDB, workerId int) {
	bulk, err := evolution.GetDownloadCountsBulk(bulkPackages)
	if err != nil {
		log.Fatalf("ERROR: %v", err)
	}
	for p, d := range bulk {
		mongoDB.InsertOneSimple(p, d)
		if err != nil {
			log.Fatalf("ERROR: inserting %v into mongo with %s", p, err)
		}

		log.Printf("Inserted download counts of %v downloadCountDlWorker: %v", p, workerId)
	}
}
