package evolution

import (
	"fmt"
	"github.com/blang/semver"
	"testing"
)

// TODO: move tests to respective testfile

func TestLicenseParsing(t *testing.T) {
	testPackage := MustReadMetadataFromTestFile(lodashTestJsonPath, t)

	for _, v := range testPackage.Versions {
		license := ProcessLicense(v)
		if license == "" {
			license = ProcessLicenses(v)
		}
		if license != "MIT" {
			t.Errorf("Wanted MIT but got %v", license)
		}
	}
}

func TestLicenseChanges(t *testing.T) {
	testPackage := MustReadMetadataFromTestFile(lodashTestJsonPath, t)

	changes, err := ProcessLicenseChanges(testPackage, timeCutoff)
	if err != nil {
		t.Fatal(err)
	}
	if len(changes) != 0 {
		t.Log(changes)
		t.Errorf("Expected zero license changes for lodash")
	}
}

func TestVersionParsing(t *testing.T) {
	testVersion := "4.0.0-feature-remove-unsave-lifecycles-5d227cb7"
	v := semver.MustParse(testVersion)
	if v.String() != testVersion {
		t.Errorf("Parsed version to string not equal")
	}
}

func TestMaintainerRegexFull(t *testing.T) {
	sampleMaintainer := "rreverser"
	person := parseSingleMaintainerStr(sampleMaintainer)
	if person.Name != "rreverser" {
		t.Errorf("expected rreverser but got %v", person.Name)
	}
}

func TestMaintainerRegexOnlyName(t *testing.T) {
	sampleMaintainer := "Aadit M Shah (https://aaditmshah.github.io/) <aaditmshah@fastmail.fm>"
	person := parseSingleMaintainerStr(sampleMaintainer)
	if person.Name != "Aadit M Shah" {
		t.Errorf("expected Aadit M Shah but got %v", person.Name)
	}
}

func TestRegressionMaintainerChange(t *testing.T) {
	testPackage := MustReadMetadataFromTestFile(lodashTestJsonPath, t)

	regressionString := `[{lodash jdalton 2012-04-23 16:37:12.603 +0000 UTC INITIAL 0.1.0} {lodash mathias 2013-09-23 05:57:42.595 +0000 UTC ADDED 2.1.0} {lodash phated 2013-09-23 05:57:42.595 +0000 UTC ADDED 2.1.0} {lodash kitcambridge 2013-09-23 05:57:42.595 +0000 UTC ADDED 2.1.0} {lodash d10 2015-01-30 09:33:51.621 +0000 UTC ADDED 3.0.1} {lodash kitcambridge 2016-01-12 23:13:20.539 +0000 UTC REMOVED 4.0.0} {lodash d10 2016-01-12 23:13:20.539 +0000 UTC REMOVED 4.0.0} {lodash jridgewell 2016-02-16 07:10:16.856 +0000 UTC ADDED 4.4.0} {lodash jridgewell 2016-05-02 15:01:02.189 +0000 UTC REMOVED 4.11.2} {lodash phated 2016-10-31 06:49:14.797 +0000 UTC REMOVED 4.16.5}]`

	changes, err := ProcessMaintainersTimeSorted(testPackage, timeCutoff)
	if err != nil {
		t.Fatal(err)
	}

	t.Log()

	actual := fmt.Sprint(changes)
	if regressionString != actual {
		t.Errorf("REGRESSION in maintainer change list - manual check necessary")
	}
}

func TestDependencyChangeList(t *testing.T) {
	testPackage := MustReadMetadataFromTestFile(expressTestJsonPath, t)

	changes, err := ProcessDependencies(testPackage, timeCutoff)
	if err != nil {
		t.Fatal(err)
	}
	t.Log(changes)
}

func TestDependencyChangeListInitial(t *testing.T) {
	testPackage := MustReadMetadataFromTestFile(devCliTestJsonPath, t)

	changes, err := ProcessDependencies(testPackage, timeCutoff)
	if err != nil {
		t.Fatal(err)
	}
	t.Log(changes)
}
