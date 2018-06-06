package evolution

import (
	"testing"
)

func TestGetPackageMetadataForEachMonthExpress(t *testing.T) {
	testPackage := MustReadMetadataFromTestFile(expressTestJsonPath, t)

	timeMap := GetPackageMetadataForEachMonth(testPackage)

	t.Log(timeMap)
}

func TestGetPackageMetadataForEachMonthReact(t *testing.T) {
	testPackage := MustReadMetadataFromTestFile(reactTestJsonPath, t)

	timeMap := GetPackageMetadataForEachMonth(testPackage)

	t.Log(timeMap)
}
