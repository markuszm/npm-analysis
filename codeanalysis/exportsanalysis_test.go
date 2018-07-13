package codeanalysis

import (
	"fmt"
	"github.com/stretchr/testify/assert"
	"go.uber.org/zap"
	"testing"
)

func TestExportCommonJS(t *testing.T) {
	logger := zap.NewNop().Sugar()
	analysis := NewExportsAnalysis(logger)
	result, err := analysis.AnalyzePackage("./testfiles/export/exportcjstest")

	if err != nil {
		t.Fatal(err)
	}

	exports := result.([]Export)

	expectedExports := []Export{
		{"class", "default", "commonjs"},
		{"function", "add(a,b)", "commonjs"},
		{"function", "subtract(a,b)", "commonjs"},
		{"class", "calc", "commonjs"},
		{"function", "calc.add(a,b)", "commonjs"},
		{"function", "calc.subtract(a,b)", "commonjs"},
		{"class", "AdvancedCalculator", "commonjs"},
		{"function", "AdvancedCalculator.multiply(a,b)", "commonjs"},
		{"function", "AdvancedCalculator.divide(a,b)", "commonjs"},
		{"function", "Calculator.add(a,b)", "commonjs"},
		{"function", "Calculator.substract(a,b)", "commonjs"},
		{"function", "Base.toString(str)", "commonjs"},
		{"const", "max", "commonjs"},
		{"member", "min", "commonjs"},
		{"unknown", "default", "commonjs"},
		{"function", "add(a,b)", "commonjs"},
		{"function", "divide(a,b)", "commonjs"},
		{"function", "multiply(a,b)", "commonjs"},
		{"object", "foo", "commonjs"},
		{"function", "foo.add2(a)", "commonjs"},
		{"function", "foo.add4(a)", "commonjs"},
		{"function", "foo.add8(a)", "commonjs"},
		{"function", "foo.add16(a)", "commonjs"},
		{"function", "foo.add32(a)", "commonjs"},
		{"member", "e", "commonjs"},
		{"class", "calculator", "commonjs"},
		{"function", "calculator.add(a,b)", "commonjs"},
		{"function", "calculator.substract(a,b)", "commonjs"},
		{"function", "abs(x)", "commonjs"},
		{"function", "sqrt(x)", "commonjs"},
		{"function", "pow(x,exp)", "commonjs"},
		{"function", "floor(x)", "commonjs"},
		{"function", "parseInt(x,r)", "commonjs"},
		{"function", "add2(a)", "commonjs"},
		{"function", "add4(a)", "commonjs"},
		{"function", "add8(a)", "commonjs"},
		{"function", "add16(a)", "commonjs"},
		{"function", "add32(a)", "commonjs"},
		{"function", "addAll(a,RestElement)", "commonjs"},
		{"member", "theSolution", "commonjs"},
		{"object", "foo", "commonjs"},
		{"function", "foo.sub2(a)", "commonjs"},
		{"function", "foo.sub4(a)", "commonjs"},
		{"function", "foo.sub8(a)", "commonjs"},
		{"function", "foo.sub16(a)", "commonjs"},
		{"member", "foo.theSolution", "commonjs"},
	}
	assert.ElementsMatch(t, exports, expectedExports, fmt.Sprint(exports))
}

func TestExportES6(t *testing.T) {
	logger := zap.NewNop().Sugar()
	analysis := NewExportsAnalysis(logger)
	result, err := analysis.AnalyzePackage("./testfiles/export/exportes6test")

	if err != nil {
		t.Fatal(err)
	}

	exports := result.([]Export)

	expectedExports := []Export{
		{"function", "default.abs()", "es6"},
		{"class", "Calculator", "es6"},
		{"function", "Calculator.add(a,b)", "es6"},
		{"function", "Calculator.substract(a,b)", "es6"},
		{"function", "default.default function(obj)", "es6"},
		{"function", "each(obj,iterator,context)", "es6"},
		{"unknown", "forEach", "es6"},
		{"function", "cube(x)", "es6"},
		{"const", "foo", "es6"},
		{"var", "graph.options", "es6"},
		{"function", "graph.draw", "es6"},
		{"let", "graph", "es6"},
		{"const", "sqrt", "es6"},
		{"function", "square(x)", "es6"},
		{"function", "diag(x,y)", "es6"},
		{"unknown", "default", "es6"},
		{"all", "./other-module", "es6"},
	}
	assert.ElementsMatch(t, exports, expectedExports, fmt.Sprint(exports))
}
