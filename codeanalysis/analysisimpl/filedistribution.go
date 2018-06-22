package analysisimpl

import (
	"fmt"
	"github.com/pkg/errors"
	"path"
	"strings"
)

type FileDistributionAnalysis struct {
}

func (e *FileDistributionAnalysis) AnalyzePackage(packagePath string) ([]string, error) {
	var results []string
	result, err := ExecuteCommand("find", packagePath)
	if err != nil {
		return results, errors.Wrapf(err, "error analyzing package %v", packagePath)
	}
	lines := strings.Split(result, "\n")

	extensionMap := make(map[string]int, 0)
	for _, l := range lines {
		ext := path.Ext(l)
		if ext != "" {
			extensionMap[ext]++
		}
	}

	for ext, count := range extensionMap {
		results = append(results, fmt.Sprintf("%v-%v", ext, count))
	}
	return results, nil
}
