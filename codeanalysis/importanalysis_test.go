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
	const analysisPath = "./import-analysis/analysis"

	logger := zap.NewNop().Sugar()
	analysis := NewASTAnalysis(logger, analysisPath)
	result, err := analysis.AnalyzePackage("./testfiles/import/requiretest")

	if err != nil {
		t.Fatal(err)
	}

	imports, err := transformToImports(result)
	if err != nil {
		t.Fatal(err)
	}

	expectedImports := []resultprocessing.Import{
		{
			"foo",
			"foo",
			"commonjs",
			"",
		},
		{
			"bar",
			"bar",
			"commonjs",
			"",
		},
		{
			"OAuth",
			"oauth",
			"commonjs",
			"OAuth",
		},
		{
			"OAuthC",
			"oauth",
			"commonjs",
			"OAuth.a.b.c",
		},
		{
			"someNotInstalledModule",
			"abc",
			"commonjs",
			"",
		},
		{
			"ignoreLocalModules2",
			"/someotherjsfile.js",
			"commonjs",
			"",
		},
		{
			"ignoreLocalModules1",
			"./someotherjsfile.js",
			"commonjs",
			"",
		},
		{
			"ignoreLocalModules3",
			"../someotherjsfile.js",
			"commonjs",
			"",
		},
	}

	assert.ElementsMatch(t, imports, expectedImports, fmt.Sprint(imports))
}

func TestImportRequireMinified(t *testing.T) {
	const analysisPath = "./import-analysis/analysis"

	logger := zap.NewNop().Sugar()
	analysis := NewASTAnalysis(logger, analysisPath)
	result, err := analysis.AnalyzePackage("./testfiles/import/requireminifiedtest")

	if err != nil {
		t.Fatal(err)
	}

	imports, err := transformToImports(result)
	if err != nil {
		t.Fatal(err)
	}

	expectedImports := []resultprocessing.Import{
		{
			"foo",
			"foo",
			"commonjs",
			"",
		},
		{
			"bar",
			"bar",
			"commonjs",
			"",
		},
		{
			"someNotInstalledModule",
			"abc",
			"commonjs",
			"",
		},
		{
			"ignoreLocalModules2",
			"/someotherjsfile.js",
			"commonjs",
			"",
		},
		{
			"ignoreLocalModules1",
			"./someotherjsfile.js",
			"commonjs",
			"",
		},
		{
			"ignoreLocalModules3",
			"../someotherjsfile.js",
			"commonjs",
			"",
		},
	}

	assert.ElementsMatch(t, imports, expectedImports, fmt.Sprint(imports))
}

func TestImportES6(t *testing.T) {
	const analysisPath = "./import-analysis/analysis"

	logger := zap.NewNop().Sugar()
	analysis := NewASTAnalysis(logger, analysisPath)
	result, err := analysis.AnalyzePackage("./testfiles/import/importtest")

	if err != nil {
		t.Fatal(err)
	}

	imports, err := transformToImports(result)
	if err != nil {
		t.Fatal(err)
	}

	expectedImports := []resultprocessing.Import{
		{
			"foo",
			"foo",
			"es6",
			"",
		},
		{
			"bar",
			"bar",
			"es6",
			"",
		},
		{
			"a",
			"a",
			"es6",
			"a",
		},
		{
			"b",
			"b",
			"es6",
			"e",
		},
		{
			"c1",
			"c",
			"es6",
			"c1",
		},
		{
			"c2",
			"c",
			"es6",
			"c2",
		},
		{
			"e1",
			"d",
			"es6",
			"e1",
		},
		{
			"d1",
			"d",
			"es6",
			"e2",
		},
		{
			"d2",
			"d",
			"es6",
			"d2",
		},
		{
			"f1",
			"e",
			"es6",
			"",
		},
		{
			"f",
			"e",
			"es6",
			"",
		},
		{
			"@side-effect",
			"f",
			"es6",
			"",
		},
		{
			"@side-effect",
			"g",
			"es6",
			"",
		},
		{
			"@side-effect",
			"/f",
			"es6",
			"",
		},
		{
			"@side-effect",
			"./f",
			"es6",
			"",
		},
		{
			"@side-effect",
			"../f",
			"es6",
			"",
		},
		{
			"foo1",
			"/foo",
			"es6",
			"",
		},
		{
			"foo2",
			"./foo",
			"es6",
			"",
		},
		{
			"foo3",
			"../foo",
			"es6",
			"",
		},
	}

	assert.ElementsMatch(t, imports, expectedImports, fmt.Sprint(imports))
}

func TestImportES6Minified(t *testing.T) {
	const analysisPath = "./import-analysis/analysis"

	logger := zap.NewNop().Sugar()
	analysis := NewASTAnalysis(logger, analysisPath)
	result, err := analysis.AnalyzePackage("./testfiles/import/importminifiedtest")

	if err != nil {
		t.Fatal(err)
	}

	imports, err := transformToImports(result)
	if err != nil {
		t.Fatal(err)
	}

	expectedImports := []resultprocessing.Import{
		{
			"foo",
			"foo",
			"es6",
			"",
		},
		{
			"bar",
			"bar",
			"es6",
			"",
		},
		{
			"a",
			"a",
			"es6",
			"a",
		},
		{
			"b",
			"b",
			"es6",
			"e",
		},
		{
			"c1",
			"c",
			"es6",
			"c1",
		},
		{
			"c2",
			"c",
			"es6",
			"c2",
		},
		{
			"e1",
			"d",
			"es6",
			"e1",
		},
		{
			"d1",
			"d",
			"es6",
			"e2",
		},
		{
			"d2",
			"d",
			"es6",
			"d2",
		},
		{
			"f1",
			"e",
			"es6",
			"",
		},
		{
			"f",
			"e",
			"es6",
			"",
		},
		{
			"@side-effect",
			"f",
			"es6",
			"",
		},
		{
			"@side-effect",
			"g",
			"es6",
			"",
		},
		{
			"@side-effect",
			"/f",
			"es6",
			"",
		},
		{
			"@side-effect",
			"./f",
			"es6",
			"",
		},
		{
			"@side-effect",
			"../f",
			"es6",
			"",
		},
		{
			"foo1",
			"/foo",
			"es6",
			"",
		},
		{
			"foo2",
			"./foo",
			"es6",
			"",
		},
		{
			"foo3",
			"../foo",
			"es6",
			"",
		},
	}

	assert.ElementsMatch(t, imports, expectedImports, fmt.Sprint(imports))
}

func TestImportTypescript(t *testing.T) {
	const analysisPath = "./import-analysis/analysis"

	logger := zap.NewNop().Sugar()
	analysis := NewASTAnalysis(logger, analysisPath)
	result, err := analysis.AnalyzePackage("./testfiles/import/typescripttest")

	if err != nil {
		t.Fatal(err)
	}

	imports, err := transformToImports(result)
	if err != nil {
		t.Fatal(err)
	}

	expectedImports := []resultprocessing.Import{
		{
			"bar",
			"bar",
			"es6",
			"bar",
		},
		{
			"$",
			"JQuery",
			"es6",
			"",
		},
		{
			"@side-effect",
			"abc",
			"es6",
			"",
		},
		{
			"ZCV",
			"./ZipCodeValidator",
			"es6",
			"ZipCodeValidator",
		}}

	assert.ElementsMatch(t, imports, expectedImports, fmt.Sprint(imports))
}
