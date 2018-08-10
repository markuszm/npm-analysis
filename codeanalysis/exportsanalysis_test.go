package codeanalysis

import (
	"encoding/json"
	"fmt"
	"github.com/markuszm/npm-analysis/resultprocessing"
	"github.com/stretchr/testify/assert"
	"go.uber.org/zap"
	"testing"
)

func transformToExports(result interface{}) ([]resultprocessing.Export, error) {
	objs := result.([]interface{})

	var exports []resultprocessing.Export

	for _, value := range objs {
		export := resultprocessing.Export{}
		bytes, err := json.Marshal(value)
		if err != nil {
			return exports, err
		}
		err = json.Unmarshal(bytes, &export)
		if err != nil {
			return exports, err
		}
		exports = append(exports, export)
	}
	return exports, nil
}

func TestExportCommonJS(t *testing.T) {
	const analysisPath = "./exports-analysis/analysis"

	logger := zap.NewNop().Sugar()
	analysis := NewASTAnalysis(logger, analysisPath)
	result, err := analysis.AnalyzePackage("./testfiles/export/exportcjstest")

	if err != nil {
		t.Fatal(err)
	}

	exports, err := transformToExports(result)
	if err != nil {
		t.Fatal(err)
	}

	expectedExports := []resultprocessing.Export{
		{"class", "Calculator", []string{}, "commonjs", "classExport.js", true, "Calculator"},
		{"function", "add", []string{"a", "b"}, "commonjs", "classExport.js", true, "Calculator.add"},
		{"function", "subtract", []string{"a", "b"}, "commonjs", "classExport.js", true, "Calculator.subtract"},
		{"class", "calc", []string{}, "commonjs", "classExport.js", false, "Calculator2"},
		{"function", "calc.add", []string{"a", "b"}, "commonjs", "classExport.js", false, "Calculator2.add"},
		{"function", "calc.subtract", []string{"a", "b"}, "commonjs", "classExport.js", false, "Calculator2.subtract"},
		{"class", "AdvancedCalculator", []string{}, "commonjs", "classInstanceExport.js", true, ""},
		{"function", "AdvancedCalculator.multiply", []string{"a", "b"}, "commonjs", "classInstanceExport.js", true, ""},
		{"function", "AdvancedCalculator.divide", []string{"a", "b"}, "commonjs", "classInstanceExport.js", true, ""},
		{"function", "Calculator.add", []string{"a", "b"}, "commonjs", "classInstanceExport.js", true, ""},
		{"function", "Calculator.subtract", []string{"a", "b"}, "commonjs", "classInstanceExport.js", true, ""},
		{"function", "Base.toString", []string{"str"}, "commonjs", "classInstanceExport.js", true, ""},
		{"const", "max", []string{}, "commonjs", "constantExport.js", false, "max"},
		{"member", "min", []string{}, "commonjs", "constantExport.js", false, "Number.MIN_VALUE"},
		{"function", "addThree", []string{"a", "b", "c"}, "commonjs", "directExport.js", true, "addThree"},
		{"unknown", "foo.subThree", []string{}, "commonjs", "directExport.js", true, ""},
		{"function", "add", []string{"a", "b"}, "commonjs", "indirectExport.js", false, "add"},
		{"function", "divide", []string{"a", "b"}, "commonjs", "indirectExport.js", false, ""},
		{"function", "multiply", []string{"a", "b"}, "commonjs", "indirectExport.js", false, ""},
		{"object", "foo", []string{}, "commonjs", "indirectExport.js", false, ""},
		{"function", "foo.add2", []string{"a"}, "commonjs", "indirectExport.js", false, ""},
		{"function", "foo.add4", []string{"a"}, "commonjs", "indirectExport.js", false, ""},
		{"function", "foo.add8", []string{"a"}, "commonjs", "indirectExport.js", false, ""},
		{"function", "foo.add16", []string{"a"}, "commonjs", "indirectExport.js", false, ""},
		{"function", "foo.add32", []string{"a"}, "commonjs", "indirectExport.js", false, ""},
		{"member", "e", []string{}, "commonjs", "indirectExport.js", false, "Math.E"},
		{"class", "calculator", []string{}, "commonjs", "indirectExport.js", false, "Calculator"},
		{"function", "calculator.add", []string{"a", "b"}, "commonjs", "indirectExport.js", false, "Calculator.add"},
		{"function", "calculator.subtract", []string{"a", "b"}, "commonjs", "indirectExport.js", false, "Calculator.subtract"},
		{"function", "abs", []string{"x"}, "commonjs", "methodExport.js", false, "abs"},
		{"function", "sqrt", []string{"x"}, "commonjs", "methodExport.js", false, "sqrtDefault"},
		{"function", "pow", []string{"x", "exp"}, "commonjs", "methodExport.js", false, ""},
		{"function", "floor", []string{"x"}, "commonjs", "methodExport.js", false, ""},
		{"function", "parseInt", []string{"x", "r"}, "commonjs", "methodExport.js", false, ""},
		{"function", "add2", []string{"a"}, "commonjs", "objectExport.js", true, ""},
		{"function", "add4", []string{"a"}, "commonjs", "objectExport.js", true, ""},
		{"function", "add8", []string{"a"}, "commonjs", "objectExport.js", true, ""},
		{"function", "add16", []string{"a"}, "commonjs", "objectExport.js", true, ""},
		{"function", "add32", []string{"a"}, "commonjs", "objectExport.js", true, ""},
		{"function", "addAll", []string{"a", "...b"}, "commonjs", "objectExport.js", true, ""},
		{"member", "theSolution", []string{}, "commonjs", "objectExport.js", true, ""},
		{"var", "foo", []string{}, "commonjs", "objectExport.js", false, "bar"},
		{"function", "foo.sub2", []string{"a"}, "commonjs", "objectExport.js", false, "bar.sub2"},
		{"function", "foo.sub4", []string{"a"}, "commonjs", "objectExport.js", false, "bar.sub4"},
		{"function", "foo.sub8", []string{"a"}, "commonjs", "objectExport.js", false, "bar.sub8"},
		{"function", "foo.sub16", []string{"a"}, "commonjs", "objectExport.js", false, "bar.sub16"},
		{"var", "foo.theSolution", []string{}, "commonjs", "objectExport.js", false, "bar.theSolution"},
	}
	assert.ElementsMatch(t, exports, expectedExports, fmt.Sprint(exports))
}

