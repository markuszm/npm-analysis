package codeanalysis

import (
	"encoding/json"
	"github.com/pkg/errors"
	"go.uber.org/zap"
)

type ASTAnalysis struct {
	logger       *zap.SugaredLogger
	analysisPath string
}

func NewASTAnalysis(logger *zap.SugaredLogger, analysisPath string) *ASTAnalysis {
	return &ASTAnalysis{logger, analysisPath}
}

func (e *ASTAnalysis) AnalyzePackage(packagePath string) (interface{}, error) {
	result, err := ExecuteCommand(e.analysisPath, packagePath)
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
