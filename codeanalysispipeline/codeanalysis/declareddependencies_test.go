package codeanalysis

import (
	"fmt"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestRequireDetected(t *testing.T) {
	analysis := UsedDependenciesAnalysis{}
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

func TestImportDetected(t *testing.T) {
	analysis := UsedDependenciesAnalysis{}
	result, err := analysis.AnalyzePackage("./testfiles/importtest")

	if err != nil {
		t.Fatal(err)
	}

	dependencyResult := result.(DependencyResult)

	t.Log(dependencyResult)

	a := assert.New(t)
	a.ElementsMatch(dependencyResult.Dependencies, []string{"foo", "bar", "foobar"}, fmt.Sprint(dependencyResult.Dependencies))
	a.ElementsMatch(dependencyResult.Imported, []string{"foo", "bar", "a", "b", "c", "d", "e", "f"}, fmt.Sprint(dependencyResult.Imported))
	a.ElementsMatch(dependencyResult.Used, []string{"foo", "bar"}, fmt.Sprint(dependencyResult.Used))
	a.ElementsMatch(dependencyResult.Required, []string{}, fmt.Sprint(dependencyResult.Required))
}

func TestTypescript(t *testing.T) {
	analysis := UsedDependenciesAnalysis{}
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
