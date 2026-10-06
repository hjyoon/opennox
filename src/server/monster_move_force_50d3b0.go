package server

import "math"

// These local operations model precision-53 / ToZero x87 in the original
// gameplay environment. Finite binary32 path/speed inputs bound the retained
// products, quotients and exact residuals inside binary64's exponent range.
// FMA recovers a residual only; it does not fuse original FMUL/FDIV/FSQRT
// with another operation or change the calling thread's floating-point mode.
//
//go:noinline
func monsterMoveForceMulChop53_50D4FE(a, b float64) float64 {
	product := float64(a * b)
	residual := math.FMA(a, b, -product)
	if product > 0 && residual < 0 || product < 0 && residual > 0 {
		return math.Nextafter(product, 0)
	}
	return product
}

//go:noinline
func monsterMoveForceDivChop53_50D581(a, b float64) float64 {
	quotient := float64(a / b)
	residual := math.FMA(-quotient, b, a)
	// a-q*b has the opposite sign to a exactly when nearest-even rounded
	// away from zero. Keep nonfinite arithmetic's propagation unchanged.
	if residual != 0 && !math.IsNaN(residual) && math.Signbit(residual) != math.Signbit(a) {
		return math.Nextafter(quotient, 0)
	}
	return quotient
}

//go:noinline
func monsterMoveForceSqrtChop53_50D50C(value float64) float64 {
	root := math.Sqrt(value)
	if math.FMA(root, root, -value) > 0 {
		return math.Nextafter(root, 0)
	}
	return root
}
