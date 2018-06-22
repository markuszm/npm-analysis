package analysisimpl

import (
	"github.com/pkg/errors"
	"strconv"
	"strings"
)

type EmptyPackageAnalysis struct {
}

func (e *EmptyPackageAnalysis) AnalyzePackage(packagePath string) ([]string, error) {
	var results []string
	result, err := ExecuteCommand("find", packagePath, "-name", "*.js")
	if err != nil {
		return results, errors.Wrapf(err, "error analyzing package %v", packagePath)
	}
	lines := strings.Split(result, "\n")
	jsCount := strconv.Itoa(len(lines) - 1)
	results = append(results, jsCount)
	return results, nil
}
