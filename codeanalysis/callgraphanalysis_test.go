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
	calls := getCallsFromPackagePath("./testfiles/callgraph/local", t)

	expectedCalls := []resultprocessing.Call{
		{FromFile: "call.js", FromFunction: ".root", Receiver: "this", Module: []string{}, ToFile: "fun.js", ToFunction: "myfun", Arguments: []string{"2"}},
		{FromFile: "fun.js", FromFunction: "myfun", Receiver: "this", Module: []string{}, ToFile: "fun.js", ToFunction: "otherfun", Arguments: []string{"x"}},
		{FromFile: "fun.js", FromFunction: "otherfun", Receiver: "this", Module: []string{}, ToFile: "fun.js", ToFunction: "anotherfun", Arguments: []string{"y"}},
	}
	assert.ElementsMatch(t, calls, expectedCalls, fmt.Sprint(calls))
}

func TestCallgraphModule(t *testing.T) {
	calls := getCallsFromPackagePath("./testfiles/callgraph/modules", t)

	expectedCalls := []resultprocessing.Call{
		{FromFile: "calls.js", FromFunction: ".root", Receiver: "this", Module: []string{}, ToFunction: "require", Arguments: []string{"foo"}},
		{FromFile: "calls.js", FromFunction: ".root", Receiver: "this", Module: []string{}, ToFunction: "require", Arguments: []string{"bar"}},
		{FromFile: "calls.js", FromFunction: ".root", Receiver: "this", Module: []string{}, ToFunction: "require", Arguments: []string{"foobar"}},
		{FromFile: "calls.js", FromFunction: ".root", Receiver: "f", Module: []string{"foo"}, ToFile: "calls.js", ToFunction: "a", Arguments: []string{}},
		{FromFile: "calls.js", FromFunction: ".root", Receiver: "bar", Module: []string{"bar"}, ToFile: "calls.js", ToFunction: "b", Arguments: []string{"a"}},
		{FromFile: "calls.js", FromFunction: ".root", Receiver: "foobar", Module: []string{"foobar"}, ToFile: "calls.js", ToFunction: "func", Arguments: []string{"a", "b"}},
	}

	assert.ElementsMatch(t, calls, expectedCalls, fmt.Sprint(calls))
}

func TestCallgraphES6Module(t *testing.T) {
	calls := getCallsFromPackagePath("./testfiles/callgraph/es6modules", t)

	expectedCalls := []resultprocessing.Call{
		{FromFile: "call.js", FromFunction: "foo", Receiver: "_", Module: []string{"underscore"}, ToFile: "call.js", ToFunction: "map", Arguments: []string{"aList", "(i) => {...}"}},
		{FromFile: "call.js", FromFunction: "foo", Receiver: "bar", Module: []string{"foobar"}, ToFile: "call.js", ToFunction: "add", Arguments: []string{"i"}},
		{FromFile: "call.js", FromFunction: "foo", Receiver: "this", Module: []string{"b"}, ToFile: "call.js", ToFunction: "a", Arguments: []string{"mappedList"}},
	}

	assert.ElementsMatch(t, calls, expectedCalls, fmt.Sprint(calls))
}

func TestCallgraphMix(t *testing.T) {
	calls := getCallsFromPackagePath("./testfiles/callgraph/mix", t)

	expectedCalls := []resultprocessing.Call{
		{FromFile: "anotherFile.js", FromFunction: "aFnInAnotherFile", Receiver: "console", Module: []string{}, ToFunction: "log", Arguments: []string{"cool"}},
		{FromFile: "file.js", FromFunction: ".root", Receiver: "this", Module: []string{}, ToFunction: "require", Arguments: []string{"./anotherFile"}},
		{FromFile: "file.js", FromFunction: "aFunction", Receiver: "_", Module: []string{}, ToFunction: "curry", Arguments: []string{"libVar.referencedFn"}},
		{FromFile: "file.js", FromFunction: "aFunction", Receiver: "async", Module: []string{}, ToFunction: "series", Arguments: []string{"[_.curry(libVar.referencedFn)]"}},
		{FromFile: "file.js", FromFunction: "aFunction", Receiver: "libVar", Module: []string{"./anotherFile"}, ToFile: "file.js", ToFunction: "aFnInAnotherFile", Arguments: []string{"n + 1"}},
		{FromFile: "file.js", FromFunction: ".root", Receiver: "libVar", Module: []string{"./anotherFile"}, ToFile: "file.js", ToFunction: "aFnInAnotherFile", Arguments: []string{"2"}},
		{FromFile: "file.js", FromFunction: ".root", Receiver: "this", Module: []string{}, ToFile: "file.js", ToFunction: "aFunction", Arguments: []string{}},
	}

	assert.ElementsMatch(t, calls, expectedCalls, fmt.Sprint(calls))
}

