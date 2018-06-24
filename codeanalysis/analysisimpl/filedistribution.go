package analysisimpl

import (
	"github.com/pkg/errors"
	"path"
	"strings"
)

type FileDistributionAnalysis struct {
}

func (e *FileDistributionAnalysis) AnalyzePackage(packagePath string) (interface{}, error) {
	extensionMap := make(map[string]int, 0)
	result, err := ExecuteCommand("find", packagePath)
	if err != nil {
		return extensionMap, errors.Wrapf(err, "error analyzing package %v", packagePath)
	}
	lines := strings.Split(result, "\n")

	for _, l := range lines {
		ext := path.Ext(l)
		if ext != "" {
			extensionMap[ext]++
		}
	}

	return extensionMap, nil
}
