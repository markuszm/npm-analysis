package cmd

import (
	"database/sql"
	"fmt"
	"github.com/markuszm/npm-analysis/database"
	"github.com/markuszm/npm-analysis/evolution"
	"github.com/markuszm/npm-analysis/plots"
	"github.com/spf13/cobra"
	"log"
	"os"
	"sync"
	"time"
)

const plotterMysqlUser = "root"

const plotterMysqlPassword = "npm-analysis"

var plotterDB *sql.DB

var plotterWorkerNumber = 100

var plotterCreatePlot bool

var plotterPlotType string

var plotterCmd = &cobra.Command{
	Use:   "plotter",
	Short: "Plots maintainer box plot or retrieve data as JSON",
	Long:  `...`,
	Run: func(cmd *cobra.Command, args []string) {
		mysqlInitializer := &database.Mysql{}
		mysql, err := mysqlInitializer.InitDB(fmt.Sprintf("%s:%s@/npm?charset=utf8mb4&collation=utf8mb4_bin", plotterMysqlUser, plotterMysqlPassword))
		if err != nil {
			log.Fatal(err)
		}
		defer mysql.Close()

		plotterDB = mysql

		if plotterPlotType == "lineplot" {
			workerWait := sync.WaitGroup{}

			jobs := make(chan string, 100)

			for w := 1; w <= plotterWorkerNumber; w++ {
				workerWait.Add(1)
				go plotterWorker(w, jobs, &workerWait)
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

		if plotterPlotType == "boxplot" {
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
	},
}

func init() {
	rootCmd.AddCommand(plotterCmd)

	plotterCmd.Flags().StringVar(&plotterPlotType, "type", "", "specify which plot type")
	plotterCmd.Flags().BoolVar(&plotterCreatePlot, "createPlot", false, "whether to create plot or just get values as json")
}

func plotterWorker(id int, jobs chan string, workerWait *sync.WaitGroup) {
	for maintainerName := range jobs {
		fileName := plots.GetPlotFileName(maintainerName, "maintainer-evolution")
		if _, err := os.Stat(fileName); err != nil {
			plots.CreateLinePlotForMaintainerPackageCount(maintainerName, plotterDB, plotterCreatePlot)
		}
	}
	workerWait.Done()
}
