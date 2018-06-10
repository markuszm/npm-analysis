package main

import (
	"database/sql"
	"flag"
	"fmt"
	"github.com/markuszm/npm-analysis/database"
	"github.com/markuszm/npm-analysis/evolution"
	"github.com/markuszm/npm-analysis/plots"
	"log"
	"os"
	"sync"
	"time"
)

const MYSQL_USER = "root"

const MYSQL_PW = "npm-analysis"

var db *sql.DB

var workerNumber = 100

func main() {
	plotType := flag.String("type", "", "specify which plot type")
	flag.Parse()

	mysqlInitializer := &database.Mysql{}
	mysql, err := mysqlInitializer.InitDB(fmt.Sprintf("%s:%s@/npm?charset=utf8mb4&collation=utf8mb4_bin", MYSQL_USER, MYSQL_PW))
	if err != nil {
		log.Fatal(err)
	}
	defer mysql.Close()

	db = mysql

	if *plotType == "lineplot" {
		workerWait := sync.WaitGroup{}

		jobs := make(chan string, 100)

		for w := 1; w <= workerNumber; w++ {
			workerWait.Add(1)
			go worker(w, jobs, &workerWait)
		}
		maintainerNames, err := database.GetMaintainerNames(mysql)
		if err != nil {
			log.Fatal(err)
		}
		log.Print("Retrieved maintainer names from database")

		for _, m := range maintainerNames {
			jobs <- m
		}

		close(jobs)

		workerWait.Wait()
	}

	if *plotType == "boxplot" {
		changes, err := database.GetMaintainerChanges(mysql)
		if err != nil {
			log.Fatalf("ERROR: loading changes from mysql with %v", err)
		}

		log.Print("Finished retrieving changes from db")

		countMap := evolution.CalculateMaintainerCounts(changes)

		allCounts := make(map[time.Time][]int)
		for _, counts := range countMap {
			isActive := false

			for year := 2010; year < 2019; year++ {
				startMonth := 1
				endMonth := 12
				if year == 2010 {
					startMonth = 11
				}
				if year == 2018 {
					endMonth = 4
				}
				for month := startMonth; month <= endMonth; month++ {
					count := counts.Counts[year][month]
					date := time.Date(year, time.Month(month), 1, 0, 0, 0, 0, time.UTC)
					values := allCounts[date]
					if values == nil {
						values = make([]int, 0)
					}
					if count > 0 || isActive {
						values = append(values, count)
						isActive = true
					}
					allCounts[date] = values
				}
			}
		}

		plots.CreateBoxPlot(allCounts)
	}

}

func worker(id int, jobs chan string, workerWait *sync.WaitGroup) {
	for maintainerName := range jobs {
		fileName := plots.GetPlotFileName(maintainerName)
		if _, err := os.Stat(fileName); err != nil {
			plots.CreateLinePlotForMaintainerPackageCount(maintainerName, db)
		}
	}
	workerWait.Done()
}