func TestCallgraphScoping(t *testing.T) {
	calls := getCallsFromPackagePath("./testfiles/callgraph/scoping", t)

	expectedCalls := []resultprocessing.Call{
		{FromFile: "calls.js", FromFunction: ".root", Receiver: "this", Module: []string{}, ToFunction: "require", Arguments: []string{"foo"}},
		{FromFile: "calls.js", FromFunction: ".root", Receiver: "this", Module: []string{}, ToFunction: "require", Arguments: []string{"bar"}},
		{FromFile: "calls.js", FromFunction: ".root", Receiver: "this", Module: []string{}, ToFunction: "require", Arguments: []string{"foobar"}},
		{FromFile: "calls.js", FromFunction: ".root", Receiver: "this", Module: []string{}, ToFunction: "require", Arguments: []string{"foobar"}},
		{FromFile: "calls.js", FromFunction: "f", Receiver: "this", Module: []string{}, ToFunction: "require", Arguments: []string{"foo"}},
		{FromFile: "calls.js", FromFunction: "f", Receiver: "this", Module: []string{}, ToFunction: "require", Arguments: []string{"bar"}},
		{FromFile: "calls.js", FromFunction: "g", Receiver: "foo", Module: []string{"foobar", "foo", "bar"}, ToFile: "calls.js", ToFunction: "someMethod", Arguments: []string{}},
		{FromFile: "calls.js", FromFunction: "g", Receiver: "foobar", Module: []string{"foobar"}, ToFile: "calls.js", ToFunction: "otherMethod", Arguments: []string{}},
		{FromFile: "calls.js", FromFunction: "h", Receiver: "bar", Module: []string{"foo", "bar"}, ToFile: "calls.js", ToFunction: "someMethod", Arguments: []string{}},
		{FromFile: "calls.js", FromFunction: "h", Receiver: "fooVar", Module: []string{"foo"}, ToFile: "calls.js", ToFunction: "someMethod", Arguments: []string{}},
	}

	assert.ElementsMatch(t, calls, expectedCalls, fmt.Sprint(calls))
}

func TestCallgraphEdgecases(t *testing.T) {
	calls := getCallsFromPackagePath("./testfiles/callgraph/edgecases", t)

	expectedCalls := []resultprocessing.Call{
		{
			FromFile:     "methodChaining.js",
			FromFunction: ".root",
			Receiver:     "this",
			Module:       []string{},
			ToFile:       "methodChaining.js",
			ToFunction:   "a",
			Arguments:    []string{},
		},
		{
			FromFile:     "methodChaining.js",
			FromFunction: ".root",
			Receiver:     "a",
			Module:       []string{},
			ToFile:       "methodChaining.js",
			ToFunction:   "b",
			Arguments:    []string{},
		},
		{
			FromFile:     "methodChaining.js",
			FromFunction: ".root",
			Receiver:     "a.b",
			Module:       []string{},
			ToFile:       "methodChaining.js",
			ToFunction:   "c",
			Arguments:    []string{},
		},
		{
			FromFile:     "moduleClass.js",
			FromFunction: ".root",
			Receiver:     "this",
			Module:       []string{},
			ToFunction:   "require",
			Arguments:    []string{"oauth"},
		},
		{
			FromFile:     "moduleClass.js",
			FromFunction: ".root",
			Receiver:     "oauth",
			Module:       []string{"oauth"},
			ToFile:       "moduleClass.js",
			ToFunction:   "new OAuth",
			Arguments:    []string{"a"},
		},
		{
			FromFile:     "moduleClass.js",
			FromFunction: ".root",
			Receiver:     "oauth",
			Module:       []string{"oauth"},
			ToFile:       "moduleClass.js",
			ToFunction:   "someMethod",
			Arguments:    []string{},
		},
		{
			FromFile:     "moduleClass.js",
			FromFunction: ".root",
			Receiver:     "foo",
			Module:       []string{"oauth"},
			ToFile:       "moduleClass.js",
			ToFunction:   "new OAuth",
			Arguments:    []string{"b"},
		},
		{
			FromFile:     "moduleClass.js",
			FromFunction: ".root",
			Receiver:     "this",
			Module:       []string{},
			ToFunction:   "require",
			Arguments:    []string{"auth0"},
		},

		{
			FromFile:     "moduleClass.js",
			FromFunction: ".root",
			Receiver:     "foo",
			Module:       []string{"oauth", "auth0"},
			ToFile:       "moduleClass.js",
			ToFunction:   "someMethod",
			Arguments:    []string{},
		},
	}

	assert.ElementsMatch(t, calls, expectedCalls, fmt.Sprint(calls))
}

func getCallsFromPackagePath(packagePath string, t *testing.T) []resultprocessing.Call {
	const analysisPath = "./callgraph-analysis/analysis"
	logger := zap.NewNop().Sugar()
	analysis := NewASTAnalysis(logger, analysisPath)
	result, err := analysis.AnalyzePackage(packagePath)
	if err != nil {
		t.Fatal(err)
	}
	calls, err := transformToCalls(result)
	if err != nil {
		t.Fatal(err)
	}

	return calls
}
