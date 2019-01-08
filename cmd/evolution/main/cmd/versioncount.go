package cmd

import (
	"database/sql"
	"fmt"
	"github.com/markuszm/npm-analysis/database"
	"github.com/markuszm/npm-analysis/database/insert"
	"github.com/markuszm/npm-analysis/evolution"
	"github.com/spf13/cobra"
	"log"
	"sync"
)

const versionCountMysqlUser = "root"

const versionCountMysqlPassword = "npm-analysis"

const versionCountWorkerNumber = 100

var versionCountDB *sql.DB

// Calculates version count based on version changes
var versionCountCmd = &cobra.Command{
	Use:   "versionCount",
	Short: "Calculates version count based on version changes",
	Long:  `...`,
	Run: func(cmd *cobra.Command, args []string) {
		mysqlInitializer := &database.Mysql{}
		versionCountDB, databaseInitErr := mysqlInitializer.InitDB(fmt.Sprintf("%s:%s@/npm?charset=utf8mb4&collation=utf8mb4_bin", versionCountMysqlUser, versionCountMysqlPassword))
		if databaseInitErr != nil {
			log.Fatal(databaseInitErr)
		}
		defer versionCountDB.Close()

		packages, err := database.GetPackages(versionCountDB)
		if err != nil {
			log.Fatalf("ERROR: loading packages from mysql with %v", err)
		}

		log.Print("Finished retrieving packages from db")

		err = database.CreateVersionCount(versionCountDB)
		if err != nil {
			log.Fatalf("ERROR: creating table with %v", err)
		}

		workerWait := sync.WaitGroup{}

		jobs := make(chan string, 100)

		for w := 1; w <= versionCountWorkerNumber; w++ {
			workerWait.Add(1)
			go versionCountWorker(w, jobs, &workerWait)
		}

		for i, p := range packages {
			if i%10000 == 0 {
				log.Printf("Finished %v packages", i)
			}
			jobs <- p
		}

		close(jobs)
		workerWait.Wait()
	},
}

func init() {
	rootCmd.AddCommand(versionCountCmd)
}

func versionCountWorker(id int, jobs chan string, workerWait *sync.WaitGroup) {
	for p := range jobs {
		versionChanges, err := database.GetVersionChangesForPackage(p, versionCountDB)
		if err != nil {
			log.Fatalf("ERROR: retrieving version changes for package %v with %v", p, err)
		}

		evolution.SortVersionChange(versionChanges)

		versionCount := evolution.CountVersions(versionChanges)

		err = insert.StoreVersionCount(versionCountDB, p, versionCount)
		if err != nil {
			log.Fatalf("ERROR: writing to database with %v", err)
		}
	}
	workerWait.Done()
}
