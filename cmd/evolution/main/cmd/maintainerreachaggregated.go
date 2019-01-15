package cmd

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/markuszm/npm-analysis/database"
	"github.com/markuszm/npm-analysis/util"
	"github.com/mongodb/mongo-go-driver/bson"
	"github.com/pkg/errors"
	"github.com/spf13/cobra"
	"io/ioutil"
	"log"
	"os"
	"path"
	"sort"
	"strings"
	"time"
)

// Commandline arguments
var maintainerReachAggMongoUrl string
var maintainerReachAggOptimal bool
var maintainerReachAggResultPath string
var maintainerReachAggMaintainerRanking string
var maintainerReachAggMaintainerReachResults string
var maintainerReachMonth int
var maintainerReachYear int
var maintainerReachLimit int

var maintainerRankingList []string
var maintainerReachAggResults map[string][]string

var packageReachedMap map[string]bool

var reachTo100Percent []int

var maintainerReachAggCmd = &cobra.Command{
	Use:   "maintainerReachAgg",
	Short: "Aggregates package reach of maintainers and create plot results",
	Long:  `...`,
	Run: func(cmd *cobra.Command, args []string) {
		maintainerReachAggCalculatePackageReach()
	},
}

func init() {
	rootCmd.AddCommand(maintainerReachAggCmd)

	maintainerReachAggCmd.Flags().StringVar(&maintainerReachAggMaintainerRanking, "ranking", "", "input file containing ranked list of maintainers as json")
	maintainerReachAggCmd.Flags().StringVar(&maintainerReachAggMaintainerReachResults, "reachResults", "", "input file containing complete reach results")
	maintainerReachAggCmd.Flags().BoolVar(&maintainerReachAggOptimal, "optimal", false, "whether it should find optimal distribution")
	maintainerReachAggCmd.Flags().StringVar(&maintainerReachAggResultPath, "resultPath", "/home/markus/npm-analysis/maintainerReachAgg", "path for single maintainer result")
	maintainerReachAggCmd.Flags().IntVar(&maintainerReachMonth, "month", 4, "month for date to calculate")
	maintainerReachAggCmd.Flags().IntVar(&maintainerReachYear, "year", 2018, "year for date to calculate")
	maintainerReachAggCmd.Flags().IntVar(&maintainerReachLimit, "limit", 20, "optimal ranking limit")
	maintainerReachAggCmd.Flags().StringVar(&maintainerReachAggMongoUrl, "mongodb", "mongodb://npm:npm123@localhost:27017", "mongo url")
}

func loadMaintainerRanking() error {
	file, err := ioutil.ReadFile(maintainerReachAggMaintainerRanking)
	if err != nil {
		return errors.Wrap(err, "could not read file")
	}
	err = json.Unmarshal(file, &maintainerRankingList)

	return err
}

func loadMaintainerReachResults() error {
	file, err := ioutil.ReadFile(maintainerReachAggMaintainerReachResults)
	if err != nil {
		return errors.Wrap(err, "could not read file")
	}
	err = json.Unmarshal(file, &maintainerReachAggResults)

	return err
}

func maintainerReachAggCalculatePackageReach() {
	if maintainerReachAggOptimal {
		rankMaintainersOptimal()
		return
	}

	mongoDBPackages := database.NewMongoDB(maintainerReachAggMongoUrl, "npm", "maintainerPackages")
	err := mongoDBPackages.Connect()
	if err != nil {
		log.Fatal(err)
	}
	defer mongoDBPackages.Disconnect()

	mongoDBReach := database.NewMongoDB(maintainerReachAggMongoUrl, "npm", "packageReach")
	err = mongoDBReach.Connect()
	if err != nil {
		log.Fatal(err)
	}
	defer mongoDBReach.Disconnect()

	log.Print("Connected to mongodb")

	log.Print("Loading maintainer package data from mongoDB")

	date := time.Date(maintainerReachYear, time.Month(maintainerReachMonth), 1, 0, 0, 0, 0, time.UTC)

	if maintainerReachAggMaintainerRanking == "" {
		rankMaintainers(mongoDBPackages, mongoDBReach, date)
	} else {
		// old way of sorting by top reach
		calculateReachIntersection(mongoDBPackages, mongoDBReach, date)
	}
}

