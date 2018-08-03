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
		{"class", "Calculator", "commonjs", "classExport.js", true, "Calculator"},
		{"function", "add(a,b)", "commonjs", "classExport.js", true, "Calculator.add(a,b)"},
		{"function", "subtract(a,b)", "commonjs", "classExport.js", true, "Calculator.subtract(a,b)"},
		{"class", "calc", "commonjs", "classExport.js", false, "Calculator2"},
		{"function", "calc.add(a,b)", "commonjs", "classExport.js", false, "Calculator2.add(a,b)"},
		{"function", "calc.subtract(a,b)", "commonjs", "classExport.js", false, "Calculator2.subtract(a,b)"},
		{"class", "AdvancedCalculator", "commonjs", "classInstanceExport.js", true, ""},
		{"function", "AdvancedCalculator.multiply(a,b)", "commonjs", "classInstanceExport.js", true, ""},
		{"function", "AdvancedCalculator.divide(a,b)", "commonjs", "classInstanceExport.js", true, ""},
		{"function", "Calculator.add(a,b)", "commonjs", "classInstanceExport.js", true, ""},
		{"function", "Calculator.subtract(a,b)", "commonjs", "classInstanceExport.js", true, ""},
		{"function", "Base.toString(str)", "commonjs", "classInstanceExport.js", true, ""},
		{"const", "max", "commonjs", "constantExport.js", false, "max"},
		{"member", "min", "commonjs", "constantExport.js", false, "Number.MIN_VALUE"},
		{"function", "addThree(a,b,c)", "commonjs", "directExport.js", true, "addThree(a,b,c)"},
		{"unknown", "foo.subThree()", "commonjs", "directExport.js", true, ""},
		{"function", "add(a,b)", "commonjs", "indirectExport.js", false, "add(a,b)"},
		{"function", "divide(a,b)", "commonjs", "indirectExport.js", false, ""},
		{"function", "multiply(a,b)", "commonjs", "indirectExport.js", false, ""},
		{"object", "foo", "commonjs", "indirectExport.js", false, ""},
		{"function", "foo.add2(a)", "commonjs", "indirectExport.js", false, ""},
		{"function", "foo.add4(a)", "commonjs", "indirectExport.js", false, ""},
		{"function", "foo.add8(a)", "commonjs", "indirectExport.js", false, ""},
		{"function", "foo.add16(a)", "commonjs", "indirectExport.js", false, ""},
		{"function", "foo.add32(a)", "commonjs", "indirectExport.js", false, ""},
		{"member", "e", "commonjs", "indirectExport.js", false, "Math.E"},
		{"class", "calculator", "commonjs", "indirectExport.js", false, "Calculator"},
		{"function", "calculator.add(a,b)", "commonjs", "indirectExport.js", false, "Calculator.add(a,b)"},
		{"function", "calculator.subtract(a,b)", "commonjs", "indirectExport.js", false, "Calculator.subtract(a,b)"},
		{"function", "abs(x)", "commonjs", "methodExport.js", false, "abs(x)"},
		{"function", "sqrt(x)", "commonjs", "methodExport.js", false, "sqrtDefault(x)"},
		{"function", "pow(x,exp)", "commonjs", "methodExport.js", false, ""},
		{"function", "floor(x)", "commonjs", "methodExport.js", false, ""},
		{"function", "parseInt(x,r)", "commonjs", "methodExport.js", false, ""},
		{"function", "add2(a)", "commonjs", "objectExport.js", true, ""},
		{"function", "add4(a)", "commonjs", "objectExport.js", true, ""},
		{"function", "add8(a)", "commonjs", "objectExport.js", true, ""},
		{"function", "add16(a)", "commonjs", "objectExport.js", true, ""},
		{"function", "add32(a)", "commonjs", "objectExport.js", true, ""},
		{"function", "addAll(a,...b)", "commonjs", "objectExport.js", true, ""},
		{"member", "theSolution", "commonjs", "objectExport.js", true, ""},
		{"var", "foo", "commonjs", "objectExport.js", false, "bar"},
		{"function", "foo.sub2(a)", "commonjs", "objectExport.js", false, "bar.sub2(a)"},
		{"function", "foo.sub4(a)", "commonjs", "objectExport.js", false, "bar.sub4(a)"},
		{"function", "foo.sub8(a)", "commonjs", "objectExport.js", false, "bar.sub8(a)"},
		{"function", "foo.sub16(a)", "commonjs", "objectExport.js", false, "bar.sub16(a)"},
		{"var", "foo.theSolution", "commonjs", "objectExport.js", false, "bar.theSolution"},
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
		{"function", "abs()", "es6", "defaultExport.js", true, ""},
		{"class", "Calculator", "es6", "exportNamedClass.js", false, ""},
		{"function", "Calculator.add(a,b)", "es6", "exportNamedClass.js", false, ""},
		{"function", "Calculator.subtract(a,b)", "es6", "exportNamedClass.js", false, ""},
		{"var", "foo", "es6", "exportOtherName.js", false, "f"},
		{"function", "bar()", "es6", "exportOtherName.js", false, "b"},
		{"function", "default function(obj)", "es6", "mixExport.js", true, ""},
		{"function", "each(obj,iterator,context)", "es6", "mixExport.js", false, ""},
		{"function", "forEach(obj,iterator,context)", "es6", "mixExport.js", false, "each"},
		{"function", "cube(x)", "es6", "namedExport.js", false, "cube"},
		{"const", "foo", "es6", "namedExport.js", false, "foo"},
		{"var", "graph.options", "es6", "namedExport.js", false, "graph.options"},
		{"function", "graph.draw()", "es6", "namedExport.js", false, "graph.draw()"},
		{"let", "graph", "es6", "namedExport.js", false, "graph"},
		{"const", "sqrt", "es6", "namedExportDirect.js", false, ""},
		{"function", "square(x)", "es6", "namedExportDirect.js", false, ""},
		{"function", "diag(x,y)", "es6", "namedExportDirect.js", false, ""},
		{"unknown", "default", "es6", "redirectExport.js", false, "default"},
		{"all", "./other-module", "es6", "redirectExport.js", false, ""},
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
		{"function", "foo()", "commonjs", "scopingtest.js", false, "foo()"},
		{"const", "bar", "commonjs", "scopingtest.js", false, "foo"},
		{"var", "foobar", "commonjs", "scopingtest.js", false, "bar"},
	}
	assert.ElementsMatch(t, exports, expectedExports, fmt.Sprint(exports))
}
