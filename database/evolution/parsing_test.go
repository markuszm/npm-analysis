package evolution

import (
	"encoding/json"
	"github.com/blang/semver"
	"github.com/markuszm/npm-analysis/database/model"
	"io/ioutil"
	"testing"
)

const lodashTestJsonPath = "./testfiles/lodash-all.json"

func TestLicenseParsing(t *testing.T) {
	bytes, err := ioutil.ReadFile(lodashTestJsonPath)
	if err != nil {
		t.Fatal(err)
	}
	var testPackage model.Metadata
	err = json.Unmarshal(bytes, &testPackage)
	if err != nil {
		t.Fatal(err)
	}

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

func TestVersionParsing(t *testing.T) {
	testVersion := "4.0.0-feature-remove-unsave-lifecycles-5d227cb7"
	v := semver.MustParse(testVersion)
	if v.String() != testVersion {
		t.Errorf("Parsed version to string not equal")
	}
}

func TestMaintainerRegex(t *testing.T) {
	sampleMaintainer := "Aadit M Shah (https://aaditmshah.github.io/) <aaditmshah@fastmail.fm>"
	person := parseSingleMaintainerStr(sampleMaintainer)
	if person.Name != "Aadit M Shah" {
		t.Errorf("expected Aadit M Shah but got %v", person.Name)
	}
}

func TestMaintainerChangeList(t *testing.T) {
	bytes, err := ioutil.ReadFile(lodashTestJsonPath)
	if err != nil {
		t.Fatal(err)
	}
	var testPackage model.Metadata
	err = json.Unmarshal(bytes, &testPackage)
	if err != nil {
		t.Fatal(err)
	}

	changes, err := ProcessMaintainers(testPackage)
	if err != nil {
		t.Fatal(err)
	}
	t.Log(changes)
}