func calculateReachIntersection(mongoDBPackages, mongoDBReach *database.MongoDB, date time.Time) {
	err := loadMaintainerRanking()
	if err != nil {
		log.Fatal(err)
	}
	maintainerIndex := 0

	packageReachedMap = make(map[string]bool)

	for _, maintainer := range maintainerRankingList {
		doc, err := mongoDBPackages.FindOneSimple("key", maintainer)
		if err != nil {
			log.Fatal(err)
		}

		var data StoreMaintainedPackages
		err = json.Unmarshal([]byte(doc), &data)
		if err != nil {
			log.Fatal(err)
		}

		if data.Name == "" {
			continue
		}

		maintainedPackages := data.PackagesTimeline[date]

		allPackages := make(map[string]bool, 0)
		for _, pkg := range maintainedPackages {
			reachDocument, err := mongoDBReach.FindPackageReach(pkg, date)
			if err != nil {
				log.Fatalf("cant find reach for pkg: %v with err: %v", pkg, err)
			}
			reachedPackages := reachDocument.ReachedPackages
			for _, p := range reachedPackages {
				allPackages[p] = true
			}
		}
		newlyAdded := 0

		var packageReachList []string

		for pkg, ok := range allPackages {
			if ok {
				if !packageReachedMap[pkg] {
					newlyAdded++
				}
				packageReachedMap[pkg] = true
				packageReachList = append(packageReachList, pkg)
			}
		}

		if maintainerIndex == 0 {
			reachTo100Percent = append(reachTo100Percent, newlyAdded)
		} else {
			oldCount := reachTo100Percent[maintainerIndex-1]
			reachTo100Percent = append(reachTo100Percent, newlyAdded+oldCount)
		}

		jsonBytes, err := json.MarshalIndent(packageReachList, "", "\t")
		if err != nil {
			log.Fatal(err)
		}

		filePath := GetFilePathForMaintainer(data.Name)
		err = ioutil.WriteFile(filePath, jsonBytes, os.ModePerm)
		if err != nil {
			log.Print(err)
		}

		log.Printf("Wrote results to file %v", filePath)

		maintainerIndex++
	}
	jsonBytes, err := json.Marshal(reachTo100Percent)
	if err != nil {
		log.Fatal(err)
	}
	filePath := path.Join(maintainerReachAggResultPath, "reachTo100Percent.json")
	err = ioutil.WriteFile(filePath, jsonBytes, os.ModePerm)
	if err != nil {
		log.Fatal(err)
	}
}

func rankMaintainers(mongoDBPackages, mongoDBReach *database.MongoDB, date time.Time) {
	cursor, err := mongoDBPackages.ActiveCollection.Find(context.Background(), bson.D{})
	if err != nil {
		log.Fatal(err)
	}
	var results []util.MaintainerReachResult

	reachResultMap := make(map[string][]string)

	idx := 1

	for cursor.Next(context.Background()) {
		doc, err := mongoDBPackages.DecodeValue(cursor)
		if err != nil {
			log.Fatal(err)
		}

		var data StoreMaintainedPackages
		err = json.Unmarshal([]byte(doc.Value), &data)
		if err != nil {
			log.Fatal(err)
		}

		maintainer := data.Name
		if maintainer == "" {
			continue
		}

		maintainedPackages := data.PackagesTimeline[date]

		allPackages := make(map[string]bool, 0)
		for _, pkg := range maintainedPackages {
			reachDocument, err := mongoDBReach.FindPackageReach(pkg, date)
			if err != nil {
				log.Printf("ERROR: cant find reach for maintainer %v pkg: %v with err: %v", maintainer, pkg, err)
			}
			reachedPackages := reachDocument.ReachedPackages
			for _, p := range reachedPackages {
				allPackages[p] = true
			}
		}

		var packageReachList []string

		for pkg, ok := range allPackages {
			if ok {
				packageReachList = append(packageReachList, pkg)
			}
		}

		maintainerReachResult := util.MaintainerReachResult{
			Count: len(packageReachList),
			Name:  maintainer,
		}

		results = append(results, maintainerReachResult)

		reachResultMap[maintainer] = packageReachList

		log.Print(idx)
		idx++
	}

	sort.Sort(sort.Reverse(util.MaintainerReachResultList(results)))
	var maintainers []string
	for _, r := range results {
		maintainers = append(maintainers, r.Name)
	}
	jsonBytes, err := json.MarshalIndent(maintainers, "", "\t")
	if err != nil {
		log.Fatal(err)
	}
	filePath := path.Join(maintainerReachAggResultPath, "maintainerRanking.json")
	err = ioutil.WriteFile(filePath, jsonBytes, os.ModePerm)
	if err != nil {
		log.Fatal(err)
	}
	log.Printf("Wrote results to file %v", filePath)

	jsonBytes, err = json.MarshalIndent(reachResultMap, "", "\t")
	if err != nil {
		log.Fatal(err)
	}
	filePath = path.Join(maintainerReachAggResultPath, "maintainerRankingComplete.json")
	err = ioutil.WriteFile(filePath, jsonBytes, os.ModePerm)
	if err != nil {
		log.Fatal(err)
	}
	log.Printf("Wrote results to file %v", filePath)
}

