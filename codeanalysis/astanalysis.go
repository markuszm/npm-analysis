package codeanalysis

import (
	"encoding/json"
	"github.com/markuszm/npm-analysis/model"
	"github.com/pkg/errors"
	"go.uber.org/zap"
)

const FileSizeLimit = "250000"

type ASTAnalysis struct {
	logger       *zap.SugaredLogger
	analysisPath string
}

func NewASTAnalysis(logger *zap.SugaredLogger, analysisPath string) *ASTAnalysis {
	return &ASTAnalysis{logger, analysisPath}
}

func (e *ASTAnalysis) AnalyzePackageFiles(packagePath string) (interface{}, error) {
	result, err := ExecuteCommand(e.analysisPath, packagePath, FileSizeLimit)
	if err != nil {
		return nil, errors.Wrap(err, "error executing ast analysis")
	}

	var analysisResult []interface{}
	err = json.Unmarshal([]byte(result), &analysisResult)
	if err != nil {
		e.logger.Error(result)
		return nil, errors.Wrap(err, "error parsing result")
	}

	return analysisResult, nil
}

func (e *ASTAnalysis) AnalyzePackage(version model.PackageVersionPair) (interface{}, error) {
	return nil, errors.New("Unsupported for this analysis")
}
