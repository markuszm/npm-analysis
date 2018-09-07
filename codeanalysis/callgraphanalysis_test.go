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
		{FromModule: "call", FromFunction: ".root", Receiver: "", Modules: []string{}, ToFunction: "myfun", Arguments: []string{"2"}, IsLocal: false},
		{FromModule: "fun", FromFunction: "myfun", Receiver: "", Modules: []string{}, ToFunction: "otherfun", Arguments: []string{"x"}, IsLocal: true},
		{FromModule: "fun", FromFunction: "otherfun", Receiver: "", Modules: []string{}, ToFunction: "anotherfun", Arguments: []string{"y"}, IsLocal: true},
		{FromModule: "fun", FromFunction: ".root", Receiver: "", Modules: []string{}, ToFunction: "myfun", Arguments: []string{"2"}, IsLocal: true},
		{FromModule: "fun", FromFunction: ".root", Receiver: "console", Modules: []string{}, ToFunction: "log", Arguments: []string{"myfun(2)"}, IsLocal: false},
	}
	assert.ElementsMatch(t, calls, expectedCalls, fmt.Sprint(calls))
}

func TestCallgraphModule(t *testing.T) {
	calls := getCallsFromPackagePath("./testfiles/callgraph/modules", t)

	expectedCalls := []resultprocessing.Call{
		{FromModule: "calls", FromFunction: ".root", Receiver: "", Modules: []string{}, ToFunction: "require", Arguments: []string{"foo"}},
		{FromModule: "calls", FromFunction: ".root", Receiver: "", Modules: []string{}, ToFunction: "require", Arguments: []string{"bar"}},
		{FromModule: "calls", FromFunction: ".root", Receiver: "", Modules: []string{}, ToFunction: "require", Arguments: []string{"foobar"}},
		{FromModule: "calls", FromFunction: ".root", Receiver: "f", Modules: []string{"foo"}, ToFunction: "a", Arguments: []string{}},
		{FromModule: "calls", FromFunction: ".root", Receiver: "bar", Modules: []string{"bar"}, ToFunction: "b", Arguments: []string{"a"}},
		{FromModule: "calls", FromFunction: ".root", Receiver: "foobar", Modules: []string{"foobar"}, ToFunction: "func", Arguments: []string{"a", "b"}},
	}

	assert.ElementsMatch(t, calls, expectedCalls, fmt.Sprint(calls))
}

func TestCallgraphES6Module(t *testing.T) {
	calls := getCallsFromPackagePath("./testfiles/callgraph/es6modules", t)

	expectedCalls := []resultprocessing.Call{
		{FromModule: "call", FromFunction: "foo", Receiver: "_", Modules: []string{"underscore"}, ToFunction: "map", Arguments: []string{"aList", "(i) => {...}"}},
		{FromModule: "call", FromFunction: "foo", Receiver: "bar", Modules: []string{"foobar"}, ToFunction: "add", Arguments: []string{"i"}},
		{FromModule: "call", FromFunction: "foo", Receiver: "", Modules: []string{"b"}, ToFunction: "a", Arguments: []string{"mappedList"}},
	}

	assert.ElementsMatch(t, calls, expectedCalls, fmt.Sprint(calls))
}

func TestCallgraphMix(t *testing.T) {
	calls := getCallsFromPackagePath("./testfiles/callgraph/mix", t)

	expectedCalls := []resultprocessing.Call{
		{FromModule: "anotherFile", FromFunction: "aFnInAnotherFile", Receiver: "console", Modules: []string{}, ToFunction: "log", Arguments: []string{"cool"}},
		{FromModule: "file", FromFunction: ".root", Receiver: "", Modules: []string{}, ToFunction: "require", Arguments: []string{"./anotherFile"}},
		{FromModule: "file", FromFunction: "aFunction", Receiver: "_", Modules: []string{}, ToFunction: "curry", Arguments: []string{"libVar.referencedFn"}},
		{FromModule: "file", FromFunction: "aFunction", Receiver: "async", Modules: []string{}, ToFunction: "series", Arguments: []string{"[_.curry(libVar.referencedFn)]"}},
		{FromModule: "file", FromFunction: "aFunction", Receiver: "libVar", Modules: []string{"./anotherFile"}, ToFunction: "aFnInAnotherFile", Arguments: []string{"n + 1"}},
		{FromModule: "file", FromFunction: ".root", Receiver: "libVar", Modules: []string{"./anotherFile"}, ToFunction: "aFnInAnotherFile", Arguments: []string{"2"}},
		{FromModule: "file", FromFunction: ".root", Receiver: "", Modules: []string{}, ToFunction: "aFunction", Arguments: []string{}, IsLocal: true},
	}

	assert.ElementsMatch(t, calls, expectedCalls, fmt.Sprint(calls))
}