func rankMaintainersOptimal() {
	err := loadMaintainerRanking()
	if err != nil {
		log.Fatal(err)
	}

	err = loadMaintainerReachResults()
	if err != nil {
		log.Fatal(err)
	}

	startMaintainer := maintainerRankingList[0]

	var optimalRanking []string
	var intersectionCounts []int
	intersectionSet := make(map[string]bool)

	alreadyVisitedMaintainers := make(map[string]bool)

	var previousIntersectionReach = 0

	previousIntersectionReach, intersectionCounts, optimalRanking = addMaintainerToOptimalRanking(startMaintainer, previousIntersectionReach, intersectionCounts, intersectionSet, optimalRanking)

	for {
		var reachDiffs []util.Pair
		// TODO: concurrent if performance is bad
		for m, r := range maintainerReachAggResults {
			if alreadyVisitedMaintainers[m] {
				continue
			}

			diff := 0
			for _, d := range r {
				if !intersectionSet[d] {
					diff++
				}
			}
			reachDiffs = append(reachDiffs, util.Pair{Key: m, Value: diff})
		}

		sort.Sort(sort.Reverse(util.PairList(reachDiffs)))

		nextOptimalMaintainer := reachDiffs[0]
		if nextOptimalMaintainer.Value == 0 || len(optimalRanking) >= maintainerReachLimit {
			break
		}

		log.Printf("adding next optimal maintainer %v with diff %v", nextOptimalMaintainer.Key, nextOptimalMaintainer.Value)

		previousIntersectionReach, intersectionCounts, optimalRanking = addMaintainerToOptimalRanking(nextOptimalMaintainer.Key, previousIntersectionReach, intersectionCounts, intersectionSet, optimalRanking)
		alreadyVisitedMaintainers[nextOptimalMaintainer.Key] = true
	}

	results := map[string]interface{}{
		"OptimalRanking":     optimalRanking,
		"IntersectionCounts": intersectionCounts,
		"MaximumReach":       intersectionCounts[len(optimalRanking)-1],
	}

	jsonBytes, err := json.MarshalIndent(results, "", "\t")
	if err != nil {
		log.Fatal(err)
	}
	filePath := path.Join(maintainerReachAggResultPath, "optimalRanking.json")
	err = ioutil.WriteFile(filePath, jsonBytes, os.ModePerm)
	if err != nil {
		log.Fatal(err)
	}
	log.Printf("Wrote results to file %v", filePath)
}

func addMaintainerToOptimalRanking(maintainer string, previousIntersectionReach int, intersectionCounts []int, intersectionSet map[string]bool, optimalRanking []string) (int, []int, []string) {
	maintainerReach := maintainerReachAggResults[maintainer]
	newlyAdded := 0
	for _, d := range maintainerReach {
		if !intersectionSet[d] {
			intersectionSet[d] = true
			newlyAdded++
		}
	}
	newIntersectionReach := previousIntersectionReach + newlyAdded
	intersectionCounts = append(intersectionCounts, newIntersectionReach)
	optimalRanking = append(optimalRanking, maintainer)

	return newIntersectionReach, intersectionCounts, optimalRanking
}

func GetFilePathForMaintainer(maintainerName string) string {
	nestedDir := GetNestedDirName(maintainerName)
	err := os.MkdirAll(nestedDir, os.ModePerm)
	if err != nil {
		log.Fatalf("Could not create nested directory with %v", err)
	}

	maintainerName = strings.Replace(maintainerName, "/", "", -1)
	maintainerName = strings.Replace(maintainerName, " ", "", -1)
	return fmt.Sprintf("%v/%v-reach.json", GetNestedDirName(maintainerName), maintainerName)
}

func GetNestedDirName(maintainerName string) string {
	return fmt.Sprintf("%v/%v", maintainerReachAggResultPath, string(maintainerName[0]))

}
