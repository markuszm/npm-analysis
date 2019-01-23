package maintainerreach

import (
	"context"
	"encoding/json"
	"github.com/markuszm/npm-analysis/database"
	"github.com/markuszm/npm-analysis/model"
	"github.com/mongodb/mongo-go-driver/bson"
	"io/ioutil"
	"log"
	"os"
	"time"
)

func LoadJSONDependenciesTimeline(path string) map[time.Time]map[string]map[string]bool {
	log.Print("Loading json")
	bytes, err := ioutil.ReadFile(path)
	if err != nil {
		log.Fatal(err)
	}
	var timeline map[time.Time]map[string]map[string]bool
	err = json.Unmarshal(bytes, &timeline)
	if err != nil {
		log.Fatal(err)
	}
	log.Print("Finished loading json")
	return timeline
}

func LoadJSONLatestDependencies(path string) map[string]map[string]bool {
	log.Print("Loading json")
	bytes, err := ioutil.ReadFile(path)
	if err != nil {
		log.Fatal(err)
	}
	var latestDependencies map[string]map[string]bool
	err = json.Unmarshal(bytes, &latestDependencies)
	if err != nil {
		log.Fatal(err)
	}
	log.Print("Finished loading json")
	return latestDependencies
}

func LoadJSONMaintainersTimeline(path string) map[time.Time]map[string][]string {
	log.Print("Loading json")
	bytes, err := ioutil.ReadFile(path)
	if err != nil {
		log.Fatal(err)
	}
	var timeline map[time.Time]map[string][]string
	err = json.Unmarshal(bytes, &timeline)
	if err != nil {
		log.Fatal(err)
	}
	log.Print("Finished loading json")
	return timeline
}

func LoadJSONLatestMaintainers(path string) map[string][]string {
	log.Print("Loading json")
	bytes, err := ioutil.ReadFile(path)
	if err != nil {
		log.Fatal(err)
	}
	var latestMaintainers map[string][]string
	err = json.Unmarshal(bytes, &latestMaintainers)
	if err != nil {
		log.Fatal(err)
	}
	log.Print("Finished loading json")
	return latestMaintainers
}

func GenerateJSONLatestMaintainers(timelinePath, outputPath string, date time.Time) {
	maintainerCostMaintainerTimeline := LoadJSONMaintainersTimeline(timelinePath)
	latestMaintainers := maintainerCostMaintainerTimeline[date]
	bytes, err := json.Marshal(latestMaintainers)
	if err != nil {
		log.Fatal(err)
	}
	log.Print("Finished transforming to JSON")
	err = ioutil.WriteFile(outputPath, bytes, os.ModePerm)
	if err != nil {
		log.Fatal(err)
	}

}

func GenerateJSONLatestDependencies(timelinePath, outputPath string, date time.Time) {
	maintainerCostDependenciesTimeline := LoadJSONDependenciesTimeline(timelinePath)
	latestDependencies := maintainerCostDependenciesTimeline[date]

	bytes, err := json.Marshal(latestDependencies)
	if err != nil {
		log.Fatal(err)
	}
	log.Print("Finished transforming to JSON")
	err = ioutil.WriteFile(outputPath, bytes, os.ModePerm)
	if err != nil {
		log.Fatal(err)
	}
}

func GenerateDependentsMaps(dependenciesTimeline map[time.Time]map[string]map[string]bool) map[time.Time]map[string][]string {
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

func GenerateTimeLatestVersionMap(mongoUrl, outputPath string) {
	dependenciesTimeline := make(map[time.Time]map[string]map[string]bool, 0)

	mongoDB := database.NewMongoDB(mongoUrl, "npm", "timeline")

	mongoDB.Connect()
	defer mongoDB.Disconnect()

	startTime := time.Now()

	cursor, err := mongoDB.ActiveCollection.Find(context.Background(), bson.D{})
	if err != nil {
		log.Fatal(err)
	}
	i := 0
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
	ioutil.WriteFile(outputPath, bytes, os.ModePerm)
}

func GenerateTimeMaintainersMap(mongoUrl, outputPath string) {
	maintainersTimeline := make(map[time.Time]map[string][]string, 0)

	mongoDB := database.NewMongoDB(mongoUrl, "npm", "timeline")

	mongoDB.Connect()
	defer mongoDB.Disconnect()

	startTime := time.Now()

	cursor, err := mongoDB.ActiveCollection.Find(context.Background(), bson.D{})
	if err != nil {
		log.Fatal(err)
	}
	i := 0
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

		for t, pkg := range timeMap {
			if maintainersTimeline[t] == nil {
				maintainersTimeline[t] = make(map[string][]string, 0)
			}
			maintainers := make([]string, 0)
			for _, m := range pkg.Maintainers {
				maintainers = append(maintainers, m)
			}

			maintainersTimeline[t][doc.Key] = maintainers
		}

		if i%10000 == 0 {
			log.Printf("Finished %v packages", i)
		}
		i++
	}
	cursor.Close(context.Background())
	endTime := time.Now()
	log.Printf("Took %v minutes to process all Documents from MongoDB", endTime.Sub(startTime).Minutes())
	bytes, err := json.Marshal(maintainersTimeline)
	if err != nil {
		log.Fatal(err)
	}
	log.Print("Finished transforming to JSON")
	ioutil.WriteFile(outputPath, bytes, os.ModePerm)
}
