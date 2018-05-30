package evolution

import (
	"encoding/json"
	"testing"
)

// TODO: maybe for learning replace real API with http mock
// INTEGRATION TEST WITH REAL API
func TestDownloadCount(t *testing.T) {
	downloadCounts, err := GetDownloadCountsForPackage("lodash")
	if err != nil {
		t.Errorf("Failed downloading packages with error %v", err)
	}

	actual := DownloadCountResponse{}
	err = json.Unmarshal([]byte(downloadCounts), &actual)
	if err != nil {
		t.Errorf("Error unmarshalling: %v", err)
	}

	expectedStart := "2015-01-10"
	actualStart := actual.Start
	if actualStart != expectedStart {
		t.Errorf("Expected start to be %v but got %v", expectedStart, actualStart)
	}

	expectedEnd := "2018-04-13"
	actualEnd := actual.End
	if actualEnd != expectedEnd {
		t.Errorf("Expected end to be %v but got %v", expectedEnd, actualEnd)
	}

	expectedDownloadCountDays := 1190
	actualDownloadCountDays := len(actual.DownloadCounts)
	if actualDownloadCountDays != expectedDownloadCountDays {
		t.Errorf("Expected download count list to have %v days but only got %v days", expectedDownloadCountDays, actualDownloadCountDays)
	}

	t.Log(actual)
}

// INTEGRATION TEST WITH REAL API
func TestDownloadCountBulk(t *testing.T) {
	packages := []string{"a", "a-app", "a-art-dialog", "a-autocomplete", "a-average", "a-b-c-z", "a-baas", "a-baas-util", "a-bear", "a-better-console", "a-big-triangle", "a-bit-of", "a-black-hole", "a-brief-history", "a-browser-info", "a-builder", "a-buildev", "a-calendar-plugin", "a-carousel", "a-casa-tutti-bene-streaming-ita", "a-cm-2.5", "a-color-picker", "a-configuration", "a-construct-fn", "a-contains", "a-css-loader", "a-css-loader_postcss-modules-values", "a-csv", "a-cube", "a-d-d", "a-d-s-r", "a-day", "a-demo-app", "a-dep-b", "a-deprecated-modal", "a-di", "a-difference", "a-dispatcher", "a-draftjs-to-html", "a-dropzone", "a-ejs", "a-events", "a-extractor", "a-fill", "a-find", "a-first", "a-for-apple", "a-form", "a-frame", "a-framedc", "a-french-javascript-developer", "a-global", "a-gulp-license", "a-html-to-draftjs", "a-i", "a-i18n", "a-input-test", "a-javascript-and-typescript-documentation-generator-based-on-typescript-compiler", "a-json-validator", "a-jwk-generator", "a-kind-of-magic", "a-kit", "a-last", "a-lerna-test", "a-lerna-test-button", "a-lerna-test-nav", "a-letter-for-you", "a-library", "a-line", "a-loader", "a-locale", "a-localization", "a-logger", "a-max", "a-median", "a-min", "a-mmd", "a-mode-mt-cache", "a-module-doing-pretty-much-nothing", "a-module-using-loose-envify", "a-module-with-babelrc", "a-napi-example", "a-native-example", "a-native-module", "a-native-module-without-prebuild", "a-nice-time", "a-node-module", "a-nonce-generator", "a-normal-testing-repo", "a-npm-module", "a-npm-package", "a-npm-publishing-sample", "a-object", "a-p-i", "a-painter-loader-component", "a-partition", "a-passport-ldap", "a-plugin", "a-plus-forms", "a-plus-forms-bootstrap", "a-plus-forms-json-validator", "a-pollo", "a-popupjs", "a-promise", "a-promise-queue", "a-pure-typescript-test", "a-random", "a-range", "a-rangy", "a-ray", "a-react-datepicker", "a-react-simple-tab", "a-react-template", "a-react-timepicker", "a-record", "a-recorder", "a-rel", "a-replace-webpack-plugin", "a-roller", "a-rule", "a-scrollbar-fill", "a-seal", "a-server", "a-sfake-style-loader", "a-shuffle", "a-simple-carousel", "a-simple-connect-webserver", "a-simple-package"}

	downloadCounts, err := GetDownloadCountsBulk(packages)
	if err != nil {
		t.Errorf("Failed downloading packages with error %v", err)
	}

	for _, d := range downloadCounts {
		actual := DownloadCountResponse{}
		err = json.Unmarshal([]byte(d), &actual)
		if err != nil {
			t.Errorf("Error unmarshalling: %v", err)
		}

		expectedStart := "2015-01-10"
		actualStart := actual.Start
		if actualStart != expectedStart {
			t.Errorf("Expected start to be %v but got %v", expectedStart, actualStart)
		}

		expectedEnd := "2018-04-13"
		actualEnd := actual.End
		if actualEnd != expectedEnd {
			t.Errorf("Expected end to be %v but got %v", expectedEnd, actualEnd)
		}

		expectedDownloadCountDays := 1190
		actualDownloadCountDays := len(actual.DownloadCounts)
		if actualDownloadCountDays != expectedDownloadCountDays {
			t.Errorf("Expected download count list to have %v days but only got %v days", expectedDownloadCountDays, actualDownloadCountDays)
		}

	}

}

func TestDownloadCountScopedName(t *testing.T) {
	downloadCounts, err := GetDownloadCountsForPackage("@angular/core")
	if err != nil {
		t.Errorf("Failed downloading packages with error %v", err)
	}

	actual := DownloadCountResponse{}
	err = json.Unmarshal([]byte(downloadCounts), &actual)
	if err != nil {
		t.Errorf("Error unmarshalling: %v", err)
	}

	expectedStart := "2015-01-10"
	actualStart := actual.Start
	if actualStart != expectedStart {
		t.Errorf("Expected start to be %v but got %v", expectedStart, actualStart)
	}

	expectedEnd := "2018-04-13"
	actualEnd := actual.End
	if actualEnd != expectedEnd {
		t.Errorf("Expected end to be %v but got %v", expectedEnd, actualEnd)
	}

	expectedDownloadCountDays := 1190
	actualDownloadCountDays := len(actual.DownloadCounts)
	if actualDownloadCountDays != expectedDownloadCountDays {
		t.Errorf("Expected download count list to have %v days but only got %v days", expectedDownloadCountDays, actualDownloadCountDays)
	}

	t.Log(actual)
}

func TestParseDate(t *testing.T) {
	exampleDate := "2018-04-13"

	date := MustParseDate(exampleDate)

	expectedYear := 2018
	expectedMonth := "April"
	expectedDay := 13

	actualYear := date.Year()
	actualMonth := date.Month().String()
	actualDay := date.Day()

	if actualYear != expectedYear {
		t.Errorf("Expected year %v but got %v", expectedYear, actualYear)
	}

	if actualMonth != expectedMonth {
		t.Errorf("Expected month %v but got %v", expectedMonth, actualMonth)
	}

	if actualDay != expectedDay {
		t.Errorf("Expected day %v but got %v", expectedDay, actualDay)
	}

	t.Log(date)
}
