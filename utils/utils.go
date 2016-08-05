package utils

import "math"

func IsEmptyStr(s string) bool {
	return len(s) == 0
}

func TruncateFloat(some float64, precise int) float64 {
	p := float64(math.Pow(10,  float64(precise)))
	return float64(int(some * p)) / p
}