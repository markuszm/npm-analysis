package codeanalysis

import (
	"encoding/json"
	"github.com/pkg/errors"
	"go.uber.org/zap"
	"os"
	"path"
)

type ExportsAnalysis struct {
	logger *zap.SugaredLogger
}

func NewExportsAnalysis(logger *zap.SugaredLogger) *ExportsAnalysis {
	return &ExportsAnalysis{logger}
}

func (e *ExportsAnalysis) AnalyzePackage(packagePath string) (interface{}, error) {
	workingDir, err := os.Getwd()

	if err != nil {
		return nil, errors.Wrap(err, "error retrieving working directory of executable")
	}

	var pathToAnalysisExecutable string
	if path.Base(workingDir) == "codeanalysis" {
		pathToAnalysisExecutable = "./exports-analysis/analysis"
	} else {
		pathToAnalysisExecutable = "./codeanalysis/exports-analysis/analysis"
	}

	result, err := ExecuteCommand(pathToAnalysisExecutable, "folder", packagePath)
	if err != nil {
		return nil, errors.Wrap(err, "error executing export analysis")
	}

	var analysisResult []Export
	err = json.Unmarshal([]byte(result), &analysisResult)
	if err != nil {
		e.logger.Error(result)
		return nil, errors.Wrap(err, "error parsing result")
	}

	return analysisResult, nil
}

type Export struct {
	ExportType string `json:"type"`
	Identifier string `json:"id"`
	BundleType string `json:"bundleType"`
}
