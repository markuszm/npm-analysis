package evolution

import "testing"

func TestCalculatePopularity(t *testing.T) {
	downloadCounts := MustReadDownloadCountsFromTestFile(lodashDownloadCountsPath, t)

	popularity := CalculatePopularity(downloadCounts)
	t.Log(popularity)
}
