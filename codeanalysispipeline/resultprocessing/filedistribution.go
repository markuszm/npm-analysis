package resultprocessing

import (
	"encoding/json"
	"github.com/markuszm/npm-analysis/codeanalysispipeline"
	"github.com/markuszm/npm-analysis/util"
	"log"
	"os"
	"sort"
)

func MergeFileDistributionResult(resultPath string, filter int) ([]util.Pair, error) {
	file, err := os.Open(resultPath)
	if err != nil {
		return nil, err
	}

	decoder := json.NewDecoder(file)

	mergedDistribution := make(map[string]int, 0)

	for {
		result := codeanalysispipeline.PackageResult{}
		err := decoder.Decode(&result)
		if err != nil {
			if err.Error() == "EOF" {
				log.Print("finished decoding result json")
				break
			} else {
				return nil, err
			}
		}

		switch result.Result.(type) {
		case map[string]interface{}:
			extensionMap := result.Result.(map[string]interface{})
			for ext, count := range extensionMap {
				// JSON numbers are decoded float64 so this transformation to int is necessary
				mergedDistribution[ext] += int(count.(float64))
			}
		default:
			continue
		}

	}

	var sortedDistribution []util.Pair

	for k, v := range mergedDistribution {
		if v > filter {
			sortedDistribution = append(sortedDistribution, util.Pair{Key: k, Value: v})
		}
	}

	sort.Sort(sort.Reverse(util.PairList(sortedDistribution)))

	return sortedDistribution, nil
}
