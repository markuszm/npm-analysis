package cmd

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/markuszm/npm-analysis/database"
	reach "github.com/markuszm/npm-analysis/evolution/maintainerreach"
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

const maintainerReachAggMongoUrl = "mongodb://npm:npm123@localhost:27017"

const maintainerReachAggJsonPath = "./db-data/dependenciesTimeline.json"

var packageReachedMap map[string]bool

var reachTo100Percent []int

var maintainerReachAggGenerateData bool

var maintainerReachAggResultPath string

var maintainerReachAggMaintainerRanking string

var maintainerReachAggPackageInput string

var maintainerRankingList []string

var maintainerReachAggCmd = &cobra.Command{
	Use:   "maintainerReachAgg",
	Short: "Aggregates package reach of maintainers and create plot results",
	Long:  `...`,
	Run: func(cmd *cobra.Command, args []string) {
		if maintainerReachAggGenerateData {
			reach.GenerateTimeLatestVersionMap(maintainerReachAggMongoUrl, maintainerReachAggJsonPath)
		}

		err := loadPackagesToReachedMap(maintainerReachAggPackageInput)
		if err != nil {
			log.Fatal(err)
		}

		maintainerReachAggCalculatePackageReach()
	},
}

func init() {
	rootCmd.AddCommand(maintainerReachAggCmd)

	maintainerReachAggCmd.Flags().StringVar(&maintainerReachAggPackageInput, "packageInput", "", "input file containing packages")
	maintainerReachAggCmd.Flags().StringVar(&maintainerReachAggMaintainerRanking, "maintainerInput", "", "input file containing ranked list of maintainers as json")
	maintainerReachAggCmd.Flags().BoolVar(&maintainerReachAggGenerateData, "generateData", false, "whether it should generate intermediate map for performance")
	maintainerReachAggCmd.Flags().StringVar(&maintainerReachAggResultPath, "resultPath", "/home/markus/npm-analysis/maintainerReachAgg", "path for single maintainer result")
}

func loadPackagesToReachedMap(packagesInput string) error {
	file, err := ioutil.ReadFile(packagesInput)
	if err != nil {
		return errors.Wrap(err, "could not read file")
	}

	var packages []string
	json.Unmarshal(file, &packages)

	packageReachedMap = make(map[string]bool, 0)
	for _, p := range packages {
		if p == "" {
			continue
		}
		packageReachedMap[p] = false
	}

	return nil
}

func loadMaintainerRanking() error {
	file, err := ioutil.ReadFile(maintainerReachAggMaintainerRanking)
	if err != nil {
		return errors.Wrap(err, "could not read file")
	}
	json.Unmarshal(file, &maintainerRankingList)

	return nil
}

func maintainerReachAggCalculatePackageReach() {
	dependenciesTimeline := reach.LoadJSONDependenciesTimeline(maintainerReachAggJsonPath)

	dependentsMaps := reach.GenerateDependentsMaps(dependenciesTimeline)

	mongoDB := database.NewMongoDB(maintainerReachAggMongoUrl, "npm", "maintainerPackages")
	mongoDB.Connect()
	defer mongoDB.Disconnect()

	log.Print("Connected to mongodb")

	log.Print("Loading maintainer package data from mongoDB")

	maintainerIndex := 0

	if maintainerReachAggMaintainerRanking == "" {
		cursor, err := mongoDB.ActiveCollection.Find(context.Background(), bson.D{})
		if err != nil {
			log.Fatal(err)
		}

		var results []util.MaintainerReachResult

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

			lastYear := 2018
			lastMonth := 4
			date := time.Date(lastYear, time.Month(lastMonth), 1, 0, 0, 0, 0, time.UTC)

			maintainedPackages := data.PackagesTimeline[date]

			allPackages := make(map[string]bool, 0)
			for _, pkg := range maintainedPackages {
				reach.PackageReach(pkg, dependentsMaps[date], allPackages)
			}
			var packageReachList []string

			for pkg, ok := range allPackages {
				if ok {
					packageReachList = append(packageReachList, pkg)
				}
			}

			maintainerReachResult := util.MaintainerReachResult{
				Count:      len(packageReachList),
				Name:       data.Name,
				Packages:   maintainedPackages,
				Dependents: packageReachList,
			}

			results = append(results, maintainerReachResult)
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

		jsonBytes, err = json.MarshalIndent(results, "", "\t")
		if err != nil {
			log.Fatal(err)
		}

		filePath = path.Join(maintainerReachAggResultPath, "maintainerRankingComplete.json")
		err = ioutil.WriteFile(filePath, jsonBytes, os.ModePerm)
		if err != nil {
			log.Fatal(err)
		}

		log.Printf("Wrote results to file %v", filePath)
	} else {
		err := loadMaintainerRanking()
		if err != nil {
			log.Fatal(err)
		}

		for _, maintainer := range maintainerRankingList {
			doc, err := mongoDB.FindOneSimple("key", maintainer)
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

			lastYear := 2018
			lastMonth := 4
			date := time.Date(lastYear, time.Month(lastMonth), 1, 0, 0, 0, 0, time.UTC)

			maintainedPackages := data.PackagesTimeline[date]

			allPackages := make(map[string]bool, 0)
			for _, pkg := range maintainedPackages {
				reach.PackageReach(pkg, dependentsMaps[date], allPackages)
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
