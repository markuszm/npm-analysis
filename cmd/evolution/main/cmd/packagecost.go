package cmd

import (
	reach "github.com/markuszm/npm-analysis/evolution/maintainerreach"
	"github.com/markuszm/npm-analysis/plots"
	"github.com/spf13/cobra"
	"log"
	"os"
	"sync"
	"time"
)

const packageCostJsonPath = "./db-data/dependenciesTimeline.json"

var packageCostWorkerNumber int

var packageCostResultMap = sync.Map{}

var packageCostCreatePlot bool

var packageCostPackageFileInput string

var packageCostOutputFolder string

// calculates Package reach of a packages and plots it
var packageCostCmd = &cobra.Command{
	Use:   "packageCost",
	Short: "calculates Package cost of all packages and plots it",
	Long:  `...`,
	Run: func(cmd *cobra.Command, args []string) {
		initializeLogger()

		calculatePackageCost()
	},
}

func init() {
	rootCmd.AddCommand(packageCostCmd)

	packageCostCmd.Flags().BoolVar(&packageCostCreatePlot, "createPlot", false, "whether it should create plots for each package")
	packageCostCmd.Flags().IntVar(&packageCostWorkerNumber, "workers", 100, "number of workers")
	packageCostCmd.Flags().StringVar(&packageCostPackageFileInput, "packageInput", "", "input file containing packages")
	packageCostCmd.Flags().StringVar(&packageCostOutputFolder, "output", "/home/markus/npm-analysis/", "output folder for results")
}

func calculatePackageCost() {
	dependenciesTimeline := reach.LoadJSONDependenciesTimeline(packageCostJsonPath)

	startTime := time.Now()

	workerWait := sync.WaitGroup{}

	jobs := make(chan string, 100)

	go streamPackageNamesFromFile(jobs, packageCostPackageFileInput)

	for w := 1; w <= packageCostWorkerNumber; w++ {
		workerWait.Add(1)
		go packageCostWorker(w, jobs, dependenciesTimeline, &workerWait)
	}

	workerWait.Wait()

	endTime := time.Now()

	log.Printf("Took %v minutes to process all Documents from MongoDB", endTime.Sub(startTime).Minutes())

	reach.CalculateAverageResults("averagePackageCost", packageCostOutputFolder, &packageCostResultMap)

	reach.CalculateMaintainerReachDiff("packageCostDiff", packageCostOutputFolder, &packageCostResultMap)

	err := reach.CalculatePackageReachDiffs(&packageCostResultMap, "packageCostDiffs", packageCostOutputFolder)
	if err != nil {
		log.Fatal(err)
	}
}

func packageCostWorker(workerId int, jobs chan string, dependencies map[time.Time]map[string]map[string]bool, workerWait *sync.WaitGroup) {
	for j := range jobs {
		var counts []float64
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
				packages := make(map[string]bool, 0)
				reach.PackageCost(j, dependencies[date], packages)
				counts = append(counts, float64(len(packages)))
			}
		}

		packageCostResultMap.Store(j, counts)

		outputFolder := "package-cost"
		if packageCostCreatePlot {
			fileName := plots.GetPlotFileName(j, outputFolder)
			if _, err := os.Stat(fileName); err == nil {
				continue
			}
			plots.GenerateLinePlotForMaintainerReach(outputFolder, j, counts, packageCostCreatePlot)
		} else {
			plots.GenerateLinePlotForMaintainerReach(outputFolder, j, counts, packageCostCreatePlot)
		}

		//log.Printf("Finished %v", j.Name)
	}
	workerWait.Done()
}
