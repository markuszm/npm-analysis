package codeanalysis

import (
	"encoding/json"
	"fmt"
	"github.com/markuszm/npm-analysis/resultprocessing"
	"github.com/stretchr/testify/assert"
	"go.uber.org/zap"
	"testing"
)

func transformToImports(result interface{}) ([]resultprocessing.Import, error) {
	objs := result.([]interface{})

	var imports []resultprocessing.Import

	for _, value := range objs {
		importObj := resultprocessing.Import{}
		bytes, err := json.Marshal(value)
		if err != nil {
			return imports, err
		}
		err = json.Unmarshal(bytes, &importObj)
		if err != nil {
			return imports, err
		}
		imports = append(imports, importObj)
	}
	return imports, nil
}

func TestImportRequire(t *testing.T) {
	imports := getImportsFromPackagePath("./testfiles/import/requiretest", t)

	expectedImports := []resultprocessing.Import{
		{
			"foo",
			"main",
			"foo",
			"commonjs",
			"",
		},
		{
			"bar",
			"main",
			"bar",
			"commonjs",
			"",
		},
		{
			"OAuth",
			"main",
			"oauth",
			"commonjs",
			"OAuth",
		},
		{
			"OAuthC",
			"main",
			"oauth",
			"commonjs",
			"OAuth.a.b.c",
		},
		{
			"someNotInstalledModule",
			"main",
			"abc",
			"commonjs",
			"",
		},
		{
			"ignoreLocalModules2",
			"main",
			"/someotherjsfile.js",
			"commonjs",
			"",
		},
		{
			"ignoreLocalModules1",
			"main",
			"./someotherjsfile.js",
			"commonjs",
			"",
		},
		{
			"ignoreLocalModules3",
			"main",
			"../someotherjsfile.js",
			"commonjs",
			"",
		},
	}

	assert.ElementsMatch(t, imports, expectedImports, fmt.Sprint(imports))
}

func TestImportRequireMinified(t *testing.T) {
	imports := getImportsFromPackagePath("./testfiles/import/requireminifiedtest", t)

	expectedImports := []resultprocessing.Import{
		{
			"foo",
			"main",
			"foo",
			"commonjs",
			"",
		},
		{
			"bar",
			"main",
			"bar",
			"commonjs",
			"",
		},
		{
			"someNotInstalledModule",
			"main",
			"abc",
			"commonjs",
			"",
		},
		{
			"ignoreLocalModules2",
			"main",
			"/someotherjsfile.js",
			"commonjs",
			"",
		},
		{
			"ignoreLocalModules1",
			"main",
			"./someotherjsfile.js",
			"commonjs",
			"",
		},
		{
			"ignoreLocalModules3",
			"main",
			"../someotherjsfile.js",
			"commonjs",
			"",
		},
	}

	assert.ElementsMatch(t, imports, expectedImports, fmt.Sprint(imports))
}

func TestImportES6(t *testing.T) {
	imports := getImportsFromPackagePath("./testfiles/import/importtest", t)

	expectedImports := []resultprocessing.Import{
		{
			"foo",
			"main",
			"foo",
			"es6",
			"",
		},
		{
			"bar",
			"main",
			"bar",
			"es6",
			"",
		},
		{
			"a",
			"main",
			"a",
			"es6",
			"a",
		},
		{
			"b",
			"main",
			"b",
			"es6",
			"e",
		},
		{
			"c1",
			"main",
			"c",
			"es6",
			"c1",
		},
		{
			"c2",
			"main",
			"c",
			"es6",
			"c2",
		},
		{
			"e1",
			"main",
			"d",
			"es6",
			"e1",
		},
		{
			"d1",
			"main",
			"d",
			"es6",
			"e2",
		},
		{
			"d2",
			"main",
			"d",
			"es6",
			"d2",
		},
		{
			"f1",
			"main",
			"e",
			"es6",
			"",
		},
		{
			"f",
			"main",
			"e",
			"es6",
			"",
		},
		{
			"@side-effect",
			"main",
			"f",
			"es6",
			"",
		},
		{
			"@side-effect",
			"main",
			"g",
			"es6",
			"",
		},
		{
			"@side-effect",
			"main",
			"/f",
			"es6",
			"",
		},
		{
			"@side-effect",
			"main",
			"./f",
			"es6",
			"",
		},
		{
			"@side-effect",
			"main",
			"../f",
			"es6",
			"",
		},
		{
			"foo1",
			"main",
			"/foo",
			"es6",
			"",
		},
		{
			"foo2",
			"main",
			"./foo",
			"es6",
			"",
		},
		{
			"foo3",
			"main",
			"../foo",
			"es6",
			"",
		},
	}

	assert.ElementsMatch(t, imports, expectedImports, fmt.Sprint(imports))
}

