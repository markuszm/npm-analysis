package resultprocessing

import (
	"encoding/json"
	"github.com/markuszm/npm-analysis/model"
	"strings"
	"testing"
)

const example = `{
  "Name": "ui",
  "Version": "0.2.4",
  "Result": {
    "Required": [
      "std/Class",
      "std/isArray",
      "std/curry",
      "std/extend",
      "std/unique",
      "std/client",
      "std/recall",
      "std/Publisher",
      "std/each",
      "std/slice",
      "std/isArguments",
      "std/arrayToObject",
      "std/bind"
    ],
    "Imported": [],
    "Dependencies": [
      "std"
    ],
    "Used": [
      "std/Class",
      "std/Publisher",
      "std/each",
      "std/slice",
      "std/isArguments",
      "std/isArray",
      "std/arrayToObject",
      "std/curry",
      "std/bind",
      "std/extend",
      "std/unique",
      "std/client",
      "std/recall"
    ]
  }
}
`

func TestCalculateDependencyRatio(t *testing.T) {
	result := model.PackageResult{}
	err := json.Unmarshal([]byte(example), &result)
	if err != nil {
		t.Fatal(err)
	}

	dependencyResult := result.Result.(map[string]interface{})

	dependencies := dependencyResult["Dependencies"].([]interface{})
	usedDependencies := dependencyResult["Used"].([]interface{})
	dependencyCount := len(dependencies)
	usedDependenciesCount := 0
	for _, d := range dependencies {
		for _, u := range usedDependencies {
			if strings.Contains(u.(string), d.(string)) {
				usedDependenciesCount++
				break
			}
		}
	}
	ratio := 0.0
	ratio = float64(usedDependenciesCount) / float64(dependencyCount)

	t.Log(ratio)
}
