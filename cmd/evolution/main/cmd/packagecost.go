package cmd

import (
	"encoding/json"
	"fmt"
	reach "github.com/markuszm/npm-analysis/evolution/maintainerreach"
	"github.com/markuszm/npm-analysis/plots"
	"github.com/spf13/cobra"
	"io/ioutil"
	"log"
	"os"
	"path"
	"sync"
	"time"
)

const packageCostJsonPath = "./db-data/dependenciesTimeline.json"
const packageCostLatestDependenciesJsonPath = "./db-data/latestDependenciesTimeline.json"

var packageCostWorkerNumber int

var packageCostIsTrustedPackages bool

var packageCostLimit int

var packageCostResultMap = sync.Map{}

var packageCostCreatePlot bool

var packageCostPackageFileInput string

var packageCostOutputFolder string

var packageCostReachRankingInput string

var trustedPackagesCostResultMap = sync.Map{}
var trustedPackagesSecuredResultMap = sync.Map{}

var packageCostLatestDependencies map[string]map[string]bool

// calculates Package reach of a packages and plots it
var packageCostCmd = &cobra.Command{
	Use:   "packageCost",
	Short: "calculates Package cost of all packages and plots it",
	Long:  `...`,
	Run: func(cmd *cobra.Command, args []string) {
		initializeLogger()

		if packageCostIsTrustedPackages {
			packageCostLatestDependencies = reach.LoadJSONLatestDependencies(packageCostLatestDependenciesJsonPath)
			calculateTrustedAggregation()
		} else {
			calculatePackageCost()
		}
	},
}

func init() {
	rootCmd.AddCommand(packageCostCmd)

	packageCostCmd.Flags().BoolVar(&packageCostCreatePlot, "createPlot", false, "whether it should create plots for each package")
	packageCostCmd.Flags().IntVar(&packageCostWorkerNumber, "workers", 100, "number of workers")
	packageCostCmd.Flags().IntVar(&packageCostLimit, "limit", 500, "limit how long trusted packages aggregation should calculate")
	packageCostCmd.Flags().StringVar(&packageCostPackageFileInput, "packageInput", "", "input file containing packages")
	packageCostCmd.Flags().StringVar(&packageCostReachRankingInput, "rankingInput", "", "input file containing ranking of packages by reach")
	packageCostCmd.Flags().StringVar(&packageCostOutputFolder, "output", "/home/markus/npm-analysis/", "output folder for results")
	packageCostCmd.Flags().BoolVar(&packageCostIsTrustedPackages, "trusted", false, "whether to calculate trusted aggregation")

}

