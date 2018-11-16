package main

import (
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"github.com/markuszm/npm-analysis/database"
	reach "github.com/markuszm/npm-analysis/evolution/maintainerreach"
	"github.com/markuszm/npm-analysis/plots"
	"github.com/markuszm/npm-analysis/util"
	"github.com/mongodb/mongo-go-driver/bson"
	"io/ioutil"
	"log"
	"os"
	"path"
	"sort"
	"sync"
	"time"
)

const MONGOURL = "mongodb://npm:npm123@localhost:27017"

const JSONPATH = "./db-data/dependenciesTimeline.json"

const workerNumber = 100

// calculates Package reach of a maintainer and plots it
func main() {
	createPlot = flag.Bool("createPlot", false, "whether it should create plots for each maintainer")
	generateData = flag.Bool("generateData", false, "whether it should generate intermediate map for performance")
	maintainer := flag.String("maintainer", "", "specifiy maintainer to get detailed results for the one")
	resultPath = flag.String("resultPath", "/home/markus/npm-analysis/", "path for single maintainer result")
	flag.Parse()

	if *generateData {
		reach.GenerateTimeLatestVersionMap(MONGOURL, JSONPATH)
	}

	calculatePackageReach(*maintainer)
}

type StoreMaintainedPackages struct {
	Name             string                 `json:"name"`
	PackagesTimeline map[time.Time][]string `json:"packages"`
}

var resultMap = sync.Map{}

var createPlot *bool
var generateData *bool

var resultPath *string

func calculatePackageReach(maintainer string) {
	dependenciesTimeline := reach.LoadJSONDependenciesTimeline(JSONPATH)

	dependentsMaps := reach.GenerateDependentsMaps(dependenciesTimeline)

	mongoDB := database.NewMongoDB(MONGOURL, "npm", "maintainerPackages")
	mongoDB.Connect()
	defer mongoDB.Disconnect()

	log.Print("Connected to mongodb")

	if maintainer == "" {
		startTime := time.Now()

		workerWait := sync.WaitGroup{}

		jobs := make(chan StoreMaintainedPackages, 100)

		for w := 1; w <= workerNumber; w++ {
			workerWait.Add(1)
			go worker(w, jobs, dependentsMaps, &workerWait)
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

		reach.CalculateAverageMaintainerReach(&resultMap)

		reach.CalculateMaintainerReachDiff(&resultMap)
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

		filePath := path.Join(*resultPath, fmt.Sprintf("%v-reach.json", maintainer))
		err = ioutil.WriteFile(filePath, jsonBytes, os.ModePerm)
		if err != nil {
			log.Fatal(err)
		}

		log.Printf("Wrote results to file %v", filePath)
	}

}

func worker(workerId int, jobs chan StoreMaintainedPackages, dependentsMaps map[time.Time]map[string][]string, workerWait *sync.WaitGroup) {
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

		resultMap.Store(j.Name, counts)

		if *createPlot {
			fileName := plots.GetPlotFileName(j.Name, "maintainer-reach")
			if _, err := os.Stat(fileName); err == nil {
				continue
			}
			plots.GenerateLinePlotForMaintainerReach(j.Name, counts)
		}

		//log.Printf("Finished %v", j.Name)
	}
	workerWait.Done()
}
