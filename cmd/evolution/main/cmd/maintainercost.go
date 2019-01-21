package cmd

import (
	"encoding/json"
	"fmt"
	reach "github.com/markuszm/npm-analysis/evolution/maintainerreach"
	"github.com/markuszm/npm-analysis/plots"
	"github.com/markuszm/npm-analysis/util"
	"github.com/spf13/cobra"
	"io/ioutil"
	"log"
	"os"
	"path"
	"sort"
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

var maintainerCostIsEvolution bool

var maintainerCostIsTrustedMaintainers bool
var maintainerCostMaintainerReachRankingInput string

var maintainerCostOutputFolder string

var maintainerCostDependenciesTimeline map[time.Time]map[string]map[string]bool
var maintainerCostMaintainerTimeline map[time.Time]map[string][]string

var latestDependencies map[string]map[string]bool
var latestMaintainers map[string][]string

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
	maintainerCostCmd.Flags().BoolVar(&maintainerCostIsEvolution, "evolution", false, "whether to calculate evolution of maintainer cost")
	maintainerCostCmd.Flags().BoolVar(&maintainerCostIsTrustedMaintainers, "trusted", false, "whether to calculate trusted aggregation")
	maintainerCostCmd.Flags().StringVar(&maintainerCostMaintainerReachRankingInput, "reachOrder", "", "input file containing maintainer reach order")
	maintainerCostCmd.Flags().BoolVar(&maintainerCostGenerateData, "generateData", false, "whether to generate cached data")
	maintainerCostCmd.Flags().IntVar(&maintainerCostWorkerNumber, "workers", 8, "number of workers")
	maintainerCostCmd.Flags().StringVar(&maintainerCostPackageFileInput, "packageInput", "", "input file containing packages")
	maintainerCostCmd.Flags().StringVar(&maintainerCostOutputFolder, "output", "/home/markus/npm-analysis/", "output folder for results")
}

func maintainerCostCalculate() {
	if !maintainerCostIsEvolution {
		date := time.Date(2018, time.Month(4), 1, 0, 0, 0, 0, time.UTC)
		maintainerCostMaintainerTimeline := reach.LoadJSONMaintainersTimeline(maintainerCostMaintainersTimelineJsonPath)
		latestMaintainers = maintainerCostMaintainerTimeline[date]

		maintainerCostDependenciesTimeline := reach.LoadJSONDependenciesTimeline(maintainerCostDependenciesTimelineJsonPath)
		latestDependencies = maintainerCostDependenciesTimeline[date]
	} else {
		maintainerCostMaintainerTimeline = reach.LoadJSONMaintainersTimeline(maintainerCostMaintainersTimelineJsonPath)
		maintainerCostDependenciesTimeline = reach.LoadJSONDependenciesTimeline(maintainerCostDependenciesTimelineJsonPath)
	}

	if maintainerCostIsTrustedMaintainers {
		trustedMaintainersSet := make(map[string]bool, 0)

		var maintainerReachRanking []string

		var averages []float64

		bytes, err := ioutil.ReadFile(maintainerCostMaintainerReachRankingInput)
		if err != nil {
			log.Fatal(err)
		}

		err = json.Unmarshal(bytes, &maintainerReachRanking)
		if err != nil {
			log.Fatal(err)
		}

		packages := getPackageNamesFromFile(maintainerCostPackageFileInput)

		cost := calculateAverageMaintainerCost(packages, trustedMaintainersSet)

		averages = append(averages, cost)

		for _, m := range maintainerReachRanking {
			trustedMaintainersSet[m] = true

			cost := calculateAverageMaintainerCost(packages, trustedMaintainersSet)

			averages = append(averages, cost)

			log.Printf("added %v as trusted maintainer", m)
		}

		bytes, err = json.Marshal(averages)
		if err != nil {
			log.Fatal(err)
		}

		outputPath := path.Join(maintainerCostOutputFolder, fmt.Sprintf("%s.json", "trustedMaintainersCost"))
		err = ioutil.WriteFile(outputPath, bytes, os.ModePerm)

		return
	}

	startTime := time.Now()

	workerWait := sync.WaitGroup{}

	jobs := make(chan string, 100)

	go streamPackageNamesFromFile(jobs, maintainerCostPackageFileInput)

	for w := 1; w <= maintainerCostWorkerNumber; w++ {
		workerWait.Add(1)
		go maintainerCostWorker(w, jobs, &workerWait)
	}

	workerWait.Wait()

	endTime := time.Now()

	log.Printf("Took %v minutes to process all packages", endTime.Sub(startTime).Minutes())

	if maintainerCostIsEvolution {
		reach.CalculateAverageResults("averageMaintainerCost", maintainerCostOutputFolder, &maintainerCostResultMap)

		reach.CalculateMaintainerReachDiff("maintainerCostDiff", maintainerCostOutputFolder, &maintainerCostResultMap)
	} else {
		var countPairs []util.FloatPair

		maintainerCostResultMap.Range(func(key, value interface{}) bool {
			pkg := key.(string)
			count := value.(float64)

			countPairs = append(countPairs, util.FloatPair{
				Key:   pkg,
				Value: count,
			})
			return true
		})

		sort.Sort(sort.Reverse(util.FloatPairList(countPairs)))

		bytes, err := json.Marshal(countPairs)
		if err != nil {
			log.Fatal(err)
		}

		outputPath := path.Join(maintainerCostOutputFolder, fmt.Sprintf("%s.json", "maintainerCostLatest"))
		err = ioutil.WriteFile(outputPath, bytes, os.ModePerm)
	}

}

