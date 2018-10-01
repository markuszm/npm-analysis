package codeanalysis

import (
	"fmt"
	"github.com/stretchr/testify/assert"
	"go.uber.org/zap"
	"testing"
)

const analysisPath = "./import-analysis/analysis"

func TestRequireDetected(t *testing.T) {
	logger := zap.NewNop().Sugar()
	analysis := NewUsedDependenciesAnalysis(logger, analysisPath)
	result, err := analysis.AnalyzePackageFiles("./testfiles/import/requiretest")

	if err != nil {
		t.Fatal(err)
	}

	dependencyResult := result.(DependencyResult)

	t.Log(dependencyResult)

	a := assert.New(t)
	a.ElementsMatch(dependencyResult.Dependencies, []string{"foo", "bar", "foobar"}, fmt.Sprint(dependencyResult.Dependencies))
	a.ElementsMatch(dependencyResult.Required, []string{"foo", "bar", "abc", "oauth", "b", "f"}, fmt.Sprint(dependencyResult.Required))
	a.ElementsMatch(dependencyResult.Imported, []string{}, fmt.Sprint(dependencyResult.Imported))
	a.ElementsMatch(dependencyResult.Used, []string{"foo", "bar"}, fmt.Sprint(dependencyResult.Used))
}

func TestRequireMinified(t *testing.T) {
	logger := zap.NewNop().Sugar()
	analysis := NewUsedDependenciesAnalysis(logger, analysisPath)
	result, err := analysis.AnalyzePackageFiles("./testfiles/import/requireminifiedtest")

	if err != nil {
		t.Fatal(err)
	}

	dependencyResult := result.(DependencyResult)

	t.Log(dependencyResult)

	a := assert.New(t)
	a.ElementsMatch(dependencyResult.Dependencies, []string{"foo", "bar", "foobar"}, fmt.Sprint(dependencyResult.Dependencies))
	a.ElementsMatch(dependencyResult.Required, []string{"foo", "bar", "abc"}, fmt.Sprint(dependencyResult.Required))
	a.ElementsMatch(dependencyResult.Imported, []string{}, fmt.Sprint(dependencyResult.Imported))
	a.ElementsMatch(dependencyResult.Used, []string{"foo", "bar"}, fmt.Sprint(dependencyResult.Used))
}

func TestImportDetected(t *testing.T) {
	logger := zap.NewNop().Sugar()
	analysis := NewUsedDependenciesAnalysis(logger, analysisPath)
	result, err := analysis.AnalyzePackageFiles("./testfiles/import/importtest")

	if err != nil {
		t.Fatal(err)
	}

	dependencyResult := result.(DependencyResult)

	t.Log(dependencyResult)

	a := assert.New(t)
	a.ElementsMatch(dependencyResult.Dependencies, []string{"foo", "bar", "foobar"}, fmt.Sprint(dependencyResult.Dependencies))
	a.ElementsMatch(dependencyResult.Imported, []string{"foo", "bar", "a", "b", "c", "d", "e", "f", "g"}, fmt.Sprint(dependencyResult.Imported))
	a.ElementsMatch(dependencyResult.Used, []string{"foo", "bar"}, fmt.Sprint(dependencyResult.Used))
	a.ElementsMatch(dependencyResult.Required, []string{}, fmt.Sprint(dependencyResult.Required))
}

func TestImportMinified(t *testing.T) {
	logger := zap.NewNop().Sugar()
	analysis := NewUsedDependenciesAnalysis(logger, analysisPath)
	result, err := analysis.AnalyzePackageFiles("./testfiles/import/importminifiedtest")

	if err != nil {
		t.Fatal(err)
	}

	dependencyResult := result.(DependencyResult)

	t.Log(dependencyResult)

	a := assert.New(t)
	a.ElementsMatch(dependencyResult.Dependencies, []string{"foo", "bar", "foobar"}, fmt.Sprint(dependencyResult.Dependencies))
	a.ElementsMatch(dependencyResult.Imported, []string{"foo", "bar", "a", "b", "c", "d", "e", "f", "g"}, fmt.Sprint(dependencyResult.Imported))
	a.ElementsMatch(dependencyResult.Used, []string{"foo", "bar"}, fmt.Sprint(dependencyResult.Used))
	a.ElementsMatch(dependencyResult.Required, []string{}, fmt.Sprint(dependencyResult.Required))
}

func TestTypescript(t *testing.T) {
	logger := zap.NewNop().Sugar()
	analysis := NewUsedDependenciesAnalysis(logger, analysisPath)
	result, err := analysis.AnalyzePackageFiles("./testfiles/import/typescripttest")

	if err != nil {
		t.Fatal(err)
	}

	dependencyResult := result.(DependencyResult)

	t.Log(dependencyResult)

	a := assert.New(t)
	a.ElementsMatch(dependencyResult.Dependencies, []string{"foo", "bar", "foobar"}, fmt.Sprint(dependencyResult.Dependencies))
	a.ElementsMatch(dependencyResult.Imported, []string{"abc", "bar", "JQuery"}, fmt.Sprint(dependencyResult.Imported))
	a.ElementsMatch(dependencyResult.Used, []string{"bar"}, fmt.Sprint(dependencyResult.Used))
	a.ElementsMatch(dependencyResult.Required, []string{}, fmt.Sprint(dependencyResult.Required))
}
