package codeanalysis

import (
	"github.com/markuszm/npm-analysis/model"
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

func (e *FileDistributionAnalysis) AnalyzePackageFiles(packagePath string) (interface{}, error) {
	extensionMap := make(map[string]int, 0)
	result, err := ExecuteCommand("find", packagePath, "-type", "f")
	if err != nil {
		return extensionMap, errors.Wrapf(err, "error analyzing package %v", packagePath)
	}
	lines := strings.Split(result, "\n")

	for _, l := range lines {
		if l == "" {
			continue
		}
		ext := path.Ext(l)
		if ext == "" {
			fileType, err := ExecuteCommand("file", l)
			if err != nil {
				continue
			}

			if strings.Contains(fileType, "executable") {
				extensionMap["binary"]++
				continue
			}

			if strings.Contains(fileType, "text") {
				extensionMap["text"]++
				continue
			}
		} else {
			extensionMap[ext]++
		}
	}

	return extensionMap, nil
}

func (e *FileDistributionAnalysis) AnalyzePackage(version model.PackageVersionPair) (interface{}, error) {
	return nil, errors.New("Unsupported for this analysis")
}