func maintainerCostWorker(workerId int, jobs chan string, workerWait *sync.WaitGroup) {
	for pkg := range jobs {

		if maintainerCostIsEvolution {
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

					maintainers := make(map[string]bool)
					visited := make(map[string]bool)

					calculateMaintainerCost(pkg, maintainers, visited, maintainerCostMaintainerTimeline[date], maintainerCostDependenciesTimeline[date])

					counts = append(counts, float64(len(maintainers)))
				}
			}

			maintainerCostResultMap.Store(pkg, counts)

			outputFolder := path.Join(maintainerCostOutputFolder, "maintainer-cost")
			if maintainerCostCreatePlot {
				fileName := plots.GetPlotFileName(pkg, outputFolder)
				if _, err := os.Stat(fileName); err == nil {
					continue
				}
				plots.GenerateLinePlotForMaintainerReach(outputFolder, pkg, counts, maintainerCostCreatePlot)
			} else {
				plots.GenerateLinePlotForMaintainerReach(outputFolder, pkg, counts, maintainerCostCreatePlot)
			}

		} else {
			maintainers := make(map[string]bool)
			visited := make(map[string]bool)

			calculateMaintainerCost(pkg, maintainers, visited, latestMaintainers, latestDependencies)

			count := float64(len(maintainers))

			maintainerCostResultMap.Store(pkg, count)
		}

		log.Printf("Worker %v Finished %v", workerId, pkg)
	}
	workerWait.Done()
}

func calculateAverageMaintainerCost(packages []string, trustedMaintainers map[string]bool) float64 {
	totalCost := 0

	for _, pkg := range packages {
		maintainers := make(map[string]bool)
		visited := make(map[string]bool)

		// only first party protection
		calculateMaintainerCost(pkg, maintainers, visited, latestMaintainers, latestDependencies)

		for m, ok := range maintainers {
			if ok {
				if !trustedMaintainers[m] {
					totalCost += 1
				}
			}
		}
	}

	return float64(totalCost) / float64(len(packages))
}

func calculateMaintainerCost(pkg string, maintainers map[string]bool, visited map[string]bool, maintainersMap map[string][]string, dependenciesMap map[string]map[string]bool) {
	for _, m := range maintainersMap[pkg] {
		maintainers[m] = true
	}

	for dep, ok := range dependenciesMap[pkg] {
		if ok && !visited[dep] {
			visited[dep] = true
			calculateMaintainerCost(dep, maintainers, visited, maintainersMap, dependenciesMap)
		}
	}
}
