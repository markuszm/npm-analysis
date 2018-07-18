package codeanalysis

import (
	"github.com/stretchr/testify/assert"
	"go.uber.org/zap"
	"testing"
)

func TestFileDistribution(t *testing.T) {
	logger := zap.NewNop().Sugar()
	analysis := NewFileDistributionAnalysis(logger)
	result, err := analysis.AnalyzePackage("./testfiles/filedistribution")

	if err != nil {
		t.Fatal(err)
	}

	extensionMap := result.(map[string]int)

	expectedMap := map[string]int{
		".json":  1,
		".js":    8,
		".ts":    2,
		".jsx":   1,
		".mjs":   1,
		"binary": 2,
		"text":   1,
	}

	t.Log(extensionMap)

	a := assert.New(t)
	a.Equal(expectedMap, extensionMap)
}
