package server

import "math"

// GAME.EXE runs player stat division at precision 53 toward zero. Correct the
// nearest quotient using its FMA residual without changing host FP state. The
// signed-dword and binary32 operands keep finite quotients in binary64 range.
//
//go:noinline
func playerStatDiv53TowardZero4EEDC0(a, b float64) float64 {
	quotient := a / b
	residual := math.FMA(-quotient, b, a)
	if residual == 0 || math.IsNaN(residual) || math.IsInf(quotient, 0) || math.IsNaN(quotient) {
		return quotient
	}
	negativeError := math.Signbit(residual) != math.Signbit(b)
	if (quotient > 0 && negativeError) || (quotient < 0 && !negativeError) {
		return math.Nextafter(quotient, 0)
	}
	return quotient
}
