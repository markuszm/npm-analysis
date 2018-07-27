package resultprocessing

import (
	"encoding/csv"
	"encoding/json"
	"github.com/markuszm/npm-analysis/model"
	"log"
	"os"
	"path"
	"strconv"
)

func CalculateUsedDependenciesRatio(resultPath string) (map[string]float64, map[string]float64, error) {
	file, err := os.Open(resultPath)
	if err != nil {
		return nil, nil, err
	}

	decoder := json.NewDecoder(file)

	usedDependencyRatios := make(map[string]float64, 0)
	usedDependencyRatiosWithoutZeroDependencies := make(map[string]float64, 0)

	for {
		result := model.PackageResult{}
		err := decoder.Decode(&result)
		if err != nil {
			if err.Error() == "EOF" {
				log.Print("finished decoding result json")
				break
			} else {
				return nil, nil, err
			}
		}

		switch result.Result.(type) {
		case map[string]interface{}:
			dependencyResult := result.Result.(map[string]interface{})
			dependencyCount := len(dependencyResult["Dependencies"].([]interface{}))
			usedDependenciesCount := len(dependencyResult["Used"].([]interface{}))
			ratio := 0.0
			if dependencyCount > 0 {
				ratio = float64(usedDependenciesCount) / float64(dependencyCount)
				usedDependencyRatios[result.Name] = ratio
				usedDependencyRatiosWithoutZeroDependencies[result.Name] = ratio
			} else {
				usedDependencyRatios[result.Name] = ratio
			}

		default:
			continue
		}

	}

	return usedDependencyRatios, usedDependencyRatiosWithoutZeroDependencies, nil
}

func WriteUsedDependencyRatios(dependencyRatios map[string]float64, filePath string) error {
	err := os.MkdirAll(path.Dir(filePath), os.ModePerm)
	if err != nil {
		return err
	}

	file, err := os.Create(filePath)
	if err != nil {
		return err
	}

	writer := csv.NewWriter(file)
	defer writer.Flush()
	for p, r := range dependencyRatios {
		writer.Write([]string{p, strconv.FormatFloat(r, 'f', 2, 64)})
	}

	return nil
}
