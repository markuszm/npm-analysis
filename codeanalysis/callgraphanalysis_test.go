package codeanalysis

import (
	"fmt"
	"github.com/markuszm/npm-analysis/resultprocessing"
	"github.com/stretchr/testify/assert"
	"go.uber.org/zap"
	"testing"
)

func TestCallgraphLocal(t *testing.T) {
	calls := getCallsFromPackagePath("./testfiles/callgraph/local", t)

	expectedCalls := []resultprocessing.Call{
		{FromModule: "call", FromFunction: ".root", Receiver: "", Module: []string{}, ToFunction: "myfun", Arguments: []string{"2"}, IsLocal: false},
		{FromModule: "fun", FromFunction: "myfun", Receiver: "", Module: []string{}, ToFunction: "otherfun", Arguments: []string{"x"}, IsLocal: true},
		{FromModule: "fun", FromFunction: "otherfun", Receiver: "", Module: []string{}, ToFunction: "anotherfun", Arguments: []string{"y"}, IsLocal: true},
		{FromModule: "fun", FromFunction: ".root", Receiver: "", Module: []string{}, ToFunction: "myfun", Arguments: []string{"2"}, IsLocal: true},
		{FromModule: "fun", FromFunction: ".root", Receiver: "console", Module: []string{}, ToFunction: "log", Arguments: []string{"myfun(2)"}, IsLocal: false},
		{FromModule: "fun", FromFunction: ".root", Receiver: "", Module: []string{}, ToFunction: "function() {...}", Arguments: []string{}, IsLocal: false},
	}
	assert.ElementsMatch(t, calls, expectedCalls, fmt.Sprint(calls))
}

func TestCallgraphModule(t *testing.T) {
	calls := getCallsFromPackagePath("./testfiles/callgraph/modules", t)

	expectedCalls := []resultprocessing.Call{
		{FromModule: "calls", FromFunction: ".root", Receiver: "", Module: []string{}, ToFunction: "require", Arguments: []string{"foo"}},
		{FromModule: "calls", FromFunction: ".root", Receiver: "", Module: []string{}, ToFunction: "require", Arguments: []string{"bar"}},
		{FromModule: "calls", FromFunction: ".root", Receiver: "", Module: []string{}, ToFunction: "require", Arguments: []string{"foobar"}},
		{FromModule: "calls", FromFunction: ".root", Receiver: "f", Module: []string{"foo"}, ToFunction: "a", Arguments: []string{}},
		{FromModule: "calls", FromFunction: ".root", Receiver: "bar", Module: []string{"bar"}, ToFunction: "b", Arguments: []string{"a"}},
		{FromModule: "calls", FromFunction: ".root", Receiver: "foobar", Module: []string{"foobar"}, ToFunction: "func", Arguments: []string{"a", "b"}},
	}

	assert.ElementsMatch(t, calls, expectedCalls, fmt.Sprint(calls))
}

func TestCallgraphES6Module(t *testing.T) {
	calls := getCallsFromPackagePath("./testfiles/callgraph/es6modules", t)

	expectedCalls := []resultprocessing.Call{
		{FromModule: "call", FromFunction: "foo", Receiver: "_", Module: []string{"underscore"}, ToFunction: "map", Arguments: []string{"aList", "(i) => {...}"}},
		{FromModule: "call", FromFunction: "foo", Receiver: "bar", Module: []string{"foobar"}, ToFunction: "add", Arguments: []string{"i"}},
		{FromModule: "call", FromFunction: "foo", Receiver: "", Module: []string{"b"}, ToFunction: "a", Arguments: []string{"mappedList"}},
	}

	assert.ElementsMatch(t, calls, expectedCalls, fmt.Sprint(calls))
}

func TestCallgraphMix(t *testing.T) {
	calls := getCallsFromPackagePath("./testfiles/callgraph/mix", t)

	expectedCalls := []resultprocessing.Call{
		{FromModule: "anotherFile", FromFunction: "aFnInAnotherFile", Receiver: "console", Module: []string{}, ToFunction: "log", Arguments: []string{"cool"}},
		{FromModule: "file", FromFunction: ".root", Receiver: "", Module: []string{}, ToFunction: "require", Arguments: []string{"./anotherFile"}},
		{FromModule: "file", FromFunction: "aFunction", Receiver: "_", Module: []string{}, ToFunction: "curry", Arguments: []string{"libVar.referencedFn"}},
		{FromModule: "file", FromFunction: "aFunction", Receiver: "async", Module: []string{}, ToFunction: "series", Arguments: []string{"[_.curry(libVar.referencedFn)]"}},
		{FromModule: "file", FromFunction: "aFunction", Receiver: "libVar", Module: []string{"./anotherFile"}, ToFunction: "aFnInAnotherFile", Arguments: []string{"n + 1"}},
		{FromModule: "file", FromFunction: ".root", Receiver: "libVar", Module: []string{"./anotherFile"}, ToFunction: "aFnInAnotherFile", Arguments: []string{"2"}},
		{FromModule: "file", FromFunction: ".root", Receiver: "", Module: []string{}, ToFunction: "aFunction", Arguments: []string{}, IsLocal: true},
	}

	assert.ElementsMatch(t, calls, expectedCalls, fmt.Sprint(calls))
}