func TestExportES6(t *testing.T) {
	const analysisPath = "./exports-analysis/analysis"

	logger := zap.NewNop().Sugar()
	analysis := NewASTAnalysis(logger, analysisPath)
	result, err := analysis.AnalyzePackage("./testfiles/export/exportes6test")

	if err != nil {
		t.Fatal(err)
	}

	exports, err := transformToExports(result)
	if err != nil {
		t.Fatal(err)
	}

	expectedExports := []resultprocessing.Export{
		{"function", "abs", []string{}, "es6", "defaultExport.js", true, ""},
		{"class", "Calculator", []string{}, "es6", "exportNamedClass.js", false, ""},
		{"function", "Calculator.add", []string{"a", "b"}, "es6", "exportNamedClass.js", false, ""},
		{"function", "Calculator.subtract", []string{"a", "b"}, "es6", "exportNamedClass.js", false, ""},
		{"var", "foo", []string{}, "es6", "exportOtherName.js", false, "f"},
		{"function", "bar", []string{}, "es6", "exportOtherName.js", false, "b"},
		{"function", "default function", []string{"obj"}, "es6", "mixExport.js", true, ""},
		{"function", "each", []string{"obj", "iterator", "context"}, "es6", "mixExport.js", false, ""},
		{"function", "forEach", []string{"obj", "iterator", "context"}, "es6", "mixExport.js", false, "each"},
		{"function", "cube", []string{"x"}, "es6", "namedExport.js", false, "cube"},
		{"const", "foo", []string{}, "es6", "namedExport.js", false, "foo"},
		{"var", "graph.options", []string{}, "es6", "namedExport.js", false, "graph.options"},
		{"function", "graph.draw", []string{}, "es6", "namedExport.js", false, "graph.draw"},
		{"let", "graph", []string{}, "es6", "namedExport.js", false, "graph"},
		{"const", "sqrt", []string{}, "es6", "namedExportDirect.js", false, ""},
		{"function", "square", []string{"x"}, "es6", "namedExportDirect.js", false, ""},
		{"function", "diag", []string{"x", "y"}, "es6", "namedExportDirect.js", false, ""},
		{"unknown", "default", []string{}, "es6", "redirectExport.js", false, "default"},
		{"all", "./other-module", []string{}, "es6", "redirectExport.js", false, ""},
	}
	assert.ElementsMatch(t, exports, expectedExports, fmt.Sprint(exports))
}

func TestExportScoping(t *testing.T) {
	const analysisPath = "./exports-analysis/analysis"

	logger := zap.NewNop().Sugar()
	analysis := NewASTAnalysis(logger, analysisPath)
	result, err := analysis.AnalyzePackage("./testfiles/export/scoping")

	if err != nil {
		t.Fatal(err)
	}

	exports, err := transformToExports(result)
	if err != nil {
		t.Fatal(err)
	}

	expectedExports := []resultprocessing.Export{
		{"function", "foo", []string{}, "commonjs", "scopingtest.js", false, "foo"},
		{"const", "bar", []string{}, "commonjs", "scopingtest.js", false, "foo"},
		{"var", "foobar", []string{}, "commonjs", "scopingtest.js", false, "bar"},
	}
	assert.ElementsMatch(t, exports, expectedExports, fmt.Sprint(exports))
}