func TestCallgraphScoping(t *testing.T) {
	calls := getCallsFromPackagePath("./testfiles/callgraph/scoping", t)

	expectedCalls := []resultprocessing.Call{
		{FromModule: "calls", FromFunction: ".root", Receiver: "", Modules: []string{}, ToFunction: "require", Arguments: []string{"foo"}},
		{FromModule: "calls", FromFunction: ".root", Receiver: "", Modules: []string{}, ToFunction: "require", Arguments: []string{"bar"}},
		{FromModule: "calls", FromFunction: ".root", Receiver: "", Modules: []string{}, ToFunction: "require", Arguments: []string{"foobar"}},
		{FromModule: "calls", FromFunction: ".root", Receiver: "", Modules: []string{}, ToFunction: "require", Arguments: []string{"foobar"}},
		{FromModule: "calls", FromFunction: "f", Receiver: "", Modules: []string{}, ToFunction: "require", Arguments: []string{"foo"}},
		{FromModule: "calls", FromFunction: "f", Receiver: "", Modules: []string{}, ToFunction: "require", Arguments: []string{"bar"}},
		{FromModule: "calls", FromFunction: "g", Receiver: "foo", Modules: []string{"foobar", "foo", "bar"}, ToFunction: "someMethod", Arguments: []string{}},
		{FromModule: "calls", FromFunction: "g", Receiver: "foobar", Modules: []string{"foobar"}, ToFunction: "otherMethod", Arguments: []string{}},
		{FromModule: "calls", FromFunction: "h", Receiver: "bar", Modules: []string{"foo", "bar"}, ToFunction: "someMethod", Arguments: []string{}},
		{FromModule: "calls", FromFunction: "h", Receiver: "fooVar", Modules: []string{"foo"}, ToFunction: "someMethod", Arguments: []string{}},
		{FromModule: "nestedCrossRef", FromFunction: ".root", Receiver: "", Modules: []string{}, ToFunction: "require", Arguments: []string{"microtime"}},
		{FromModule: "nestedCrossRef", FromFunction: "nested", Receiver: "e", Modules: []string{}, ToFunction: "normalize", Arguments: []string{}},
		{FromModule: "nestedCrossRef", FromFunction: "nested", Receiver: "d", Modules: []string{}, ToFunction: "normalize", Arguments: []string{}},
		{FromModule: "nestedCrossRef", FromFunction: "func", Receiver: "foo", Modules: []string{"microtime"}, ToFunction: "now", Arguments: []string{}},
	}

	assert.ElementsMatch(t, calls, expectedCalls, fmt.Sprint(calls))
}

