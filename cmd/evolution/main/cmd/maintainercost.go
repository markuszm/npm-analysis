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

const maintainerCostMongoUrl = "mongodb://npm:npm123@localhost:27017"

const maintainerCostMaintainersTimelineJsonPath = "./db-data/maintainersTimeline.json"

const maintainerCostDependenciesTimelineJsonPath = "./db-data/dependenciesTimeline.json"

var maintainerCostWorkerNumber int

var maintainerCostCreatePlot bool

var maintainerCostPackageFileInput string

var maintainerCostResultMap = sync.Map{}

var maintainerCostGenerateData bool

var maintainerCostCmd = &cobra.Command{
	Use:   "maintainerCost",
	Short: "Calculates Package Cost of a maintainer and plots it",
	Long:  `...`,
	Run: func(cmd *cobra.Command, args []string) {
		if maintainerCostGenerateData {
			reach.GenerateTimeMaintainersMap(maintainerCostMongoUrl, maintainerCostMaintainersTimelineJsonPath)
		}

		maintainerCostCalculate()
	},
}

func init() {
	rootCmd.AddCommand(maintainerCostCmd)

	maintainerCostCmd.Flags().BoolVar(&maintainerCostCreatePlot, "createPlot", false, "whether to create plots")
	maintainerCostCmd.Flags().BoolVar(&maintainerCostGenerateData, "generateData", false, "whether to generate cached data")
	maintainerCostCmd.Flags().IntVar(&maintainerCostWorkerNumber, "workers", 100, "number of workers")
	maintainerCostCmd.Flags().StringVar(&maintainerCostPackageFileInput, "packageInput", "", "input file containing packages")
}

func maintainerCostCalculate() {
	maintainersTimeline := reach.LoadJSONMaintainersTimeline(maintainerCostMaintainersTimelineJsonPath)
	dependenciesTimeline := reach.LoadJSONDependenciesTimeline(maintainerCostDependenciesTimelineJsonPath)

	startTime := time.Now()

	workerWait := sync.WaitGroup{}

	jobs := make(chan string, 100)

	go streamPackageNamesFromFile(jobs, maintainerCostPackageFileInput)

	for w := 1; w <= maintainerCostWorkerNumber; w++ {
		workerWait.Add(1)
		go maintainerCostWorker(w, jobs, dependenciesTimeline, maintainersTimeline, &workerWait)
	}

	workerWait.Wait()

	endTime := time.Now()

	log.Printf("Took %v minutes to process all packages", endTime.Sub(startTime).Minutes())

	reach.CalculateAverageMaintainerReach("averageMaintainerCost", &maintainerCostResultMap)

	reach.CalculateMaintainerReachDiff("maintainerCostDiff", &maintainerCostResultMap)
}

func maintainerCostWorker(workerId int, jobs chan string, dependenciesTimeline map[time.Time]map[string]map[string]bool, maintainersTimeline map[time.Time]map[string][]string, workerWait *sync.WaitGroup) {
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

				calculateMaintainerCost(pkg, maintainers, date, dependenciesTimeline, maintainersTimeline)

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

func calculateMaintainerCost(pkg string, maintainers map[string]bool, date time.Time, dependenciesTimeline map[time.Time]map[string]map[string]bool, maintainersTimeline map[time.Time]map[string][]string) {
	for _, m := range maintainersTimeline[date][pkg] {
		maintainers[m] = true
	}

	for dep, ok := range dependenciesTimeline[date][pkg] {
		if ok {
			calculateMaintainerCost(dep, maintainers, date, dependenciesTimeline, maintainersTimeline)
		}
	}
}
