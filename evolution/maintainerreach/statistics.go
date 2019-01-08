package maintainerreach

import (
	"encoding/json"
	"fmt"
	"github.com/markuszm/npm-analysis/plots"
	"github.com/markuszm/npm-analysis/util"
	"io/ioutil"
	"log"
	"math"
	"os"
	"path"
	"sort"
	"strings"
	"sync"
	"time"
)

func CalculateMaintainerReachDiff(outputName string, resultMap *sync.Map) {
	maintainerReachDiffMap := make(map[time.Time][]util.MaintainerReachDiff, 0)

	resultMap.Range(func(key, value interface{}) bool {
		counts := value.([]float64)
		x := 0
		isActive := false
		previousCount := math.MaxFloat64
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
				count := counts[x]
				if count > 0 || isActive {
					diff := count - previousCount
					if previousCount == math.MaxFloat64 {
						diff = count
					}
					diffs := maintainerReachDiffMap[date]
					if diffs == nil {
						diffs = make([]util.MaintainerReachDiff, 0)
					}
					diffs = append(diffs, util.MaintainerReachDiff{Name: key.(string), Diff: diff})
					maintainerReachDiffMap[date] = diffs
					isActive = true
					previousCount = count
				}
				x++
			}
		}
		return true
	})

	builder := strings.Builder{}

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
			diffs := maintainerReachDiffMap[date]
			sortedList := util.MaintainerReachDiffList(diffs)
			sort.Sort(sortedList)

			builder.WriteString(fmt.Sprintf("Top 20 decreases in %v \n", date))
			for i, m := range sortedList {
				if i > 19 {
					break
				}
				builder.WriteString(fmt.Sprintf("%d. Name: %v Diff: %f \n", i+1, m.Name, m.Diff))
			}

			sort.Sort(sort.Reverse(sortedList))
			builder.WriteString(fmt.Sprintf("Top 20 increases in %v \n", date))
			for i, m := range sortedList {
				if i > 19 {
					break
				}
				builder.WriteString(fmt.Sprintf("%d. Name: %v Diff: %f \n", i+1, m.Name, m.Diff))
			}
		}
	}

	outputPath := "/home/markus/npm-analysis/" + outputName + ".txt"
	ioutil.WriteFile(outputPath, []byte(builder.String()), os.ModePerm)
}

type PackageReachDiff struct {
	Diff float64
	Time time.Time
}

func CalculatePackageReachDiff(resultMap *sync.Map) error {
	packageReachDiffMap := make(map[string][]PackageReachDiff, 0)

	resultMap.Range(func(key, value interface{}) bool {
		counts := value.([]float64)
		x := 0
		isActive := false
		previousCount := math.MaxFloat64
		pkg := key.(string)
		var packageReachDiffs []PackageReachDiff
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
				count := counts[x]
				if count > 0 || isActive {
					diff := count - previousCount
					if previousCount == math.MaxFloat64 {
						diff = count
					}
					packageReachDiffs = append(packageReachDiffs, PackageReachDiff{Diff: diff, Time: date})
					isActive = true
					previousCount = count
				}
				x++
			}
		}
		packageReachDiffMap[pkg] = packageReachDiffs
		return true
	})

	bytes, err := json.Marshal(packageReachDiffMap)
	if err != nil {
		return err
	}

	outputPath := "/home/markus/npm-analysis/packageReachDiffs.json"
	err = ioutil.WriteFile(outputPath, bytes, os.ModePerm)
	return err
}

func CalculateAverageMaintainerReach(outputName string, resultMap *sync.Map) {
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

	filePath := path.Join("/home/markus/npm-analysis/", outputName+".json")
	err = ioutil.WriteFile(filePath, jsonBytes, os.ModePerm)
	if err != nil {
		log.Fatal(err)
	}

	plots.GenerateLinePlotForAverageMaintainerReach(outputName, avgValues)
}