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

const packageReachJSONPath = "./db-data/dependenciesTimeline.json"

const packageReachWorkerNumber = 100

var packageReachResultMap = sync.Map{}

var packageReachCreatePlot bool

var packageReachResultPath string

var packageReachPackageFileInput string

var packageReachPackageInput string

// calculates Package reach of a packages and plots it
var packageReachCmd = &cobra.Command{
	Use:   "packageReach",
	Short: "calculates Package reach of a packages and plots it",
	Long:  `...`,
	Run: func(cmd *cobra.Command, args []string) {
		initializeLogger()

		calculatePackageReach(packageReachPackageInput)
	},
}

func init() {
	rootCmd.AddCommand(packageReachCmd)

	packageReachCmd.Flags().BoolVar(&packageReachCreatePlot, "createPlot", false, "whether it should create plots for each package")
	packageReachCmd.Flags().StringVar(&packageReachPackageFileInput, "packageInput", "", "input file containing packages")
	packageReachCmd.Flags().StringVar(&packageReachPackageInput, "package", "", "specifiy package to get detailed results for the one")
	packageReachCmd.Flags().StringVar(&packageReachResultPath, "resultPath", "/home/markus/npm-analysis/", "path for single package result")
}

func calculatePackageReach(pkg string) {
	dependenciesTimeline := reach.LoadJSONDependenciesTimeline(packageReachJSONPath)

	dependentsMaps := reach.GenerateDependentsMaps(dependenciesTimeline)

	if pkg == "" {
		startTime := time.Now()

		workerWait := sync.WaitGroup{}

		jobs := make(chan string, 100)

		go streamPackageNamesFromFile(jobs, packageReachPackageFileInput)

		for w := 1; w <= packageReachWorkerNumber; w++ {
			workerWait.Add(1)
			go packageReachWorker(w, jobs, dependentsMaps, &workerWait)
		}

		workerWait.Wait()

		endTime := time.Now()

		log.Printf("Took %v minutes to process all Documents from MongoDB", endTime.Sub(startTime).Minutes())

		reach.CalculateAverageMaintainerReach("averagePackageReach", &packageReachResultMap)

		reach.CalculateMaintainerReachDiff("packageReachDiff", &packageReachResultMap)

		err := reach.CalculatePackageReachDiff(&packageReachResultMap, "packageReachDiffs")
		if err != nil {
			log.Fatal(err)
		}
	} else {
		// calculate for one maintainer the reach of each package per month and overall reach

		log.Printf("Calculating package reach for package %v", pkg)

		var resultMap = make(map[time.Time][]util.PackageReachResult, 0)
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
				resultMap[date] = make([]util.PackageReachResult, 0)
				allPackages := make(map[string]bool, 0)
				packages := make(map[string]reach.ReachDetails, 0)
				reach.PackageReachLayer(pkg, dependentsMaps[date], packages, 1)
				reach.PackageReach(pkg, dependentsMaps[date], allPackages)
				packageKeys := make([]string, len(packages))
				i := 0
				for k, d := range packages {
					if d.Layer > 0 {
						packageKeys[i] = fmt.Sprintf("%02d_%v_%v", d.Layer, k, d.Dependency)
						i++
					}
				}

				sortedDependents := util.StringList(packageKeys)
				sort.Sort(sortedDependents)

				resultMap[date] = append(resultMap[date], util.PackageReachResult{Count: len(packages), Package: pkg, Dependents: sortedDependents})
				log.Printf("Date: %v Count %v", date, len(allPackages))
			}
		}

		for _, p := range resultMap {
			sort.Sort(sort.Reverse(util.PackageReachResultList(p)))
		}

		jsonBytes, err := json.MarshalIndent(resultMap, "", "\t")
		if err != nil {
			log.Fatal(err)
		}

		filePath := path.Join(packageReachResultPath, fmt.Sprintf("%v-reach.json", pkg))
		err = ioutil.WriteFile(filePath, jsonBytes, os.ModePerm)
		if err != nil {
			log.Fatal(err)
		}

		log.Printf("Wrote results to file %v", filePath)
	}

}

func packageReachWorker(workerId int, jobs chan string, dependentsMaps map[time.Time]map[string][]string, workerWait *sync.WaitGroup) {
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
				reach.PackageReach(j, dependentsMaps[date], packages)
				counts = append(counts, float64(len(packages)))
			}
		}

		packageReachResultMap.Store(j, counts)

		if packageReachCreatePlot {
			fileName := plots.GetPlotFileName(j, "package-reach")
			if _, err := os.Stat(fileName); err == nil {
				continue
			}
			plots.GenerateLinePlotForMaintainerReach("package-reach", j, counts, packageReachCreatePlot)
		} else {
			plots.GenerateLinePlotForMaintainerReach("package-reach", j, counts, packageReachCreatePlot)
		}

		//log.Printf("Finished %v", j.Name)
	}
	workerWait.Done()
}
