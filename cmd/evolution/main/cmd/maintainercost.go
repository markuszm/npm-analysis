package cmd

import (
	"github.com/markuszm/npm-analysis/database"
	"github.com/markuszm/npm-analysis/evolution"
	reach "github.com/markuszm/npm-analysis/evolution/maintainerreach"
	"github.com/markuszm/npm-analysis/plots"
	"github.com/spf13/cobra"
	"log"
	"os"
	"sync"
	"time"
)

const maintainerCostMongoUrl = "mongodb://npm:npm123@localhost:27017"

const maintainerCostJsonPath = "./db-data/dependenciesTimeline.json"

var maintainerCostWorkerNumber int

var maintainerCostCreatePlot bool

var maintainerCostPackageFileInput string

var maintainerCostResultMap = sync.Map{}

var maintainerCostCmd = &cobra.Command{
	Use:   "maintainerCost",
	Short: "Calculates Package Cost of a maintainer and plots it",
	Long:  `...`,
	Run: func(cmd *cobra.Command, args []string) {
		if maintainerReachGenerateData {
			reach.GenerateTimeLatestVersionMap(maintainerCostMongoUrl, maintainerCostJsonPath)
		}

		maintainerCostCalculate()
	},
}

func init() {
	rootCmd.AddCommand(maintainerCostCmd)

	maintainerCostCmd.Flags().BoolVar(&maintainerCostCreatePlot, "createPlot", false, "whether to create plots")
	maintainerCostCmd.Flags().IntVar(&maintainerCostWorkerNumber, "workers", 100, "number of workers")
	maintainerCostCmd.Flags().StringVar(&maintainerCostPackageFileInput, "packageInput", "", "input file containing packages")
}

func maintainerCostCalculate() {
	//dependenciesTimeline := reach.LoadJSONDependenciesTimeline(maintainerReachJsonPath)

	mongoDB := database.NewMongoDB(maintainerReachMongoUrl, "npm", "timelineNew")
	mongoDB.Connect()
	defer mongoDB.Disconnect()

	log.Print("Connected to mongodb")

	startTime := time.Now()

	workerWait := sync.WaitGroup{}

	jobs := make(chan string, 100)

	go streamPackageNamesFromFile(jobs, maintainerCostPackageFileInput)

	for w := 1; w <= maintainerCostWorkerNumber; w++ {
		workerWait.Add(1)
		go maintainerCostWorker(w, jobs, mongoDB, &workerWait)
	}

	workerWait.Wait()

	endTime := time.Now()

	log.Printf("Took %v minutes to process all Documents from MongoDB", endTime.Sub(startTime).Minutes())

	reach.CalculateAverageMaintainerReach("averageMaintainerCost", &maintainerCostResultMap)

	reach.CalculateMaintainerReachDiff("maintainerCostDiff", &maintainerCostResultMap)
}

func maintainerCostWorker(workerId int, jobs chan string, mongoDB *database.MongoDB, workerWait *sync.WaitGroup) {
	for j := range jobs {
		var counts []float64
		pkg := j

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
				date := time.Date(year, time.Month(month), 1, 0, 0, 0, 0, time.UTC)

				maintainers := make(map[string]bool)

				calculateMaintainerCost(pkg, maintainers, mongoDB, date)

				counts = append(counts, float64(len(maintainers)))
			}
		}

		maintainerCostResultMap.Store(pkg, counts)

		outputFolder := "maintainer-cost"
		if maintainerCostCreatePlot {
			fileName := plots.GetPlotFileName(pkg, outputFolder)
			if _, err := os.Stat(fileName); err == nil {
				continue
			}
			plots.GenerateLinePlotForMaintainerReach(outputFolder, pkg, counts, maintainerCostCreatePlot)
		} else {
			plots.GenerateLinePlotForMaintainerReach(outputFolder, pkg, counts, maintainerCostCreatePlot)
		}

		log.Printf("Worker %v Finished %v", workerId, pkg)
	}
	workerWait.Done()
}

func getPackageDataForDate(mongoDB *database.MongoDB, pkg string, date time.Time) evolution.PackageData {
	packageData, err := mongoDB.FindPackageDataInTimeline(pkg, date)
	if err != nil {
		log.Printf("did not find package even though it should exist with error: %v for package %v", err.Error(), pkg)
		return evolution.PackageData{}
	}
	return packageData
}

func calculateMaintainerCost(pkg string, maintainers map[string]bool, mongoDB *database.MongoDB, date time.Time) {
	packageData := getPackageDataForDate(mongoDB, pkg, date)

	for _, m := range packageData.Maintainers {
		maintainers[m] = true
	}

	for _, dep := range packageData.Dependencies {
		calculateMaintainerCost(dep, maintainers, mongoDB, date)
	}
}
