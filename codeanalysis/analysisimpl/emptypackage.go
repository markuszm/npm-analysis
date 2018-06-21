package analysisimpl

import (
	"github.com/pkg/errors"
	"strconv"
	"strings"
)

type EmptyPackageAnalysis struct {
}

func (e *EmptyPackageAnalysis) AnalyzePackage(packagePath string) (string, error) {
	result, err := ExecuteCommand("find", packagePath, "-name", "*.js")
	if err != nil {
		return "", errors.Wrapf(err, "error analyzing package %v", packagePath)
	}
	lines := strings.Split(result, "\n")
	return strconv.Itoa(len(lines) - 1), nil
}
