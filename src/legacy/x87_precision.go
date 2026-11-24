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

// Invalid FISTP produces INT64_MIN. The upper bound is exclusive because
// float64(MaxInt64) rounds up to 2^63.
func x87TruncQword(value float64) int64 {
	if !(value >= -0x1p63 && value < 0x1p63) {
		return math.MinInt64
	}
	return int64(value)
}

func x87TruncQwordLow(value float64) int32 {
	return int32(x87TruncQword(value))
}
