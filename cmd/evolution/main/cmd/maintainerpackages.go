package cmd

import (
	"encoding/json"
	"fmt"
	"github.com/markuszm/npm-analysis/database"
	"github.com/markuszm/npm-analysis/evolution"
	"github.com/markuszm/npm-analysis/util"
	"github.com/spf13/cobra"
	"log"
	"sync"
	"time"
)

const maintainerPackagesMongoUrl = "mongodb://npm:npm123@localhost:27017"

const maintainerPackagesWorkerNumber = 75

const maintainerPackagesMysqlUser = "root"

const maintainerPackagesMysqlPassword = "npm-analysis"

// Collects all packages that are maintained by a maintainer for a specific time and stores into mongo in collection "maintainerPackages"
var maintainerPackagesCmd = &cobra.Command{
	Use:   "maintainerPackages",
	Short: "Create packages of maintainer aggregation",
	Long:  `Collects all packages that are maintained by a maintainer for a specific time and stores into mongo in collection "maintainerPackages"`,
	Run: func(cmd *cobra.Command, args []string) {
		mysqlInitializer := &database.Mysql{}
		mysql, databaseInitErr := mysqlInitializer.InitDB(fmt.Sprintf("%s:%s@/npm?charset=utf8mb4&collation=utf8mb4_bin", maintainerPackagesMysqlUser, maintainerPackagesMysqlPassword))
		if databaseInitErr != nil {
			log.Fatal(databaseInitErr)
		}
		defer mysql.Close()

		changes, err := database.GetMaintainerChanges(mysql)
		if err != nil {
			log.Fatalf("ERROR: loading changes from mysql with %v", err)
		}

		log.Print("Finished retrieving changes from db")

		maintainedPackages := evolution.CalculateMaintainerPackages(changes)

		startTime := time.Now()

		workerWait := sync.WaitGroup{}

		jobs := make(chan StoreMaintainedPackages, 100)

		for w := 1; w <= maintainerPackagesWorkerNumber; w++ {
			workerWait.Add(1)
			go maintainerPackagesWorker(w, jobs, &workerWait)
		}

		for _, v := range maintainedPackages {
			packageTimeline := make(map[time.Time][]string)

			for year, monthMap := range v.Packages {
				for month, packages := range monthMap {
					var keys []string
					for p, ok := range packages {
						if ok {
							keys = append(keys, p)
						}
					}
					packageTimeline[time.Date(year, time.Month(month), 1, 0, 0, 0, 0, time.UTC)] = keys
				}
			}
			s := StoreMaintainedPackages{
				Name:             v.Name,
				PackagesTimeline: packageTimeline,
			}

			jobs <- s
		}

		close(jobs)

		workerWait.Wait()

		endTime := time.Now()

		log.Printf("Took %v minutes to process all Documents from MongoDB", endTime.Sub(startTime).Minutes())
	},
}

func init() {
	rootCmd.AddCommand(maintainerPackagesCmd)
}

func maintainerPackagesWorker(id int, jobs chan StoreMaintainedPackages, workerWait *sync.WaitGroup) {
	mongoDB := database.NewMongoDB(maintainerPackagesMongoUrl, "npm", "maintainerPackages")
	mongoDB.Connect()
	defer mongoDB.Disconnect()
	log.Printf("logged in mongo - workerId %v", id)

	ensureIndex(mongoDB)
	for j := range jobs {
		maintainerPackagesProcessDocument(j, mongoDB)
	}
	workerWait.Done()
}

func maintainerPackagesProcessDocument(pkgs StoreMaintainedPackages, mongoDB *database.MongoDB) {
	if val, err := mongoDB.FindOneSimple("key", pkgs.Name); val != "" && err == nil {
		log.Printf("Package %v already exists", pkgs.Name)

		val, err := util.Decompress(val)
		if err != nil {
			log.Fatalf("ERROR: Decompressing: %v", err)
		}

		if val == "" {
			err := mongoDB.RemoveWithKey(pkgs.Name)
			if err != nil {
				log.Fatalf("ERROR: could not remove already existing but wrong data for package %v", pkgs.Name)
			}
		} else {
			return
		}
	}

	bytes, err := json.Marshal(pkgs)
	if err != nil {
		log.Fatalf("ERROR: marshalling package data for %v with %v", pkgs.Name, err)
	}

	err = mongoDB.InsertOneSimple(pkgs.Name, string(bytes))
	if err != nil {
		log.Fatalf("ERROR: inserting package %v into mongo with %v", pkgs.Name, err)
	}
}
