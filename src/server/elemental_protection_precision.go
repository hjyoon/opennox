package server

import "math"

// Protection modifiers and balance values start as binary32. The original
// gameplay environment adds them with 53-bit precision toward zero. Recover
// the exact residual locally, without changing the host rounding mode.
//
//go:noinline
func elementalProtectionAddChop53(a, b float64) float64 {
	sum := float64(a + b)
	bVirtual := float64(sum - a)
	residual := float64(float64(a-float64(sum-bVirtual)) + float64(b-bVirtual))
	if sum > 0 && residual < 0 || sum < 0 && residual > 0 {
		return math.Nextafter(sum, 0)
	}
	return sum
}

// The original FST keeps the unspilled accumulator alive. Finite overflow
// stores the largest finite binary32 value under rounding toward zero.
func elementalProtectionSpill32(value float64) float32 {
	narrow := float32(value)
	if math.Abs(float64(narrow)) > math.Abs(value) {
		return math.Nextafter32(narrow, 0)
	}
	return narrow
}
