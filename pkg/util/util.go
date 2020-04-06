package util

// MinInt64 returns the min of two int64
func MinInt(a int, b int) int {
	if a <= b {
		return a
	} else {
		return b
	}
}
