package codeanalysis

import (
	"fmt"
	"github.com/markuszm/npm-analysis/resultprocessing"
	"github.com/stretchr/testify/assert"
	"go.uber.org/zap"
	"testing"
)

func TestExportCommonJS(t *testing.T) {
	exports := getExportsFromPackagePath("./testfiles/export/exportcjstest", t)

	expectedExports := []resultprocessing.Export{
		{"class", "Calculator", []string{}, "commonjs", "classExport", true, "Calculator"},
		{"function", "add", []string{"a", "b"}, "commonjs", "classExport", true, "Calculator.add"},
		{"function", "subtract", []string{"a", "b"}, "commonjs", "classExport", true, "Calculator.subtract"},
		{"class", "calc", []string{}, "commonjs", "classExport", false, "Calculator2"},
		{"function", "calc.add", []string{"a", "b"}, "commonjs", "classExport", false, "Calculator2.add"},
		{"function", "calc.subtract", []string{"a", "b"}, "commonjs", "classExport", false, "Calculator2.subtract"},
		{"class", "AdvancedCalculator", []string{}, "commonjs", "classInstanceExport", true, ""},
		{"function", "AdvancedCalculator.multiply", []string{"a", "b"}, "commonjs", "classInstanceExport", true, ""},
		{"function", "AdvancedCalculator.divide", []string{"a", "b"}, "commonjs", "classInstanceExport", true, ""},
		{"function", "Calculator.add", []string{"a", "b"}, "commonjs", "classInstanceExport", true, ""},
		{"function", "Calculator.subtract", []string{"a", "b"}, "commonjs", "classInstanceExport", true, ""},
		{"function", "Base.toString", []string{"str"}, "commonjs", "classInstanceExport", true, ""},
		{"const", "max", []string{}, "commonjs", "constantExport", false, "max"},
		{"member", "min", []string{}, "commonjs", "constantExport", false, "Number.MIN_VALUE"},
		{"function", "addThree", []string{"a", "b", "c"}, "commonjs", "directExport", true, "addThree"},
		{"member", "default", []string{}, "commonjs", "directExport", true, "foo.subThree"},
		{"function", "default", []string{"a"}, "commonjs", "directExport", true, ""},
		{"function", "add", []string{"a", "b"}, "commonjs", "indirectExport", false, "add"},
		{"function", "divide", []string{"a", "b"}, "commonjs", "indirectExport", false, ""},
		{"function", "multiply", []string{"a", "b"}, "commonjs", "indirectExport", false, ""},
		{"object", "foo", []string{}, "commonjs", "indirectExport", false, ""},
		{"function", "foo.add2", []string{"a"}, "commonjs", "indirectExport", false, ""},
		{"function", "foo.add4", []string{"a"}, "commonjs", "indirectExport", false, ""},
		{"function", "foo.add8", []string{"a"}, "commonjs", "indirectExport", false, ""},
		{"function", "foo.add16", []string{"a"}, "commonjs", "indirectExport", false, ""},
		{"function", "foo.add32", []string{"a"}, "commonjs", "indirectExport", false, ""},
		{"member", "e", []string{}, "commonjs", "indirectExport", false, "Math.E"},
		{"class", "calculator", []string{}, "commonjs", "indirectExport", false, "Calculator"},
		{"function", "calculator.add", []string{"a", "b"}, "commonjs", "indirectExport", false, "Calculator.add"},
		{"function", "calculator.subtract", []string{"a", "b"}, "commonjs", "indirectExport", false, "Calculator.subtract"},
		{"function", "abs", []string{"x"}, "commonjs", "methodExport", false, "abs"},
		{"function", "sqrt", []string{"x"}, "commonjs", "methodExport", false, "sqrtDefault"},
		{"function", "pow", []string{"x", "exp"}, "commonjs", "methodExport", false, ""},
		{"function", "floor", []string{"x"}, "commonjs", "methodExport", false, ""},
		{"function", "parseInt", []string{"x", "r"}, "commonjs", "methodExport", false, ""},
		{"function", "add2", []string{"a"}, "commonjs", "objectExport", true, ""},
		{"function", "add4", []string{"a"}, "commonjs", "objectExport", true, ""},
		{"function", "add8", []string{"a"}, "commonjs", "objectExport", true, ""},
		{"function", "add16", []string{"a"}, "commonjs", "objectExport", true, ""},
		{"function", "add32", []string{"a"}, "commonjs", "objectExport", true, ""},
		{"function", "addAll", []string{"a", "...b"}, "commonjs", "objectExport", true, ""},
		{"member", "theSolution", []string{}, "commonjs", "objectExport", true, ""},
		{"var", "foo", []string{}, "commonjs", "objectExport", false, "bar"},
		{"function", "foo.sub2", []string{"a"}, "commonjs", "objectExport", false, "bar.sub2"},
		{"function", "foo.sub4", []string{"a"}, "commonjs", "objectExport", false, "bar.sub4"},
		{"function", "foo.sub8", []string{"a"}, "commonjs", "objectExport", false, "bar.sub8"},
		{"function", "foo.sub16", []string{"a"}, "commonjs", "objectExport", false, "bar.sub16"},
		{"var", "foo.theSolution", []string{}, "commonjs", "objectExport", false, "bar.theSolution"},
	}
	assert.ElementsMatch(t, exports, expectedExports, fmt.Sprint(exports))
}

