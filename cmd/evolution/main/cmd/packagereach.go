package cmd

import (
	"encoding/json"
	"fmt"
	"github.com/markuszm/npm-analysis/database"
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
var packageReachCreatePerPackageData bool

var packageReachResultPath string

var packageReachPackageFileInput string

var packageReachPackageInput string

var packageReachMongoDB string

var packageReachIsMongoInsert bool
var packageReachIsRanking bool

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

	packageReachCmd.Flags().StringVar(&packageReachMongoDB, "mongodb", "mongodb://npm:npm123@localhost:27017", "mongodb url")
	packageReachCmd.Flags().BoolVar(&packageReachCreatePlot, "createPlot", false, "whether it should create plots for each package")
	packageReachCmd.Flags().BoolVar(&packageReachCreatePerPackageData, "createPerPackage", false, "whether it should create data files per package")
	packageReachCmd.Flags().BoolVar(&packageReachIsMongoInsert, "mongoInsert", false, "whether it should insert package reach into mongo")
	packageReachCmd.Flags().BoolVar(&packageReachIsRanking, "ranking", false, "ranks packages by reach")
	packageReachCmd.Flags().StringVar(&packageReachPackageFileInput, "packageInput", "", "input file containing packages")
	packageReachCmd.Flags().StringVar(&packageReachPackageInput, "package", "", "specifiy package to get detailed results for the one")
	packageReachCmd.Flags().StringVar(&packageReachResultPath, "resultPath", "./output/packageReach", "path for single package result")
}

func calculatePackageReach(pkg string) {
	if packageReachIsRanking {
		mongoDBReach := database.NewMongoDB(packageReachMongoDB, "npm", "packageReach")
		err := mongoDBReach.Connect()
		if err != nil {
			log.Fatal(err)
		}
		defer mongoDBReach.Disconnect()

		date := time.Date(2018, time.Month(4), 1, 0, 0, 0, 0, time.UTC)

		rankPackages(mongoDBReach, date)

		return
	}

	dependenciesTimeline := reach.LoadJSONDependenciesTimeline(packageReachJSONPath)

	dependentsMaps := reach.GenerateDependentsMaps(dependenciesTimeline)

	if pkg == "" {
		packageReachAll(dependentsMaps)
	} else {
		// calculate for one package the reach of each package per month and overall reach
		packageReachOne(pkg, dependentsMaps)
	}

}

func rankPackages(mongoDBReach *database.MongoDB, date time.Time) {
	packages := getPackageNamesFromFile(packageReachPackageFileInput)

	var results []util.PackageReachResult

	for _, pkg := range packages {
		reachDocument, err := mongoDBReach.FindPackageReach(pkg, date)
		if err != nil {
			log.Printf("ERROR: cant find reach for pkg: %v with err: %v", pkg, err)
		}
		reachedPackages := reachDocument.ReachedPackages

		results = append(results, util.PackageReachResult{
			Count:      len(reachedPackages),
			Package:    pkg,
			Dependents: nil,
		})

		log.Printf("Processed %v", pkg)
	}

	sort.Sort(sort.Reverse(util.PackageReachResultList(results)))
	var ranking []string
	for _, r := range results {
		ranking = append(ranking, r.Package)
	}
	jsonBytes, err := json.MarshalIndent(ranking, "", "\t")
	if err != nil {
		log.Fatal(err)
	}
	filePath := path.Join(packageReachResultPath, "packagesRanking.json")
	err = ioutil.WriteFile(filePath, jsonBytes, os.ModePerm)
	if err != nil {
		log.Fatal(err)
	}
	log.Printf("Wrote results to file %v", filePath)
}

func packageReachAll(dependentsMaps map[time.Time]map[string][]string) {
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
	reach.CalculateAverageResults("averagePackageReach", packageReachResultPath, &packageReachResultMap)
	reach.CalculateMaintainerReachDiff("packageReachDiff", packageReachResultPath, &packageReachResultMap)
	err := reach.CalculatePackageReachDiffs(&packageReachResultMap, "packageReachDiffs", packageReachResultPath)
	if err != nil {
		log.Fatal(err)
	}
}

func packageReachOne(pkg string, dependentsMaps map[time.Time]map[string][]string) {
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

func packageReachWorker(workerId int, jobs chan string, dependentsMaps map[time.Time]map[string][]string, workerWait *sync.WaitGroup) {
	mongoDB := database.NewMongoDB(maintainerReachMongoUrl, "npm", "packageReach")
	mongoDB.Connect()
	defer mongoDB.Disconnect()

	for pkg := range jobs {
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
				reach.PackageReach(pkg, dependentsMaps[date], packages)

				reachedPackages := make([]string, 0)

				for dependent, ok := range packages {
					if ok {
						reachedPackages = append(reachedPackages, dependent)
					}
				}

				if packageReachIsMongoInsert {
					err := mongoDB.InsertPackageReach(pkg, date, reachedPackages)
					if err != nil {
						log.Fatalf("ERROR: cannot insert package reach into mongo with error: %v", err)
					}
				}

				counts = append(counts, float64(len(packages)))
			}
		}

		packageReachResultMap.Store(pkg, counts)

		if packageReachCreatePlot {
			fileName := plots.GetPlotFileName(pkg, "package-reach")
			if _, err := os.Stat(fileName); err == nil {
				continue
			}
			plots.GenerateLinePlotForMaintainerReach("package-reach", pkg, counts, packageReachCreatePlot)
		} else {
			if packageReachCreatePerPackageData {
				plots.GenerateLinePlotForMaintainerReach("package-reach", pkg, counts, packageReachCreatePlot)
			}
		}

		//log.Printf("Finished %v", pkg.Name)
	}
	workerWait.Done()
}
