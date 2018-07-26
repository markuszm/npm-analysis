package codeanalysis

import (
	"encoding/json"
	"fmt"
	"github.com/markuszm/npm-analysis/resultprocessing"
	"github.com/stretchr/testify/assert"
	"go.uber.org/zap"
	"testing"
)

func transformToCalls(result interface{}) ([]resultprocessing.Call, error) {
	objs := result.([]interface{})

	var calls []resultprocessing.Call

	for _, value := range objs {
		call := resultprocessing.Call{}
		bytes, err := json.Marshal(value)
		if err != nil {
			return calls, err
		}
		err = json.Unmarshal(bytes, &call)
		if err != nil {
			return calls, err
		}
		calls = append(calls, call)
	}
	return calls, nil
}

func TestCallgraphLocal(t *testing.T) {
	const analysisPath = "./callgraph-analysis/analysis"

	logger := zap.NewNop().Sugar()
	analysis := NewASTAnalysis(logger, analysisPath)
	result, err := analysis.AnalyzePackage("./testfiles/callgraph/local")

	if err != nil {
		t.Fatal(err)
	}

	calls, err := transformToCalls(result)
	if err != nil {
		t.Fatal(err)
	}

	expectedCalls := []resultprocessing.Call{
		{"call.js", "call.js", "this", "", "fun.js", "myfun", []string{"2"}},
		{"fun.js", "myfun", "this", "", "fun.js", "otherfun", []string{"x"}},
		{"fun.js", "otherfun", "this", "", "fun.js", "anotherfun", []string{"y"}},
	}
	assert.ElementsMatch(t, calls, expectedCalls, fmt.Sprint(calls))
}

func TestCallgraphModule(t *testing.T) {
	const analysisPath = "./callgraph-analysis/analysis"

	logger := zap.NewNop().Sugar()
	analysis := NewASTAnalysis(logger, analysisPath)
	result, err := analysis.AnalyzePackage("./testfiles/callgraph/modules")

	if err != nil {
		t.Fatal(err)
	}

	calls, err := transformToCalls(result)
	if err != nil {
		t.Fatal(err)
	}

	expectedCalls := []resultprocessing.Call{
		{"calls.js", "calls.js", "this", "", "", "require", []string{"foo"}},
		{"calls.js", "calls.js", "this", "", "", "require", []string{"bar"}},
		{"calls.js", "calls.js", "this", "", "", "require", []string{"foobar"}},
		{"calls.js", "calls.js", "this", "", "", "require", []string{"oauth"}},
		{"calls.js", "calls.js", "OAuth", "oauth", "calls.js", "someMethod", []string{}},
		{"calls.js", "calls.js", "f", "foo", "calls.js", "a", []string{}},
		{"calls.js", "calls.js", "bar", "bar", "calls.js", "b", []string{"a"}},
		{"calls.js", "calls.js", "foobar", "foobar", "calls.js", "func", []string{"a", "b"}},
	}

	assert.ElementsMatch(t, calls, expectedCalls, fmt.Sprint(calls))
}

func TestCallgraphES6Module(t *testing.T) {
	const analysisPath = "./callgraph-analysis/analysis"

	logger := zap.NewNop().Sugar()
	analysis := NewASTAnalysis(logger, analysisPath)
	result, err := analysis.AnalyzePackage("./testfiles/callgraph/es6modules")

	if err != nil {
		t.Fatal(err)
	}

	calls, err := transformToCalls(result)
	if err != nil {
		t.Fatal(err)
	}

	expectedCalls := []resultprocessing.Call{
		{"call.js", "foo", "_", "underscore", "call.js", "map", []string{"aList", "(i) => {...}"}},
		{"call.js", "foo", "bar", "foobar", "call.js", "add", []string{"i"}},
		{"call.js", "foo", "this", "b", "call.js", "a", []string{"mappedList"}},
	}

	assert.ElementsMatch(t, calls, expectedCalls, fmt.Sprint(calls))
}

func TestCallgraphMix(t *testing.T) {
	const analysisPath = "./callgraph-analysis/analysis"

	logger := zap.NewNop().Sugar()
	analysis := NewASTAnalysis(logger, analysisPath)
	result, err := analysis.AnalyzePackage("./testfiles/callgraph/mix")

	if err != nil {
		t.Fatal(err)
	}

	calls, err := transformToCalls(result)
	if err != nil {
		t.Fatal(err)
	}

	expectedCalls := []resultprocessing.Call{
		{FromFile: "anotherFile.js", FromFunction: "aFnInAnotherFile", Receiver: "console", ToFunction: "log", Arguments: []string{"cool"}},
		{FromFile: "file.js", FromFunction: "file.js", Receiver: "this", ToFunction: "require", Arguments: []string{"./anotherFile"}},
		{FromFile: "file.js", FromFunction: "aFunction", Receiver: "_", ToFunction: "curry", Arguments: []string{"libVar.referencedFn"}},
		{FromFile: "file.js", FromFunction: "aFunction", Receiver: "async", ToFunction: "series", Arguments: []string{"[_.curry(libVar.referencedFn)]"}},
		{FromFile: "file.js", FromFunction: "aFunction", Receiver: "libVar", Module: "./anotherFile", ToFile: "file.js", ToFunction: "aFnInAnotherFile", Arguments: []string{"n + 1"}},
		{FromFile: "file.js", FromFunction: "file.js", Receiver: "libVar", Module: "./anotherFile", ToFile: "file.js", ToFunction: "aFnInAnotherFile", Arguments: []string{"2"}},
		{FromFile: "file.js", FromFunction: "file.js", Receiver: "this", ToFile: "file.js", ToFunction: "aFunction", Arguments: []string{}},
	}

	assert.ElementsMatch(t, calls, expectedCalls, fmt.Sprint(calls))
}

func TestCallgraphScoping(t *testing.T) {
	const analysisPath = "./callgraph-analysis/analysis"

	logger := zap.NewNop().Sugar()
	analysis := NewASTAnalysis(logger, analysisPath)
	result, err := analysis.AnalyzePackage("./testfiles/callgraph/scoping")

	if err != nil {
		t.Fatal(err)
	}

	calls, err := transformToCalls(result)
	if err != nil {
		t.Fatal(err)
	}

	expectedCalls := []resultprocessing.Call{
		{"calls.js", "calls.js", "this", "", "", "require", []string{"foo"}},
		{"calls.js", "calls.js", "this", "", "", "require", []string{"bar"}},
		{"calls.js", "calls.js", "this", "", "", "require", []string{"foobar"}},
		{"calls.js", "calls.js", "this", "", "", "require", []string{"foobar"}},
		{"calls.js", "f", "this", "", "", "require", []string{"foo"}},
		{"calls.js", "f", "this", "", "", "require", []string{"bar"}},
		{"calls.js", "g", "foo", "foobar", "calls.js", "someMethod", []string{}},
		{"calls.js", "g", "foobar", "foobar", "calls.js", "otherMethod", []string{}},
		{"calls.js", "h", "bar", "", "calls.js", "someMethod", []string{}},
		{"calls.js", "h", "fooVar", "", "calls.js", "someMethod", []string{}},
	}

	assert.ElementsMatch(t, calls, expectedCalls, fmt.Sprint(calls))
}