func TestExportES6(t *testing.T) {
	exports := getExportsFromPackagePath("./testfiles/export/exportes6test", t)

	expectedExports := []resultprocessing.Export{
		{"function", "abs", []string{}, "es6", "defaultExport", true, ""},
		{"class", "Calculator", []string{}, "es6", "exportNamedClass", false, ""},
		{"function", "Calculator.add", []string{"a", "b"}, "es6", "exportNamedClass", false, ""},
		{"function", "Calculator.subtract", []string{"a", "b"}, "es6", "exportNamedClass", false, ""},
		{"var", "foo", []string{}, "es6", "exportOtherName", false, "f"},
		{"function", "bar", []string{}, "es6", "exportOtherName", false, "b"},
		{"function", "default function", []string{"obj"}, "es6", "mixExport", true, ""},
		{"function", "each", []string{"obj", "iterator", "context"}, "es6", "mixExport", false, ""},
		{"function", "forEach", []string{"obj", "iterator", "context"}, "es6", "mixExport", false, "each"},
		{"function", "cube", []string{"x"}, "es6", "namedExport", false, "cube"},
		{"const", "foo", []string{}, "es6", "namedExport", false, "foo"},
		{"var", "graph.options", []string{}, "es6", "namedExport", false, "graph.options"},
		{"function", "graph.draw", []string{}, "es6", "namedExport", false, "graph.draw"},
		{"let", "graph", []string{}, "es6", "namedExport", false, "graph"},
		{"const", "sqrt", []string{}, "es6", "namedExportDirect", false, ""},
		{"function", "square", []string{"x"}, "es6", "namedExportDirect", false, ""},
		{"function", "diag", []string{"x", "y"}, "es6", "namedExportDirect", false, ""},
		{"unknown", "default", []string{}, "es6", "redirectExport", false, "default"},
		{"all", "./other-module", []string{}, "es6", "redirectExport", false, ""},
	}
	assert.ElementsMatch(t, exports, expectedExports, fmt.Sprint(exports))
}

func TestExportScoping(t *testing.T) {
	exports := getExportsFromPackagePath("./testfiles/export/scoping", t)

	expectedExports := []resultprocessing.Export{
		{"function", "foo", []string{}, "commonjs", "scopingtest", false, "foo"},
		{"const", "bar", []string{}, "commonjs", "scopingtest", false, "foo"},
		{"var", "foobar", []string{}, "commonjs", "scopingtest", false, "bar"},
	}
	assert.ElementsMatch(t, exports, expectedExports, fmt.Sprint(exports))
}

func TestExportEdgecases(t *testing.T) {
	exports := getExportsFromPackagePath("./testfiles/export/edgecases", t)

	expectedExports := []resultprocessing.Export{
		{
			ExportType: "const",
			Identifier: "c",
			Arguments:  []string{},
			BundleType: "commonjs",
			File:       "assignmentChain",
			IsDefault:  true,
			Local:      "c",
		},
		{
			ExportType: "const",
			Identifier: "a",
			Arguments:  []string{},
			BundleType: "commonjs",
			File:       "assignmentChain",
			Local:      "c",
		},
		{
			ExportType: "const",
			Identifier: "b",
			Arguments:  []string{},
			BundleType: "commonjs",
			File:       "assignmentChain",
			Local:      "c",
		},
		{
			ExportType: "const",
			Identifier: "c",
			Arguments:  []string{},
			BundleType: "commonjs",
			File:       "assignmentChain",
			Local:      "c",
		},
		{
			ExportType: "unknown",
			Identifier: "foo",
			Arguments:  []string{},
			BundleType: "commonjs",
			File:       "assignmentChain",
			Local:      "module.exports.bar = null",
		},
		{
			ExportType: "unknown",
			Identifier: "bar",
			Arguments:  []string{},
			BundleType: "commonjs",
			File:       "assignmentChain",
			Local:      "null",
		},
		{
			ExportType: "function",
			Identifier: "abs",
			Arguments:  []string{"x"},
			BundleType: "commonjs",
			File:       "methodDefinedLater",
			Local:      "abs",
		},
		{
			ExportType: "function",
			Identifier: "sqrt",
			Arguments:  []string{"x"},
			BundleType: "commonjs",
			File:       "methodDefinedLater",
			Local:      "sqrtDefault",
		},
	}
	assert.ElementsMatch(t, exports, expectedExports, fmt.Sprint(exports))
}

func getExportsFromPackagePath(packagePath string, t *testing.T) []resultprocessing.Export {
	const analysisPath = "./exports-analysis/analysis"
	logger := zap.NewNop().Sugar()
	analysis := NewASTAnalysis(logger, analysisPath)
	result, err := analysis.AnalyzePackage(packagePath)
	if err != nil {
		t.Fatal(err)
	}
	exports, err := resultprocessing.TransformToExports(result)
	if err != nil {
		t.Fatal(err)
	}

	return exports
}
