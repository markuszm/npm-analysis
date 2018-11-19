package main

import (
	"database/sql"
	"encoding/json"
	"flag"
	"fmt"
	"github.com/markuszm/npm-analysis/database"
	"github.com/markuszm/npm-analysis/database/insert"
	"github.com/markuszm/npm-analysis/evolution"
	"github.com/markuszm/npm-analysis/plots"
	"github.com/markuszm/npm-analysis/util"
	"io/ioutil"
	"log"
	"math"
	"os"
	"path"
	"sort"
	"sync"
	"time"
)

const MYSQL_USER = "root"

const MYSQL_PW = "npm-analysis"

const workerNumber = 100

var db *sql.DB

var insertDB bool

// Stores maintainer count into database and plots average maintainer count and sorted maintainer count

func main() {
	insertDBFlag := flag.Bool("insertdb", false, "specify whether maintainer count should be inserted into db")
	flag.Parse()

	insertDB = *insertDBFlag

	mysqlInitializer := &database.Mysql{}
	mysql, databaseInitErr := mysqlInitializer.InitDB(fmt.Sprintf("%s:%s@/npm?charset=utf8mb4&collation=utf8mb4_bin", MYSQL_USER, MYSQL_PW))
	if databaseInitErr != nil {
		log.Fatal(databaseInitErr)
	}
	defer mysql.Close()

	db = mysql

	changes, err := database.GetMaintainerChanges(mysql)
	if err != nil {
		log.Fatalf("ERROR: loading changes from mysql with %v", err)
	}

	log.Print("Finished retrieving changes from db")

	countMap := evolution.CalculateMaintainerCounts(changes)

	if insertDB {
		err = database.CreateMaintainerCount(db)
		if err != nil {
			log.Fatal(err)
		}

		workerWait := sync.WaitGroup{}

		jobs := make(chan evolution.MaintainerCount, 100)

		for w := 1; w <= workerNumber; w++ {
			workerWait.Add(1)
			go worker(w, jobs, &workerWait)
		}

		for _, maintainerCount := range countMap {
			jobs <- maintainerCount
		}

		close(jobs)
		workerWait.Wait()
	}

	calculateAverageMaintainerCount(countMap)
	plotSortedMaintainerPackageCount(countMap)
}

func calculateAverageMaintainerCount(countMap map[string]evolution.MaintainerCount) {
	maintainerPackageCount := make(map[time.Time]float64, 0)
	maintainerCount := make(map[time.Time]float64, 0)
	averageMaintainerPackageCountPerMonth := make(map[time.Time]float64, 0)

	for _, counts := range countMap {
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
				count := counts.Counts[year][month]
				if count > 0 || isActive {
					maintainerCount[date] = maintainerCount[date] + 1
					maintainerPackageCount[date] = maintainerPackageCount[date] + float64(count)
					isActive = true
				}
				x++
			}
		}
	}

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
			average := maintainerPackageCount[date] / maintainerCount[date]
			averageMaintainerPackageCountPerMonth[date] = average
		}
	}

	var resultList []util.TimeValue

	for k, v := range averageMaintainerPackageCountPerMonth {
		if !math.IsNaN(v) {
			resultList = append(resultList, util.TimeValue{Key: k, Value: v})
		} else {
			resultList = append(resultList, util.TimeValue{Key: k, Value: 0})
		}
	}
	sortedList := util.TimeValueList(resultList)
	sort.Sort(sortedList)

	var avgValues []float64

	for _, v := range sortedList {
		avgValues = append(avgValues, v.Value)
	}

	jsonBytes, err := json.Marshal(sortedList)
	if err != nil {
		log.Fatal(err)
	}

	filePath := path.Join("/home/markus/npm-analysis/", "averageMaintainerPackageCount.json")
	err = ioutil.WriteFile(filePath, jsonBytes, os.ModePerm)
	if err != nil {
		log.Fatal(err)
	}

	plots.GenerateLinePlotForAverageMaintainerPackageCount(avgValues)
}

func plotSortedMaintainerPackageCount(countMap map[string]evolution.MaintainerCount) {
	valuesPerYear := map[int][]float64{
		2010: make([]float64, 0),
		2011: make([]float64, 0),
		2012: make([]float64, 0),
		2013: make([]float64, 0),
		2014: make([]float64, 0),
		2015: make([]float64, 0),
		2016: make([]float64, 0),
		2017: make([]float64, 0),
		2018: make([]float64, 0),
	}

	for _, counts := range countMap {
		//isActive := false
		for year := 2010; year < 2019; year++ {
			count := counts.Counts[year][1]
			if count > 0 {
				vals := valuesPerYear[year]
				vals = append(vals, math.Log10(float64(count)))
				valuesPerYear[year] = vals
				//isActive = true
			}
		}
	}

	for y, values := range valuesPerYear {
		sortedList := util.FloatList(values)
		sort.Sort(sort.Reverse(sortedList))
		valuesPerYear[y] = sortedList
	}

	err := writeSortedMaintainerPackageCount(valuesPerYear, "/home/markus/npm-analysis/sortedMaintainerPackageCount.json")
	if err != nil {
		log.Fatal(err)
	}

	plots.GenerateSortedLinePlotMaintainerPackageCount(valuesPerYear)
}

func writeSortedMaintainerPackageCount(valuesPerYear map[int][]float64, filePath string) error {
	bytes, err := json.Marshal(valuesPerYear)
	if err != nil {
		return err
	}

	err = ioutil.WriteFile(filePath, bytes, os.ModePerm)
	return err
}

func worker(id int, jobs chan evolution.MaintainerCount, workerWait *sync.WaitGroup) {
	for m := range jobs {
		err := insert.StoreMaintainerCount(db, m)
		if err != nil {
			log.Fatalf("ERROR: writing to database with %v", err)
		}
	}
	workerWait.Done()
}
