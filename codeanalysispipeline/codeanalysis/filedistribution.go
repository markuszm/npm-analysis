package codeanalysis

import (
	"github.com/pkg/errors"
	"go.uber.org/zap"
	"path"
	"strings"
)

type FileDistributionAnalysis struct {
	logger *zap.SugaredLogger
}

func NewFileDistributionAnalysis(logger *zap.SugaredLogger) *FileDistributionAnalysis {
	return &FileDistributionAnalysis{logger}
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
