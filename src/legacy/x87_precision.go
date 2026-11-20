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

func mulFloat64TowardZero(a, b float64) float64 {
	product := float64(a * b)
	residual := math.FMA(a, b, -product)
	if product > 0 && residual < 0 || product < 0 && residual > 0 {
		return math.Nextafter(product, 0)
	}
	return product
}

func storeFloat32TowardZero(value float64) float32 {
	narrow := float32(value)
	if math.Abs(float64(narrow)) > math.Abs(value) {
		narrow = math.Nextafter32(narrow, 0)
	}
	return narrow
}

// The PE x87 helper truncates to a signed qword and consumes its low dword.
// Invalid FISTP produces INT64_MIN, whose low dword is zero. Keep the upper
// bound exclusive: float64(MaxInt64) itself rounds up to 2^63.
func x87TruncQwordLow(value float64) int32 {
	if math.IsNaN(value) || math.IsInf(value, 0) || value < -0x1p63 || value >= 0x1p63 {
		return 0
	}
	return int32(uint32(int64(value)))
}
