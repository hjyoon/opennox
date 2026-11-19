package legacy

import "math"

// Explicit conversions separate the original x87 operations and inhibit
// contraction. Match gameplay precision53/chop without changing thread state.
//
//go:noinline
func addFloat64TowardZero(a, b float64) float64 {
	sum := float64(a + b)
	vb := float64(sum - a)
	err := float64(float64(a-float64(sum-vb)) + float64(b-vb))
	if sum > 0 && err < 0 || sum < 0 && err > 0 {
		return math.Nextafter(sum, 0)
	}
	return sum
}

func storeFloat32TowardZero(value float64) float32 {
	narrow := float32(value)
	if math.Abs(float64(narrow)) > math.Abs(value) {
		narrow = math.Nextafter32(narrow, 0)
	}
	return narrow
}