func TestCallgraphScoping(t *testing.T) {
	calls := getCallsFromPackagePath("./testfiles/callgraph/scoping", t)

	expectedCalls := []resultprocessing.Call{
		{FromModule: "calls", FromFunction: ".root", Receiver: "", Module: []string{}, ToFunction: "require", Arguments: []string{"foo"}},
		{FromModule: "calls", FromFunction: ".root", Receiver: "", Module: []string{}, ToFunction: "require", Arguments: []string{"bar"}},
		{FromModule: "calls", FromFunction: ".root", Receiver: "", Module: []string{}, ToFunction: "require", Arguments: []string{"foobar"}},
		{FromModule: "calls", FromFunction: ".root", Receiver: "", Module: []string{}, ToFunction: "require", Arguments: []string{"foobar"}},
		{FromModule: "calls", FromFunction: "f", Receiver: "", Module: []string{}, ToFunction: "require", Arguments: []string{"foo"}},
		{FromModule: "calls", FromFunction: "f", Receiver: "", Module: []string{}, ToFunction: "require", Arguments: []string{"bar"}},
		{FromModule: "calls", FromFunction: "g", Receiver: "foo", Module: []string{"foobar", "foo", "bar"}, ToFunction: "someMethod", Arguments: []string{}},
		{FromModule: "calls", FromFunction: "g", Receiver: "foobar", Module: []string{"foobar"}, ToFunction: "otherMethod", Arguments: []string{}},
		{FromModule: "calls", FromFunction: "h", Receiver: "bar", Module: []string{"foo", "bar"}, ToFunction: "someMethod", Arguments: []string{}},
		{FromModule: "calls", FromFunction: "h", Receiver: "fooVar", Module: []string{"foo"}, ToFunction: "someMethod", Arguments: []string{}},
	}

	assert.ElementsMatch(t, calls, expectedCalls, fmt.Sprint(calls))
}

func TestCallgraphEdgecases(t *testing.T) {
	calls := getCallsFromPackagePath("./testfiles/callgraph/edgecases", t)

	expectedCalls := []resultprocessing.Call{
		{
			FromModule:   "methodChaining",
			FromFunction: ".root",
			Receiver:     "",
			Module:       []string{},
			ToFunction:   "a",
			Arguments:    []string{},
			IsLocal:      true,
		},
		{
			FromModule:   "methodChaining",
			FromFunction: ".root",
			Receiver:     "a",
			Module:       []string{},
			ToFunction:   "b",
			Arguments:    []string{},
		},
		{
			FromModule:   "methodChaining",
			FromFunction: ".root",
			Receiver:     "a.b",
			Module:       []string{},
			ToFunction:   "c",
			Arguments:    []string{},
		},
		{
			FromModule:   "moduleClass",
			FromFunction: ".root",
			Receiver:     "",
			Module:       []string{},
			ToFunction:   "require",
			Arguments:    []string{"oauth"},
		},
		{
			FromModule:   "moduleClass",
			FromFunction: ".root",
			Receiver:     "oauth",
			Module:       []string{"oauth"},
			ToFunction:   "new OAuth",
			Arguments:    []string{"a"},
		},
		{
			FromModule:   "moduleClass",
			FromFunction: ".root",
			Receiver:     "oauth",
			Module:       []string{"oauth"},
			ToFunction:   "someMethod",
			Arguments:    []string{},
		},
		{
			FromModule:   "moduleClass",
			FromFunction: ".root",
			Receiver:     "foo",
			Module:       []string{"oauth"},
			ToFunction:   "new OAuth",
			Arguments:    []string{"b"},
		},
		{
			FromModule:   "moduleClass",
			FromFunction: ".root",
			Receiver:     "",
			Module:       []string{},
			ToFunction:   "require",
			Arguments:    []string{"auth0"},
		},

		{
			FromModule:   "moduleClass",
			FromFunction: ".root",
			Receiver:     "foo",
			Module:       []string{"oauth", "auth0"},
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
	calls, err := resultprocessing.TransformToCalls(result)
	if err != nil {
		t.Fatal(err)
	}

	return calls
}
