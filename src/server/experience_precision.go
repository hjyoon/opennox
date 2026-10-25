package server

import "math"

// The original startup control word uses 53-bit precision with rounding toward
// zero. Reproduce its arithmetic boundaries locally without changing host FP
// state. Binary32 experience operands keep the intermediates in binary64 range.
//
//go:noinline
func experienceAdd53TowardZero(a, b float64) float64 {
	sum := a + b
	tail := sum - a
	residual := (a - (sum - tail)) + (b - tail)
	if (sum > 0 && residual < 0) || (sum < 0 && residual > 0) {
		return math.Nextafter(sum, 0)
	}
	return sum
}

//go:noinline
func experienceMul53TowardZero(a, b float64) float64 {
	product := a * b
	residual := math.FMA(a, b, -product)
	if (product > 0 && residual < 0) || (product < 0 && residual > 0) {
		return math.Nextafter(product, 0)
	}
	return product
}

//go:noinline
func experienceSpill32TowardZero(value float64) float32 {
	rounded := float32(value)
	if math.Abs(float64(rounded)) > math.Abs(value) {
		return math.Nextafter32(rounded, 0)
	}
	return rounded
}
