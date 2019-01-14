package cmd

import (
	"context"
	"encoding/json"
	"github.com/markuszm/npm-analysis/database"
	reach "github.com/markuszm/npm-analysis/evolution/maintainerreach"
	"github.com/markuszm/npm-analysis/model"
	"github.com/mongodb/mongo-go-driver/bson"
	"github.com/spf13/cobra"
	"io/ioutil"
	"log"
	"os"
	"path"
	"time"
)

const MongoUrl = "mongodb://npm:npm123@localhost:27017"

const OutputPath = "/home/markus/npm-analysis/averageDepsNew.json"

const dependenciesTimelinePath = "./db-data/dependenciesTimeline.json"

var averageDepsCmd = &cobra.Command{
	Use:   "averageDeps",
	Short: "Averages dependencies",
	Long:  `...`,
	Run: func(cmd *cobra.Command, args []string) {
		outCountMap := make(map[time.Time]int, 0)
		inCountMap := make(map[time.Time]int, 0)
		packageCountMap := make(map[time.Time]int, 0)

		err := collectData(outCountMap, inCountMap, packageCountMap)
		if err != nil {
			log.Fatalf("error while collecting data with %v", err)
		}

		averagesInMap, averagesOutMap := calculateAverages(outCountMap, inCountMap, packageCountMap)

		err = writeData(outCountMap, packageCountMap, averagesInMap, averagesOutMap)
		if err != nil {
			log.Fatalf("error writing results with %v", err)
		}
	},
}

func init() {
	rootCmd.AddCommand(averageDepsCmd)
}

func calculateAverages(outCountMap, inCountMap, packageCountMap map[time.Time]int) (map[time.Time]float64, map[time.Time]float64) {
	averagesInMap := make(map[time.Time]float64, 0)
	averagesOutMap := make(map[time.Time]float64, 0)
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
			inOverallCount := inCountMap[date]
			outOverallCount := outCountMap[date]
			packageOverallCount := packageCountMap[date]

			averageIn := float64(inOverallCount) / float64(packageOverallCount)
			averagesInMap[date] = averageIn

			averageOut := float64(outOverallCount) / float64(packageOverallCount)
			averagesOutMap[date] = averageOut
		}
	}
	return averagesInMap, averagesOutMap
}

func writeData(depCountMap map[time.Time]int, packageCountMap map[time.Time]int, averagesInMap, averagesOutMap map[time.Time]float64) error {
	bytes, err := json.Marshal(map[string]interface{}{
		"depCounts":    depCountMap,
		"packageCount": packageCountMap,
		"averagesIn":   averagesInMap,
		"averagesOut":  averagesOutMap,
	})
	if err != nil {
		return err
	}
	return ioutil.WriteFile(OutputPath, bytes, os.ModePerm)
}

func collectData(outCountMap, inCountMap, packageCountMap map[time.Time]int) error {
	mongoDB := database.NewMongoDB(MongoUrl, "npm", "timeline")
	mongoDB.Connect()
	defer mongoDB.Disconnect()
	startTime := time.Now()
	cursor, err := mongoDB.ActiveCollection.Find(context.Background(), bson.D{})
	if err != nil {
		log.Fatal(err)
	}
	i := 0

	packageTimemap := make(map[time.Time][]string)

	dependenciesTimeline := reach.LoadJSONDependenciesTimeline(dependenciesTimelinePath)

	dependentsMaps := reach.GenerateDependentsMaps(dependenciesTimeline)

	for cursor.Next(context.Background()) {
		doc, err := mongoDB.DecodeValue(cursor)
		if err != nil {
			log.Fatal(err)
		}

		var timeMap map[time.Time]model.SlimPackageData

		err = json.Unmarshal([]byte(doc.Value), &timeMap)
		if err != nil {
			log.Fatal(err)
		}

		for t, data := range timeMap {
			if data.Version == "unreleased" {
				continue
			}
			if packageTimemap[t] == nil {
				packageTimemap[t] = []string{doc.Key}
			} else {
				packageTimemap[t] = append(packageTimemap[t], doc.Key)
			}
			packageCountMap[t]++

			if data.Dependencies != nil {
				outCountMap[t] += len(data.Dependencies)
			}

			dependents := dependentsMaps[t][doc.Key]
			inCountMap[t] += len(dependents)
		}

		if i%10000 == 0 {
			log.Printf("Finished %v packages", i)
		}
		i++
	}
	cursor.Close(context.Background())
	endTime := time.Now()
	log.Printf("Took %v minutes to process all Documents from MongoDB", endTime.Sub(startTime).Minutes())

	jsonBytes, err := json.Marshal(packageTimemap)
	if err != nil {
		log.Fatal(err)
	}

	filePath := path.Join("/home/markus/npm-analysis/", "packageTimemap.json")
	err = ioutil.WriteFile(filePath, jsonBytes, os.ModePerm)
	if err != nil {
		log.Fatal(err)
	}

	return nil
}
