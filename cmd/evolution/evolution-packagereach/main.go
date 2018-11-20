package main

import (
	"encoding/json"
	"flag"
	"fmt"
	reach "github.com/markuszm/npm-analysis/evolution/maintainerreach"
	"github.com/markuszm/npm-analysis/plots"
	"github.com/markuszm/npm-analysis/util"
	"go.uber.org/zap"
	"io/ioutil"
	"log"
	"os"
	"path"
	"sort"
	"strings"
	"sync"
	"time"
)

const MYSQL_USER = "root"

const MYSQL_PW = "npm-analysis"
const JSONPATH = "./db-data/dependenciesTimeline.json"

const workerNumber = 100

var resultMap = sync.Map{}

var createPlot *bool

var resultPath *string

var packagesInput *string

var pkg *string

var logger *zap.SugaredLogger

// calculates Package reach of a packages and plots it
func main() {
	createPlot = flag.Bool("createPlot", false, "whether it should create plots for each package")
	packagesInput = flag.String("packageInput", "", "input file containing packages")
	pkg = flag.String("package", "", "specifiy package to get detailed results for the one")
	resultPath = flag.String("resultPath", "/home/markus/npm-analysis/", "path for single package result")
	flag.Parse()

	calculatePackageReach(*pkg)
}

func initializeLogger() {
	// Initialize logger for all commands
	cfg := zap.NewDevelopmentConfig()
	cfg.DisableStacktrace = true
	l, _ := cfg.Build()
	logger = l.Sugar()
}

func streamPackageNamesFromFile(packageChan chan string) {
	if strings.HasSuffix(*packagesInput, ".json") {
		file, err := ioutil.ReadFile(*packagesInput)
		if err != nil {
			logger.Fatalw("could not read file", "err", err)
		}

		var packages []string
		json.Unmarshal(file, &packages)

		for _, p := range packages {
			if p == "" {
				continue
			}
			packageChan <- p
		}
	} else {
		file, err := ioutil.ReadFile(*packagesInput)
		if err != nil {
			logger.Fatalw("could not read file", "err", err)
		}
		lines := strings.Split(string(file), "\n")
		for _, l := range lines {
			if l == "" {
				continue
			}
			packageChan <- l
		}
	}

	close(packageChan)
}

func calculatePackageReach(pkg string) {
	dependenciesTimeline := reach.LoadJSONDependenciesTimeline(JSONPATH)

	dependentsMaps := reach.GenerateDependentsMaps(dependenciesTimeline)

	if pkg == "" {
		startTime := time.Now()

		workerWait := sync.WaitGroup{}

		jobs := make(chan string, 100)

		go streamPackageNamesFromFile(jobs)

		for w := 1; w <= workerNumber; w++ {
			workerWait.Add(1)
			go worker(w, jobs, dependentsMaps, &workerWait)
		}

		workerWait.Wait()

		endTime := time.Now()

		log.Printf("Took %v minutes to process all Documents from MongoDB", endTime.Sub(startTime).Minutes())

		reach.CalculateAverageMaintainerReach("averagePackageReach", &resultMap)

		reach.CalculateMaintainerReachDiff("packageReachDiff", &resultMap)

		err := reach.CalculatePackageReachDiff(&resultMap)
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

		filePath := path.Join(*resultPath, fmt.Sprintf("%v-reach.json", pkg))
		err = ioutil.WriteFile(filePath, jsonBytes, os.ModePerm)
		if err != nil {
			log.Fatal(err)
		}

		log.Printf("Wrote results to file %v", filePath)
	}

}

func worker(workerId int, jobs chan string, dependentsMaps map[time.Time]map[string][]string, workerWait *sync.WaitGroup) {
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

		resultMap.Store(j, counts)

		if *createPlot {
			fileName := plots.GetPlotFileName(j, "package-reach")
			if _, err := os.Stat(fileName); err == nil {
				continue
			}
			plots.GenerateLinePlotForMaintainerReach("package-reach", j, counts, *createPlot)
		} else {
			plots.GenerateLinePlotForMaintainerReach("package-reach", j, counts, *createPlot)
		}

		//log.Printf("Finished %v", j.Name)
	}
	workerWait.Done()
}
