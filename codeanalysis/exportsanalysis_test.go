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
		{"class", "Calculator", "commonjs", "classExport.js", true},
		{"function", "add(a,b)", "commonjs", "classExport.js", true},
		{"function", "subtract(a,b)", "commonjs", "classExport.js", true},
		{"class", "calc", "commonjs", "classExport.js", false},
		{"function", "calc.add(a,b)", "commonjs", "classExport.js", false},
		{"function", "calc.subtract(a,b)", "commonjs", "classExport.js", false},
		{"class", "AdvancedCalculator", "commonjs", "classInstanceExport.js", true},
		{"function", "AdvancedCalculator.multiply(a,b)", "commonjs", "classInstanceExport.js", true},
		{"function", "AdvancedCalculator.divide(a,b)", "commonjs", "classInstanceExport.js", true},
		{"function", "Calculator.add(a,b)", "commonjs", "classInstanceExport.js", true},
		{"function", "Calculator.substract(a,b)", "commonjs", "classInstanceExport.js", true},
		{"function", "Base.toString(str)", "commonjs", "classInstanceExport.js", true},
		{"const", "max", "commonjs", "constantExport.js", false},
		{"member", "min", "commonjs", "constantExport.js", false},
		{"function", "addThree(a,b,c)", "commonjs", "directExport.js", true},
		{"unknown", "foo.subThree()", "commonjs", "directExport.js", true},
		{"function", "add(a,b)", "commonjs", "indirectExport.js", false},
		{"function", "divide(a,b)", "commonjs", "indirectExport.js", false},
		{"function", "multiply(a,b)", "commonjs", "indirectExport.js", false},
		{"object", "foo", "commonjs", "indirectExport.js", false},
		{"function", "foo.add2(a)", "commonjs", "indirectExport.js", false},
		{"function", "foo.add4(a)", "commonjs", "indirectExport.js", false},
		{"function", "foo.add8(a)", "commonjs", "indirectExport.js", false},
		{"function", "foo.add16(a)", "commonjs", "indirectExport.js", false},
		{"function", "foo.add32(a)", "commonjs", "indirectExport.js", false},
		{"member", "e", "commonjs", "indirectExport.js", false},
		{"class", "calculator", "commonjs", "indirectExport.js", false},
		{"function", "calculator.add(a,b)", "commonjs", "indirectExport.js", false},
		{"function", "calculator.substract(a,b)", "commonjs", "indirectExport.js", false},
		{"function", "abs(x)", "commonjs", "methodExport.js", false},
		{"function", "sqrt(x)", "commonjs", "methodExport.js", false},
		{"function", "pow(x,exp)", "commonjs", "methodExport.js", false},
		{"function", "floor(x)", "commonjs", "methodExport.js", false},
		{"function", "parseInt(x,r)", "commonjs", "methodExport.js", false},
		{"function", "add2(a)", "commonjs", "objectExport.js", true},
		{"function", "add4(a)", "commonjs", "objectExport.js", true},
		{"function", "add8(a)", "commonjs", "objectExport.js", true},
		{"function", "add16(a)", "commonjs", "objectExport.js", true},
		{"function", "add32(a)", "commonjs", "objectExport.js", true},
		{"function", "addAll(a,...b)", "commonjs", "objectExport.js", true},
		{"member", "theSolution", "commonjs", "objectExport.js", true},
		{"var", "foo", "commonjs", "objectExport.js", false},
		{"function", "foo.sub2(a)", "commonjs", "objectExport.js", false},
		{"function", "foo.sub4(a)", "commonjs", "objectExport.js", false},
		{"function", "foo.sub8(a)", "commonjs", "objectExport.js", false},
		{"function", "foo.sub16(a)", "commonjs", "objectExport.js", false},
		{"var", "foo.theSolution", "commonjs", "objectExport.js", false},
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
		{"function", "abs()", "es6", "defaultExport.js", true},
		{"class", "Calculator", "es6", "exportNamedClass.js", false},
		{"function", "Calculator.add(a,b)", "es6", "exportNamedClass.js", false},
		{"function", "Calculator.substract(a,b)", "es6", "exportNamedClass.js", false},
		{"function", "default function(obj)", "es6", "mixExport.js", true},
		{"function", "each(obj,iterator,context)", "es6", "mixExport.js", false},
		{"unknown", "forEach", "es6", "mixExport.js", false},
		{"function", "cube(x)", "es6", "namedExport.js", false},
		{"const", "foo", "es6", "namedExport.js", false},
		{"var", "graph.options", "es6", "namedExport.js", false},
		{"function", "graph.draw()", "es6", "namedExport.js", false},
		{"let", "graph", "es6", "namedExport.js", false},
		{"const", "sqrt", "es6", "namedExportDirect.js", false},
		{"function", "square(x)", "es6", "namedExportDirect.js", false},
		{"function", "diag(x,y)", "es6", "namedExportDirect.js", false},
		{"unknown", "default", "es6", "redirectExport.js", false},
		{"all", "./other-module", "es6", "redirectExport.js", false},
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
		{"function", "foo()", "commonjs", "scopingtest.js", false},
		{"const", "bar", "commonjs", "scopingtest.js", false},
		{"var", "foobar", "commonjs", "scopingtest.js", false},
	}
	assert.ElementsMatch(t, exports, expectedExports, fmt.Sprint(exports))
}
