package evolution

import "testing"

func TestCalculatePopularity(t *testing.T) {
	downloadCounts := MustReadDownloadCountsFromTestFile(lodashDownloadCountsPath, t)

	popularity := CalculatePopularity("lodash", downloadCounts)
	t.Log(popularity)
}
