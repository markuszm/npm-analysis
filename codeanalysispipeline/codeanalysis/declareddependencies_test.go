package codeanalysis

import (
	"fmt"
	"github.com/stretchr/testify/assert"
	"go.uber.org/zap"
	"testing"
)

func TestRequireDetected(t *testing.T) {
	logger := zap.NewNop().Sugar()
	analysis := NewUsedDependenciesAnalysis(logger)
	result, err := analysis.AnalyzePackage("./testfiles/requiretest")

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

func TestRequireMinified(t *testing.T) {
	logger := zap.NewNop().Sugar()
	analysis := NewUsedDependenciesAnalysis(logger)
	result, err := analysis.AnalyzePackage("./testfiles/requireminifiedtest")

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
	analysis := NewUsedDependenciesAnalysis(logger)
	result, err := analysis.AnalyzePackage("./testfiles/importtest")

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
	// skip this test for now as this can not be solved by grep regex
	t.Skip()

	logger := zap.NewNop().Sugar()
	analysis := NewUsedDependenciesAnalysis(logger)
	result, err := analysis.AnalyzePackage("./testfiles/importminifiedtest")

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

func TestParsePackageImport(t *testing.T) {
	tests := []struct {
		value, expected string
	}{
		{`import * as mediaTestHelpers from '@atlaskit/media-test-helpers'"`, "@atlaskit/media-test-helpers"},
		{"import hbs from 'htmlbars-inline-precompile';\nvar compiled = hbs`hello`;", "htmlbars-inline-precompile"},
	}

	for _, test := range tests {
		logger := zap.NewNop().Sugar()
		analysis := NewUsedDependenciesAnalysis(logger)
		t.Run(fmt.Sprintf("Value: %v Expected: %v", test.value, test.expected), func(t *testing.T) {
			assert.Equal(t, test.expected, analysis.parseModuleFromImportStmt(test.value))
		})
	}

}

func TestTypescript(t *testing.T) {
	logger := zap.NewNop().Sugar()
	analysis := NewUsedDependenciesAnalysis(logger)
	result, err := analysis.AnalyzePackage("./testfiles/typescripttest")

	if err != nil {
		t.Fatal(err)
	}

	dependencyResult := result.(DependencyResult)

	t.Log(dependencyResult)

	a := assert.New(t)
	a.ElementsMatch(dependencyResult.Dependencies, []string{"foo", "bar", "foobar"}, fmt.Sprint(dependencyResult.Dependencies))
	a.ElementsMatch(dependencyResult.Imported, []string{"abc", "bar", "JQuery"}, fmt.Sprint(dependencyResult.Imported))
	a.ElementsMatch(dependencyResult.Used, []string{"foo", "bar"}, fmt.Sprint(dependencyResult.Used))
	a.ElementsMatch(dependencyResult.Required, []string{"foo"}, fmt.Sprint(dependencyResult.Required))
}
