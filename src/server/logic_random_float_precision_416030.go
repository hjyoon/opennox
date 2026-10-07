package server

import "math"

// The CRT selects x87 precision 53 (004031F2), and the gameplay loop selects
// round toward zero (0043E2C1). Binary32 bounds, the 15-bit table, and the
// binary32 scale keep every finite result and residual within binary64's
// exponent range. Recover the exact rounding residual locally; never change
// the native thread's floating-point environment.
//
//go:noinline
func logicRandomFloatAddChop53_416030(a, b float64) float64 {
	sum := float64(a + b)
	bVirtual := float64(sum - a)
	remainder := float64(float64(a-float64(sum-bVirtual)) + float64(b-bVirtual))
	if sum > 0 && remainder < 0 || sum < 0 && remainder > 0 {
		return math.Nextafter(sum, 0)
	}
	return sum
}

//go:noinline
func logicRandomFloatMulChop53_416030(a, b float64) float64 {
	product := float64(a * b)
	// FMA recovers only the residual of this multiplication. The original
	// FSUB/FMUL/FMUL/FADD sequence is never contracted into a fused result.
	remainder := math.FMA(a, b, -product)
	if product > 0 && remainder < 0 || product < 0 && remainder > 0 {
		return math.Nextafter(product, 0)
	}
	return product
}