func TestCallgraphEdgecases(t *testing.T) {
	calls := getCallsFromPackagePath("./testfiles/callgraph/edgecases", t)

	expectedCalls := []resultprocessing.Call{
		{
			FromModule:   "basicTypes",
			FromFunction: "f",
			Receiver:     "str",
			Modules:      []string{},
			ToFunction:   "replace",
			Arguments:    []string{"a", "b"},
		},
		{
			FromModule:   "methodChaining",
			FromFunction: ".root",
			Receiver:     "",
			Modules:      []string{},
			ToFunction:   "a",
			Arguments:    []string{},
			IsLocal:      true,
		},
		{
			FromModule:   "methodChaining",
			FromFunction: ".root",
			Receiver:     "a",
			Modules:      []string{},
			ToFunction:   "b",
			Arguments:    []string{},
		},
		{
			FromModule:   "methodChaining",
			FromFunction: ".root",
			Receiver:     "a.b",
			Modules:      []string{},
			ToFunction:   "c",
			Arguments:    []string{},
		},
		{
			FromModule:   "moduleClass",
			FromFunction: ".root",
			Receiver:     "",
			Modules:      []string{},
			ToFunction:   "require",
			Arguments:    []string{"oauth"},
		},
		{
			FromModule:   "moduleClass",
			FromFunction: ".root",
			Receiver:     "oauth",
			ClassName:    "OAuth",
			Modules:      []string{"oauth"},
			ToFunction:   "new OAuth",
			Arguments:    []string{"a"},
		},
		{
			FromModule:   "moduleClass",
			FromFunction: ".root",
			Receiver:     "oauth",
			ClassName:    "OAuth",
			Modules:      []string{"oauth"},
			ToFunction:   "someMethod",
			Arguments:    []string{},
		},
		{
			FromModule:   "moduleClass",
			FromFunction: ".root",
			Receiver:     "foo",
			ClassName:    "OAuth",
			Modules:      []string{"oauth"},
			ToFunction:   "new OAuth",
			Arguments:    []string{"b"},
		},
		{
			FromModule:   "moduleClass",
			FromFunction: ".root",
			Receiver:     "",
			Modules:      []string{},
			ToFunction:   "require",
			Arguments:    []string{"auth0"},
		},
		{
			FromModule:   "moduleClass",
			FromFunction: ".root",
			Receiver:     "foo",
			ClassName:    "OAuth",
			Modules:      []string{"oauth", "auth0"},
			ToFunction:   "someMethod",
			Arguments:    []string{},
		},
		{
			FromModule:   "importedNames",
			FromFunction: ".root",
			Receiver:     "",
			Modules:      []string{},
			ToFunction:   "require",
			Arguments:    []string{"bar"},
		},
		{
			FromModule:   "importedNames",
			FromFunction: ".root",
			Receiver:     "",
			Modules:      []string{},
			ToFunction:   "require",
			Arguments:    []string{"foobar"},
		},
		{
			FromModule:   "importedNames",
			FromFunction: "f",
			Receiver:     "",
			Modules:      []string{"foo"},
			ToFunction:   "foo",
			Arguments:    []string{},
		},
		{
			FromModule:   "importedNames",
			FromFunction: "f",
			Receiver:     "",
			Modules:      []string{"bar"},
			ToFunction:   "bar",
			Arguments:    []string{},
		},
		{
			FromModule:   "importedNames",
			FromFunction: "f",
			Receiver:     "",
			Modules:      []string{"foobar"},
			ToFunction:   "bar",
			Arguments:    []string{},
		},
		{
			FromModule:   "dynamicAccess",
			FromFunction: ".root",
			Receiver:     "console",
			Modules:      []string{},
			ToFunction:   "log",
			Arguments:    []string{},
		},
		{
			FromModule:   "redefinitions",
			FromFunction: ".root",
			Receiver:     "",
			Modules:      []string{},
			ToFunction:   "require",
			Arguments:    []string{"foo"},
		},
		{
			FromModule:   "redefinitions",
			FromFunction: ".root",
			Receiver:     "",
			Modules:      []string{"foo"},
			ToFunction:   "bar",
			Arguments:    []string{},
		},
		{
			FromModule:   "redefinitions",
			FromFunction: "x",
			Receiver:     "bar2",
			Modules:      []string{"foo"},
			ToFunction:   "api",
			Arguments:    []string{},
		},
		{
			FromModule:   "redefinitions",
			FromFunction: "x",
			Receiver:     "bar3",
			Modules:      []string{"foo"},
			ToFunction:   "api",
			Arguments:    []string{},
		},
		{
			FromModule:   "regexprs",
			FromFunction: "f",
			Receiver:     "regex",
			ClassName:    "RegExp",
			Modules:      []string{},
			ToFunction:   "new RegExp",
			Arguments:    []string{"(ab)*"},
		},
		{
			FromModule:   "regexprs",
			FromFunction: "f",
			Receiver:     "regex",
			ClassName:    "RegExp",
			Modules:      []string{},
			ToFunction:   "exec",
			Arguments:    []string{"abababab"},
		},
		{
			FromModule:   "regexprs",
			FromFunction: "f",
			Receiver:     "regex2",
			ClassName:    "RegExp",
			Modules:      []string{},
			ToFunction:   "exec",
			Arguments:    []string{"ababab"},
		},
		{
			FromModule:   "regexprs",
			FromFunction: "f",
			Receiver:     "regex3",
			ClassName:    "RegExp",
			Modules:      []string{},
			ToFunction:   "exec",
			Arguments:    []string{"ab"},
		},
		{
			FromModule:   "regexprs",
			FromFunction: "f",
			Receiver:     "",
			ClassName:    "RegExp",
			Modules:      []string{},
			ToFunction:   "exec",
			Arguments:    []string{"abab"},
		},
		{
			FromModule:   "callinexported",
			FromFunction: ".root",
			Receiver:     "",
			ClassName:    "",
			Modules:      []string{},
			ToFunction:   "require",
			Arguments:    []string{"mem"},
			IsLocal:      false,
		},
		{
			FromModule:   "callinexported",
			FromFunction: ".root",
			Receiver:     "",
			ClassName:    "",
			Modules:      []string{},
			ToFunction:   "require",
			Arguments:    []string{"foo"},
			IsLocal:      false,
		},
		{
			FromModule:   "callinexported",
			FromFunction: "sync",
			Receiver:     "foo",
			ClassName:    "",
			Modules:      []string{"foo"},
			ToFunction:   "apiA",
			Arguments:    []string{},
			IsLocal:      false,
		},
		{
			FromModule:   "callinexported",
			FromFunction: "load",
			Receiver:     "foo",
			ClassName:    "",
			Modules:      []string{"foo"},
			ToFunction:   "apiB",
			Arguments:    []string{},
			IsLocal:      false,
		},
		{
			FromModule:   "callinexported",
			FromFunction: "save",
			Receiver:     "foo",
			ClassName:    "",
			Modules:      []string{"foo"},
			ToFunction:   "apiC",
			Arguments:    []string{},
			IsLocal:      false,
		},
		{
			FromModule:   "callinexported",
			FromFunction: "save",
			Receiver:     "",
			ClassName:    "",
			Modules:      []string{"mem"},
			ToFunction:   "mem",
			Arguments:    []string{"() => {...}"},
			IsLocal:      false,
		},
		{
			FromModule:   "callinexported",
			FromFunction: ".root",
			Receiver:     "",
			ClassName:    "",
			Modules:      []string{},
			ToFunction:   "exec",
			Arguments:    []string{},
			IsLocal:      false,
		},
		{
			FromModule:   "callinexported",
			FromFunction: ".root",
			Receiver:     "",
			ClassName:    "",
			Modules:      []string{"mem"},
			ToFunction:   "mem",
			Arguments:    []string{"exec()"},
			IsLocal:      false,
		},
		{
			FromModule:   "callinexported",
			FromFunction: "save",
			Receiver:     "foo",
			ClassName:    "",
			Modules:      []string{"foo"},
			ToFunction:   "apiB",
			Arguments:    []string{},
			IsLocal:      false,
		},
		{
			FromModule:   "callinexported",
			FromFunction: "loadNested",
			Receiver:     "foo",
			ClassName:    "",
			Modules:      []string{"foo"},
			ToFunction:   "apiC",
			Arguments:    []string{},
			IsLocal:      false,
		},
		{
			FromModule:   "callinexported",
			FromFunction: "load",
			Receiver:     "foo",
			ClassName:    "",
			Modules:      []string{"foo"},
			ToFunction:   "apiC",
			Arguments:    []string{},
			IsLocal:      false,
		},
		{
			FromModule:   "callinexported",
			FromFunction: "safe4",
			Receiver:     "",
			ClassName:    "",
			Modules:      []string{},
			ToFunction:   "save",
			Arguments:    []string{"v4"},
			IsLocal:      true,
		},
		{
			FromModule:   "callinexported",
			FromFunction: "load4",
			Receiver:     "",
			ClassName:    "",
			Modules:      []string{},
			ToFunction:   "load",
			Arguments:    []string{"v4"},
			IsLocal:      true,
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
