package evolution

import (
	"fmt"
	semver2 "github.com/Masterminds/semver"
	"github.com/blang/semver"
	"testing"
	"time"
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
		{"1.8.6-beta", "1.8.6", "patch"},
		{"1.8.0-beta", "1.8.0", "minor"},
		{"1.1.2-beta.0", "1.1.2", "patch"},
		{"0.0.1", "0.0.1+foo.bar", "build"},
		{"0.0.1+foo.bar", "0.0.1", "build"},
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

func TestSemverValid(t *testing.T) {
	rangeVer := "^1.0.0"
	semverRange, _ := semver2.NewConstraint(rangeVer)

	ver := "1.2.1"
	s := semver2.MustParse(ver)
	t.Log(semverRange.Check(s))
}

func TestProcessVersions(t *testing.T) {
	testPackage := MustReadMetadataFromTestFile(lodashTestJsonPath, t)

	changes, err := ProcessVersions(testPackage, timeCutoff)
	if err != nil {
		t.Fatal(err)
	}
	t.Log(changes)
}

func TestFindLatestVersionExpress(t *testing.T) {
	testPackage := MustReadMetadataFromTestFile(expressTestJsonPath, t)
	specificDate := time.Date(2017, 5, 1, 0, 0, 0, 0, time.UTC)

	latestVer := FindLatestVersion(testPackage, specificDate)

	expectedVer := "4.15.2"
	if latestVer != expectedVer {
		t.Errorf("Expected %v but got %v", expectedVer, latestVer)
	}

	t.Log(latestVer)
}

func TestFindLatestVersionReact(t *testing.T) {
	testPackage := MustReadMetadataFromTestFile(reactTestJsonPath, t)
	specificDate := time.Date(2017, 5, 1, 0, 0, 0, 0, time.UTC)

	latestVer := FindLatestVersion(testPackage, specificDate)

	expectedVer := "15.5.4"
	if latestVer != expectedVer {
		t.Errorf("Expected %v but got %v", expectedVer, latestVer)
	}

	t.Log(latestVer)
}