func calculateTrustedAggregation() {
	trustedPackages := make(map[string]bool, 0)

	var packageReachRanking []string

	var costAverages []float64
	var fullySecuredPercentages []float64

	bytes, err := ioutil.ReadFile(packageCostReachRankingInput)
	if err != nil {
		log.Fatal(err)
	}

	err = json.Unmarshal(bytes, &packageReachRanking)
	if err != nil {
		log.Fatal(err)
	}

	packages := getPackageNamesFromFile(packageCostPackageFileInput)

	allResults := calculateAllPackageCosts(packages)

	cost := calculateAverageForTrustedPackages(allResults, len(packages), trustedPackages)
	costAverages = append(costAverages, cost)

	fullySecured := calculateFullySecurePackagesByPackages(allResults, len(packages), trustedPackages)
	fullySecuredPercentages = append(fullySecuredPercentages, fullySecured)

	log.Printf("calculated avg without trusted packages: %v", cost)

	workerWait := sync.WaitGroup{}

	jobs := make(chan TrustedPackageJobItem, 100)

	for w := 1; w <= packageCostWorkerNumber; w++ {
		workerWait.Add(1)
		go trustedPackagesWorker(jobs, allResults, len(packages), &workerWait)
	}

	for i, m := range packageReachRanking {
		trustedPackages[m] = true

		copiedMap := copyMap(trustedPackages)

		jobs <- TrustedPackageJobItem{
			AddedPackageName: m,
			TrustedPackages:  copiedMap,
		}

		if i > packageCostLimit {
			break
		}
	}

	close(jobs)
	workerWait.Wait()

	for i, m := range packageReachRanking {
		cost, ok := trustedPackagesCostResultMap.Load(m)
		if !ok {
			log.Printf("no cost result found after adding %v", m)
		}
		costAverages = append(costAverages, cost.(float64))

		fullySecured, ok := trustedPackagesSecuredResultMap.Load(m)
		if !ok {
			log.Printf("no fully secured packages result found after adding %v", m)
		}
		fullySecuredPercentages = append(fullySecuredPercentages, fullySecured.(float64))

		if i > packageCostLimit {
			break
		}
	}

	bytes, err = json.Marshal(costAverages)
	if err != nil {
		log.Fatal(err)
	}

	outputPath := path.Join(packageCostOutputFolder, fmt.Sprintf("%s.json", "trustedPackagesCost"))
	err = ioutil.WriteFile(outputPath, bytes, os.ModePerm)
	if err != nil {
		log.Fatal(err)
	}

	bytes, err = json.Marshal(fullySecuredPercentages)
	if err != nil {
		log.Fatal(err)
	}

	outputPath = path.Join(packageCostOutputFolder, fmt.Sprintf("%s.json", "trustedPackagesSecured"))
	err = ioutil.WriteFile(outputPath, bytes, os.ModePerm)
	if err != nil {
		log.Fatal(err)
	}

	return
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

func calculateAllPackageCosts(packages []string) []map[string]bool {
	workerWait := sync.WaitGroup{}
	resultWait := sync.WaitGroup{}

	jobs := make(chan string, 100)
	results := make(chan map[string]bool, 100)

	// only first party protection
	for w := 1; w <= packageCostWorkerNumber; w++ {
		workerWait.Add(1)
		go packageCostWorkerWithResults(w, jobs, results, &workerWait)
	}

	var allResults []map[string]bool
	go func() {
		resultWait.Add(1)
		allResults = packageCostResultWorker(results, &resultWait)
	}()

	for _, pkg := range packages {
		jobs <- pkg
	}

	close(jobs)
	workerWait.Wait()
	close(results)
	resultWait.Wait()

	return allResults
}

func packageCostWorkerWithResults(id int, jobs chan string, results chan map[string]bool, workerWait *sync.WaitGroup) {
	for pkg := range jobs {
		packages := make(map[string]bool)

		packages[pkg] = true

		// only first party protection
		reach.PackageCost(pkg, packageCostLatestDependencies, packages)

		results <- packages
	}
	workerWait.Done()
}

func packageCostResultWorker(results chan map[string]bool, workerWait *sync.WaitGroup) []map[string]bool {
	var allResults []map[string]bool

	for r := range results {
		allResults = append(allResults, r)
	}

	workerWait.Done()
	return allResults
}

func calculateAverageForTrustedPackages(results []map[string]bool, packageCount int, trustedPackages map[string]bool) float64 {
	totalCost := 0

	for _, r := range results {
		for m, ok := range r {
			if ok {
				if !trustedPackages[m] {
					totalCost += 1
				}
			}
		}
	}

	return float64(totalCost) / float64(packageCount)
}

func calculateFullySecurePackagesByPackages(results []map[string]bool, packageCount int, trustedPackages map[string]bool) float64 {
	securedPackages := 0

	for _, r := range results {
		isSecure := true
		for m, ok := range r {
			if ok {
				if !trustedPackages[m] {
					isSecure = false
				}
			}
		}
		if isSecure {
			securedPackages++
		}
	}

	return float64(securedPackages) / float64(packageCount)
}

type TrustedPackageJobItem struct {
	AddedPackageName string
	TrustedPackages  map[string]bool
}

func trustedPackagesWorker(jobs chan TrustedPackageJobItem, results []map[string]bool, packageCount int, workerWait *sync.WaitGroup) {
	for j := range jobs {
		cost := calculateAverageForTrustedPackages(results, packageCount, j.TrustedPackages)

		trustedPackagesCostResultMap.Store(j.AddedPackageName, cost)

		fullySecured := calculateFullySecurePackagesByPackages(results, packageCount, j.TrustedPackages)

		trustedPackagesSecuredResultMap.Store(j.AddedPackageName, fullySecured)

		log.Printf("added %v as trusted package - avg is now %v", j.AddedPackageName, cost)
	}
	workerWait.Done()
}
