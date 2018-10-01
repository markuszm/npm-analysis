package codeanalysis

import (
	"encoding/json"
	"github.com/markuszm/npm-analysis/model"
	"github.com/pkg/errors"
	"go.uber.org/zap"
)

type DynamicExportsAnalysis struct {
	logger       *zap.SugaredLogger
	analysisPath string
}

const CONTAINER_TAG = "dynamicexport"

func NewDynamicExportsAnalysis(logger *zap.SugaredLogger, analysisPath string) *DynamicExportsAnalysis {
	err := BuildImage(analysisPath, CONTAINER_TAG)
	if err != nil {
		logger.Fatalw("error building analysis docker image", "err", err)
	}
	return &DynamicExportsAnalysis{logger, analysisPath}
}

func (e *DynamicExportsAnalysis) AnalyzePackage(version model.PackageVersionPair) (interface{}, error) {
	result, err := RunDockerContainer(CONTAINER_TAG, version.Name, version.Version)
	if err != nil {
		return nil, errors.Wrap(err, "error executing dynamic export analysis")
	}

	var analysisResult []interface{}
	err = json.Unmarshal([]byte(result), &analysisResult)
	if err != nil {
		e.logger.Error(result)
		return nil, errors.Wrap(err, "error parsing result")
	}

	return analysisResult, nil
}

func (e *DynamicExportsAnalysis) AnalyzePackageFiles(packagePath string) (interface{}, error) {
	return nil, errors.New("Unsupported for this analysis")
}
