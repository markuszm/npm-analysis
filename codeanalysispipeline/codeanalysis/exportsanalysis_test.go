package codeanalysis

import (
	"go.uber.org/zap"
	"testing"
)

func TestExportES6(t *testing.T) {
	logger := zap.NewNop().Sugar()
	analysis := NewExportsAnalysis(logger)
	result, err := analysis.AnalyzePackage("./testfiles/exportes6test")

	if err != nil {
		t.Fatal(err)
	}

	exportsAnalysis := result.([]Export)

	t.Log(exportsAnalysis)
}

func TestExportCommonJS(t *testing.T) {
	logger := zap.NewNop().Sugar()
	analysis := NewExportsAnalysis(logger)
	result, err := analysis.AnalyzePackage("./testfiles/exportcjstest")

	if err != nil {
		t.Fatal(err)
	}

	exportsAnalysis := result.([]Export)

	t.Log(exportsAnalysis)
}
