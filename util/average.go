package util

func AvgInts(a int, b int) float64 {
	if b == 0 {
		return 0.0
	}
	return float64(a) / float64(b)
}
