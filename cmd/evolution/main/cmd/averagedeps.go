package cmd

import (
	"context"
	"encoding/json"
	"github.com/markuszm/npm-analysis/database"
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

const OutputPath = "/home/markus/npm-analysis/averageDeps.json"

var averageDepsCmd = &cobra.Command{
	Use:   "averageDeps",
	Short: "Averages dependencies",
	Long:  `...`,
	Run: func(cmd *cobra.Command, args []string) {
		depCountMap := make(map[time.Time]int, 0)
		packageCountMap := make(map[time.Time]int, 0)

		err := collectData(depCountMap, packageCountMap)
		if err != nil {
			log.Fatalf("error while collecting data with %v", err)
		}

		averagesMap := calculateAverages(depCountMap, packageCountMap)

		err = writeData(depCountMap, packageCountMap, averagesMap)
		if err != nil {
			log.Fatalf("error writing results with %v", err)
		}
	},
}

func init() {
	rootCmd.AddCommand(averageDepsCmd)
}

func calculateAverages(depCountMap map[time.Time]int, packageCountMap map[time.Time]int) map[time.Time]float64 {
	averagesMap := make(map[time.Time]float64, 0)
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
			depOverallCount := depCountMap[date]
			packageOverallCount := packageCountMap[date]

			average := float64(depOverallCount) / float64(packageOverallCount)
			averagesMap[date] = average
		}
	}
	return averagesMap
}

func writeData(depCountMap map[time.Time]int, packageCountMap map[time.Time]int, averagesMap map[time.Time]float64) error {
	bytes, err := json.Marshal(map[string]interface{}{
		"depCounts":    depCountMap,
		"packageCount": packageCountMap,
		"averages":     averagesMap,
	})
	if err != nil {
		return err
	}
	return ioutil.WriteFile(OutputPath, bytes, os.ModePerm)
}

func collectData(depCountMap, packageCountMap map[time.Time]int) error {
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
				depCountMap[t] += len(data.Dependencies)
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
