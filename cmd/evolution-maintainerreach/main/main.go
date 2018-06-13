package main

import (
	"context"
	"encoding/json"
	"github.com/markuszm/npm-analysis/database"
	"github.com/markuszm/npm-analysis/plots"
	"github.com/markuszm/npm-analysis/util"
	"github.com/mongodb/mongo-go-driver/bson"
	"io/ioutil"
	"log"
	"math"
	"os"
	"sort"
	"sync"
	"time"
)

const MONGOURL = "mongodb://npm:npm123@localhost:27017"

const JSONPATH = "./db-data/dependenciesTimeline.json"

const workerNumber = 100

func main() {
	calculatePackageReach()
}

type StoreMaintainedPackages struct {
	Name             string                 `json:"name"`
	PackagesTimeline map[time.Time][]string `json:"packages"`
}

var resultMap = sync.Map{}

const createPlot = false

func calculatePackageReach() {
	dependenciesTimeline := loadJsonDependenciesTimeline()

	dependentsMaps := generateDependentsMaps(dependenciesTimeline)

	mongoDB := database.NewMongoDB(MONGOURL, "npm", "maintainerPackages")
	mongoDB.Connect()
	defer mongoDB.Disconnect()

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

	calculateAverageMaintainerReach()
}

func calculateAverageMaintainerReach() {
	maintainerReachCount := make(map[time.Time]float64, 0)
	maintainerCount := make(map[time.Time]float64, 0)
	averageMaintainerReachPerMonth := make(map[time.Time]float64, 0)

	resultMap.Range(func(key, value interface{}) bool {
		counts := value.([]float64)
		x := 0
		isActive := false
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
				if counts[x] > 0 || isActive {
					maintainerCount[date] = maintainerCount[date] + 1
					maintainerReachCount[date] = maintainerReachCount[date] + counts[x]
					isActive = true
				}
				x++
			}
		}
		return true
	})

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
			average := maintainerReachCount[date] / maintainerCount[date]
			averageMaintainerReachPerMonth[date] = average
		}
	}

	var resultList []util.TimeValue

	for k, v := range averageMaintainerReachPerMonth {
		if !math.IsNaN(v) {
			resultList = append(resultList, util.TimeValue{Key: k, Value: v})
		}
	}
	sortedList := util.TimeValueList(resultList)
	sort.Sort(sortedList)

	var avgValues []float64

	for _, v := range sortedList {
		avgValues = append(avgValues, v.Value)
	}

	log.Print(avgValues, averageMaintainerReachPerMonth)

	plots.GenerateLinePlotForAverageMaintainerReach(avgValues)
}

func generateDependentsMaps(dependenciesTimeline map[time.Time]map[string]map[string]bool) map[time.Time]map[string][]string {
	dependentsMaps := make(map[time.Time]map[string][]string, 0)
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
			packageMap := dependenciesTimeline[date]
			dependentsMap := GenerateDependentMap(packageMap)
			dependentsMaps[date] = dependentsMap
		}
	}
	return dependentsMaps
}

func worker(workerId int, jobs chan StoreMaintainedPackages, dependentsMaps map[time.Time]map[string][]string, workerWait *sync.WaitGroup) {
	for j := range jobs {
		if createPlot {
			fileName := plots.GetPlotFileName(j.Name, "maintainer-reach")
			if _, err := os.Stat(fileName); err == nil {
				continue
			}
		}
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
					PackageReach(pkg, dependentsMaps[date], packages)
				}

				counts = append(counts, float64(len(packages)))
			}
		}

		resultMap.Store(j.Name, counts)

		if createPlot {
			plots.GenerateLinePlotForMaintainerReach(j.Name, counts)
		}

		log.Printf("Finished %v", j.Name)
	}
	workerWait.Done()
}

func loadJsonDependenciesTimeline() map[time.Time]map[string]map[string]bool {
	log.Print("Loading json")
	bytes, err := ioutil.ReadFile(JSONPATH)
	if err != nil {
		log.Fatal(err)
	}
	var dependenciesTimeline map[time.Time]map[string]map[string]bool
	err = json.Unmarshal(bytes, &dependenciesTimeline)
	if err != nil {
		log.Fatal(err)
	}
	log.Print("Finished loading json")
	return dependenciesTimeline
}

func GenerateDependentMap(packageMap map[string]map[string]bool) map[string][]string {
	dependentsMap := make(map[string][]string, 0)
	for dependent, deps := range packageMap {
		for dep, _ := range deps {
			dependentsList := dependentsMap[dep]
			if dependentsList == nil {
				dependentsList = make([]string, 0)
			}
			dependentsList = append(dependentsList, dependent)
			dependentsMap[dep] = dependentsList
		}
	}
	return dependentsMap
}

func PackageReach(pkg string, dependentsMap map[string][]string, packages map[string]bool) {
	for _, dependent := range dependentsMap[pkg] {
		if ok := packages[dependent]; !ok {
			packages[dependent] = true
			PackageReach(dependent, dependentsMap, packages)
		}
	}
}

func generateTimeLatestVersionMap() {
	dependenciesTimeline := make(map[time.Time]map[string]map[string]bool, 0)

	mongoDB := database.NewMongoDB(MONGOURL, "npm", "timeline")

	mongoDB.Connect()
	defer mongoDB.Disconnect()

	startTime := time.Now()

	cursor, err := mongoDB.ActiveCollection.Find(context.Background(), bson.NewDocument())
	if err != nil {
		log.Fatal(err)
	}
	i := 0
	for cursor.Next(context.Background()) {
		doc, err := mongoDB.DecodeValue(cursor)
		if err != nil {
			log.Fatal(err)
		}

		var timeMap map[time.Time]SlimPackageData

		err = json.Unmarshal([]byte(doc.Value), &timeMap)
		if err != nil {
			log.Fatal(err)
		}

		for t, pkg := range timeMap {
			if dependenciesTimeline[t] == nil {
				dependenciesTimeline[t] = make(map[string]map[string]bool, 0)
			}
			if len(pkg.Dependencies) > 0 {
				dependencies := make(map[string]bool, 0)
				for _, dep := range pkg.Dependencies {
					dependencies[dep] = true
				}
				dependenciesTimeline[t][doc.Key] = dependencies
			}
		}

		if i%10000 == 0 {
			log.Printf("Finished %v packages", i)
		}
		i++
	}
	cursor.Close(context.Background())
	endTime := time.Now()
	log.Printf("Took %v minutes to process all Documents from MongoDB", endTime.Sub(startTime).Minutes())
	bytes, err := json.Marshal(dependenciesTimeline)
	if err != nil {
		log.Fatal(err)
	}
	log.Print("Finished transforming to JSON")
	ioutil.WriteFile("./db-data/dependenciesTimeline.json", bytes, os.ModePerm)
}

type SlimPackageData struct {
	Dependencies []string `json:"dependencies"`
}
