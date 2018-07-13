package codeanalysis

import (
	"encoding/json"
	"github.com/pkg/errors"
	"go.uber.org/zap"
	"os"
	"path"
)

type CallgraphAnalysis struct {
	logger *zap.SugaredLogger
}

func NewCallgraphAnalysis(logger *zap.SugaredLogger) *CallgraphAnalysis {
	return &CallgraphAnalysis{logger}
}

func (e *CallgraphAnalysis) AnalyzePackage(packagePath string) (interface{}, error) {
	workingDir, err := os.Getwd()

	if err != nil {
		return nil, errors.Wrap(err, "error retrieving working directory of executable")
	}

	var pathToAnalysisExecutable string
	if path.Base(workingDir) == "codeanalysis" {
		pathToAnalysisExecutable = "./callgraph-analysis/analysis"
	} else {
		pathToAnalysisExecutable = "./codeanalysis/callgraph-analysis/analysis"
	}

	result, err := ExecuteCommand(pathToAnalysisExecutable, packagePath)
	if err != nil {
		return nil, errors.Wrap(err, "error executing callgraph analysis")
	}

	var analysisResult []Call
	err = json.Unmarshal([]byte(result), &analysisResult)
	if err != nil {
		e.logger.Error(result)
		return nil, errors.Wrap(err, "error parsing result")
	}

	return analysisResult, nil
}

type Call struct {
	FromFile     string   `json:"fromFile"`
	FromFunction string   `json:"fromFunction"`
	Receiver     string   `json:"receiver"`
	Module       string   `json:"module"`
	ToFile       string   `json:"toFile"`
	ToFunction   string   `json:"toFunction"`
	Arguments    []string `json:"arguments"`
}
