package codeanalysis

import (
	"fmt"
	"testing"
)

func TestGetPackageFilePath(t *testing.T) {
	testCases := []struct {
		packageName, version, expected string
	}{
		{"lodash", "1.0.0", "l/o/lodash-1.0.0.tgz"},
		{"@angular/core", "2.1.5", "@/a/@angular_core-2.1.5.tgz"},
	}
	for _, test := range testCases {
		t.Run(fmt.Sprint(test), func(t *testing.T) {
			actual := GetPackageFilePath(test.packageName, test.version)

			if actual != test.expected {
				t.Errorf("expected %v but got %v", test.expected, actual)
			}
		})
	}
}
