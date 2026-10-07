package server

import "math"

// GAME.EXE's CRT 004031F2 selects 53-bit x87 precision and the
// gameplay loop 0043E2C1 selects round-toward-zero. Binary32 input
// coordinates bound the intermediates inside binary64's exponent range.
// Correct only rounding, without changing the calling thread's FPU.
//
//go:noinline
func pointOnLineAddChop53_57C8A0(a, b float64) float64 {
	sum := float64(a + b)
	bVirtual := float64(sum - a)
	remainder := float64(float64(a-float64(sum-bVirtual)) + float64(b-bVirtual))
	if sum > 0 && remainder < 0 || sum < 0 && remainder > 0 {
		return math.Nextafter(sum, 0)
	}
	return sum
}

//go:noinline
func pointOnLineMulChop53_57C8A0(a, b float64) float64 {
	product := float64(a * b)
	// FMA recovers the exact product residual only; it does not fuse any
	// of the original projection's separate multiply/add instructions.
	remainder := math.FMA(a, b, -product)
	if product > 0 && remainder < 0 || product < 0 && remainder > 0 {
		return math.Nextafter(product, 0)
	}
	return product
}

//go:noinline
func pointOnLineDivChop53_57C8A0(a, b float64) float64 {
	quotient := float64(a / b)
	// The sign of (a - quotient*b)/b identifies a nearest-even quotient
	// rounded away from zero. The bounded residual cannot underflow.
	remainder := math.FMA(-quotient, b, a)
	if remainder != 0 && !math.IsNaN(remainder) &&
		(quotient > 0 && math.Signbit(remainder) != math.Signbit(b) || quotient < 0 && math.Signbit(remainder) == math.Signbit(b)) {
		return math.Nextafter(quotient, 0)
	}
	return quotient
}

func pointOnLineSpill57C8A0(value float64) float32 {
	stored := float32(value)
	// Includes chopped subnormal stores and finite overflow (max-finite,
	// not infinity). Exact infinities and NaNs retain their normal stores.
	if math.Abs(float64(stored)) > math.Abs(value) {
		stored = math.Nextafter32(stored, 0)
	}
	return stored
}
