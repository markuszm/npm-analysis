package evolution

import (
	"encoding/json"
	"testing"
)

func TestCalculateAveragePopularityByYear(t *testing.T) {
	downloadCounts := MustReadDownloadCountsFromTestFile(lodashDownloadCountsPath, t)

	popularity := CalculateAveragePopularityByYear("lodash", downloadCounts)

	bytes, err := json.Marshal(popularity)
	if err != nil {
		t.Fatal(err)
	}
	t.Log(string(bytes))
}

func TestCalculateAveragePopularityByMonth(t *testing.T) {
	downloadCounts := MustReadDownloadCountsFromTestFile(lodashDownloadCountsPath, t)

	popularity := CalculateAveragePopularityByMonth("lodash", downloadCounts)

	bytes, err := json.Marshal(popularity)
	if err != nil {
		t.Fatal(err)
	}
	t.Log(string(bytes))
}

func TestCalculatePopularityByMonth(t *testing.T) {
	downloadCounts := MustReadDownloadCountsFromTestFile(lodashDownloadCountsPath, t)

	popularity := CalculatePopularityByMonth("lodash", downloadCounts)

	bytes, err := json.Marshal(popularity)
	if err != nil {
		t.Fatal(err)
	}
	t.Log(string(bytes))
}