func TestImportES6Minified(t *testing.T) {
	imports := getImportsFromPackagePath("./testfiles/import/importminifiedtest", t)

	expectedImports := []resultprocessing.Import{
		{
			"foo",
			"main",
			"foo",
			"es6",
			"",
		},
		{
			"bar",
			"main",
			"bar",
			"es6",
			"",
		},
		{
			"a",
			"main",
			"a",
			"es6",
			"a",
		},
		{
			"b",
			"main",
			"b",
			"es6",
			"e",
		},
		{
			"c1",
			"main",
			"c",
			"es6",
			"c1",
		},
		{
			"c2",
			"main",
			"c",
			"es6",
			"c2",
		},
		{
			"e1",
			"main",
			"d",
			"es6",
			"e1",
		},
		{
			"d1",
			"main",
			"d",
			"es6",
			"e2",
		},
		{
			"d2",
			"main",
			"d",
			"es6",
			"d2",
		},
		{
			"f1",
			"main",
			"e",
			"es6",
			"",
		},
		{
			"f",
			"main",
			"e",
			"es6",
			"",
		},
		{
			"@side-effect",
			"main",
			"f",
			"es6",
			"",
		},
		{
			"@side-effect",
			"main",
			"g",
			"es6",
			"",
		},
		{
			"@side-effect",
			"main",
			"/f",
			"es6",
			"",
		},
		{
			"@side-effect",
			"main",
			"./f",
			"es6",
			"",
		},
		{
			"@side-effect",
			"main",
			"../f",
			"es6",
			"",
		},
		{
			"foo1",
			"main",
			"/foo",
			"es6",
			"",
		},
		{
			"foo2",
			"main",
			"./foo",
			"es6",
			"",
		},
		{
			"foo3",
			"main",
			"../foo",
			"es6",
			"",
		},
	}

	assert.ElementsMatch(t, imports, expectedImports, fmt.Sprint(imports))
}

func TestImportTypescript(t *testing.T) {
	imports := getImportsFromPackagePath("./testfiles/import/typescripttest", t)

	expectedImports := []resultprocessing.Import{
		{
			"bar",
			"somets",
			"bar",
			"es6",
			"bar",
		},
		{
			"$",
			"somets",
			"JQuery",
			"es6",
			"",
		},
		{
			"@side-effect",
			"somets",
			"abc",
			"es6",
			"",
		},
		{
			"ZCV",
			"somets",
			"./ZipCodeValidator",
			"es6",
			"ZipCodeValidator",
		}}

	assert.ElementsMatch(t, imports, expectedImports, fmt.Sprint(imports))
}

func TestImportReassignments(t *testing.T) {
	imports := getImportsFromPackagePath("./testfiles/callgraph/scoping/calls.js", t)

	expectedImports := []resultprocessing.Import{
		{
			Identifier: "foo",
			FromModule: "./testfiles/callgraph/scoping/calls",
			ModuleName: "foobar",
			BundleType: "commonjs",
		},
		{
			Identifier: "foo",
			FromModule: "./testfiles/callgraph/scoping/calls",
			ModuleName: "foo",
			BundleType: "commonjs",
		},
		{
			Identifier: "bar",
			FromModule: "./testfiles/callgraph/scoping/calls",
			ModuleName: "bar",
			BundleType: "commonjs",
		},
		{
			Identifier: "foobar",
			FromModule: "./testfiles/callgraph/scoping/calls",
			ModuleName: "foobar",
			BundleType: "commonjs",
		},
		{
			Identifier: "foobar",
			FromModule: "./testfiles/callgraph/scoping/calls",
			ModuleName: "foo",
			BundleType: "commonjs",
		},
		{
			Identifier: "foo",
			FromModule: "./testfiles/callgraph/scoping/calls",
			ModuleName: "bar",
			BundleType: "commonjs",
		},
	}

	assert.ElementsMatch(t, imports, expectedImports, fmt.Sprint(imports))
}

func getImportsFromPackagePath(packagePath string, t *testing.T) []resultprocessing.Import {
	const analysisPath = "./import-analysis/analysis"
	logger := zap.NewNop().Sugar()
	analysis := NewASTAnalysis(logger, analysisPath)
	result, err := analysis.AnalyzePackage(packagePath)
	if err != nil {
		t.Fatal(err)
	}
	imports, err := transformToImports(result)
	if err != nil {
		t.Fatal(err)
	}

	return imports
}
