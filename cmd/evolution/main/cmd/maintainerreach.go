package cmd

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/markuszm/npm-analysis/database"
	reach "github.com/markuszm/npm-analysis/evolution/maintainerreach"
	"github.com/markuszm/npm-analysis/plots"
	"github.com/markuszm/npm-analysis/util"
	"github.com/mongodb/mongo-go-driver/bson"
	"github.com/spf13/cobra"
	"io/ioutil"
	"log"
	"os"
	"path"
	"sort"
	"sync"
	"time"
)

const maintainerReachMongoUrl = "mongodb://npm:npm123@localhost:27017"

const maintainerReachJsonPath = "./db-data/dependenciesTimeline.json"

const maintainerReachWorkerNumber = 100

var maintainerReachResultMap = sync.Map{}

var maintainerReachCreatePlot bool
var maintainerReachGenerateData bool

var maintainerReachResultPath string
var maintainerReachMaintainer string

// calculates Package reach of a maintainer and plots it
var maintainerReachCmd = &cobra.Command{
	Use:   "maintainerReach",
	Short: "Calculates Package reach of a maintainer and plots it",
	Long:  `...`,
	Run: func(cmd *cobra.Command, args []string) {
		if maintainerReachGenerateData {
			reach.GenerateTimeLatestVersionMap(maintainerReachMongoUrl, maintainerReachJsonPath)
		}

		maintainerReachCalculatePackageReach(maintainerReachMaintainer)
	},
}

func init() {
	rootCmd.AddCommand(maintainerReachCmd)

	maintainerReachCmd.Flags().BoolVar(&maintainerReachCreatePlot, "createPlot", false, "whether it should create plots for each maintainer")
	maintainerReachCmd.Flags().BoolVar(&maintainerReachGenerateData, "generateData", false, "whether it should generate intermediate map for performance")
	maintainerReachCmd.Flags().StringVar(&maintainerReachMaintainer, "maintainer", "", "specifiy maintainer to get detailed results for the one")
	maintainerReachCmd.Flags().StringVar(&maintainerReachResultPath, "resultPath", "/home/markus/npm-analysis/", "path for single maintainer result")
}

func maintainerReachCalculatePackageReach(maintainer string) {
	dependenciesTimeline := reach.LoadJSONDependenciesTimeline(maintainerReachJsonPath)

	dependentsMaps := reach.GenerateDependentsMaps(dependenciesTimeline)

	mongoDB := database.NewMongoDB(maintainerReachMongoUrl, "npm", "maintainerPackages")
	mongoDB.Connect()
	defer mongoDB.Disconnect()

	log.Print("Connected to mongodb")

	if maintainer == "" {
		startTime := time.Now()

		workerWait := sync.WaitGroup{}

		jobs := make(chan StoreMaintainedPackages, 100)

		for w := 1; w <= maintainerReachWorkerNumber; w++ {
			workerWait.Add(1)
			go maintainerReachWorker(w, jobs, dependentsMaps, &workerWait)
		}

		log.Print("Loading maintainer package data from mongoDB")

		cursor, err := mongoDB.ActiveCollection.Find(context.Background(), bson.NewDocument())
		if err != nil {
			log.Fatal(err)
		}
		for cursor.Next(context.Background()) {
			doc, err := mongoDB.DecodeValue(cursor)
			if err != nil {
				log.Fatal(err)
			}

			var data StoreMaintainedPackages
			err = json.Unmarshal([]byte(doc.Value), &data)
			if err != nil {
				log.Fatal(err)
			}

			if data.Name == "" {
				continue
			}

			jobs <- data
		}

		close(jobs)

		workerWait.Wait()

		endTime := time.Now()

		log.Printf("Took %v minutes to process all Documents from MongoDB", endTime.Sub(startTime).Minutes())

		reach.CalculateAverageMaintainerReach("averageMaintainerReach", &maintainerReachResultMap)

		reach.CalculateMaintainerReachDiff("maintainerReachDiff", &maintainerReachResultMap)
	} else {
		// calculate for one maintainer the reach of each package per month and overall reach

		val, err := mongoDB.FindOneSimple("key", maintainer)
		if err != nil {
			log.Fatal(err)
		}

		var data StoreMaintainedPackages
		err = json.Unmarshal([]byte(val), &data)
		if err != nil {
			log.Fatal(err)
		}

		log.Printf("Calculating package reach for each package of maintainer %v", maintainer)

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
				for _, pkg := range data.PackagesTimeline[date] {
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
				}
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

		filePath := path.Join(maintainerReachResultPath, fmt.Sprintf("%v-reach.json", maintainer))
		err = ioutil.WriteFile(filePath, jsonBytes, os.ModePerm)
		if err != nil {
			log.Fatal(err)
		}

		log.Printf("Wrote results to file %v", filePath)
	}

}

func maintainerReachWorker(workerId int, jobs chan StoreMaintainedPackages, dependentsMaps map[time.Time]map[string][]string, workerWait *sync.WaitGroup) {
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
				for _, pkg := range j.PackagesTimeline[date] {
					reach.PackageReach(pkg, dependentsMaps[date], packages)
				}

				counts = append(counts, float64(len(packages)))
			}
		}

		maintainerReachResultMap.Store(j.Name, counts)

		if maintainerReachCreatePlot {
			fileName := plots.GetPlotFileName(j.Name, "maintainer-reach")
			if _, err := os.Stat(fileName); err == nil {
				continue
			}
			plots.GenerateLinePlotForMaintainerReach("maintainer-reach", j.Name, counts, maintainerReachCreatePlot)
		} else {
			plots.GenerateLinePlotForMaintainerReach("maintainer-reach", j.Name, counts, maintainerReachCreatePlot)
		}

		//log.Printf("Finished %v", j.Name)
	}
	workerWait.Done()
}
