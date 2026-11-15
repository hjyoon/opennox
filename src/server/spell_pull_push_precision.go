package server

import "math"

// Pull/Push perform FIMUL at startup PC53/ToZero, then an explicit binary32
// store. Reuse the existing local x87 arithmetic and NaN rules without changing
// the host FP environment. Pull's FCHS follows multiplication, including NaNs.
func pullPushForce52BFA0(coefficient float64, level int32, pull bool) float32 {
	product := objectForceMul53_52DF80(coefficient, float64(level))
	if math.IsInf(product, 0) && !math.IsInf(coefficient, 0) {
		// The original x87 exponent range retains this finite product. Only
		// its subsequent ToZero binary32 store is observed, yielding MaxFloat32.
		product = math.Copysign(math.MaxFloat64, product)
	}
	if pull {
		product = -product
	}
	return experienceSpill32TowardZero(product)
}
