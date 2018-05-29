package evolution

import (
	"fmt"
	"github.com/blang/semver"
	"testing"
)

func TestSemverDiff(t *testing.T) {
	testCases := []struct {
		a, b, expected string
	}{
		{"0.0.1", "1.0.0", "major"},
		{"0.0.1", "0.1.0", "minor"},
		{"0.0.1", "0.0.2", "patch"},
		{"0.0.1-foo", "0.0.1-foo.bar", "prerelease"},
		{"0.10.0", "1.0.0-rc.1", "prerelease"},
		{"1.0.0-rc.3", "1.0.0", "major"},
		{"0.0.1", "0.0.1+foo.bar", "build"},
		{"0.0.1", "0.0.1", "equal"},
		{"0.0.2", "0.0.1", "downgrade"},
	}
	for _, test := range testCases {
		t.Run(fmt.Sprint(test), func(t *testing.T) {
			a := semver.MustParse(test.a)
			b := semver.MustParse(test.b)
			actual := SemverDiff(a, b)
			if actual != test.expected {
				t.Errorf("FAIL: Expected %v but got %v", test.expected, actual)
			}
		})
	}
}

func TestProcessVersions(t *testing.T) {
	testPackage := MustReadMetadataFromTestFile(lodashTestJsonPath, t)

	changes, err := ProcessVersions(testPackage, timeCutoff)
	if err != nil {
		t.Fatal(err)
	}
	t.Log(changes)
}
