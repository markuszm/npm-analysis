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
