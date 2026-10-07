package opennox

import "math"

// GAME.EXE 004031F2 selects 53-bit x87 precision and 0043E2C1 selects
// round-toward-zero. Fireball's binary32 kinematics and bounded balance
// coefficients keep these operations inside binary64's exponent range.
// Correct their rounding without changing the calling thread's FPU.
//
//go:noinline
func fireballAddChop53_52C790(a, b float64) float64 {
	sum := float64(a + b)
	bVirtual := float64(sum - a)
	remainder := float64(float64(a-float64(sum-bVirtual)) + float64(b-bVirtual))
	if sum > 0 && remainder < 0 || sum < 0 && remainder > 0 {
		return math.Nextafter(sum, 0)
	}
	return sum
}

//go:noinline
func fireballMulChop53_52C790(a, b float64) float64 {
	product := float64(a * b)
	// FMA recovers the product residual only; the original separate
	// multiply/add operations in the cast are never fused together.
	remainder := math.FMA(a, b, -product)
	if product > 0 && remainder < 0 || product < 0 && remainder > 0 {
		return math.Nextafter(product, 0)
	}
	return product
}

func fireballSpill32_52C790(value float64) float32 {
	stored := float32(value)
	// Includes chopped subnormals and finite overflow. Exact infinities
	// and NaNs retain their normal stores.
	if math.Abs(float64(stored)) > math.Abs(value) {
		stored = math.Nextafter32(stored, 0)
	}
	return stored
}
