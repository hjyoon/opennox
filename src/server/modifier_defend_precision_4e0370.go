package server

// ModifierScalarMultiply32 applies the original PC53 product and binary32
// store with the startup rounding mode.
func ModifierScalarMultiply32(value, coefficient float32) float32 {
	product := experienceMul53TowardZero(float64(coefficient), float64(value))
	return experienceSpill32TowardZero(product)
}

// ModifierArmorMultiplier4E0370 applies the armor coefficient to the value.
func ModifierArmorMultiplier4E0370(value, coefficient float32) float32 {
	return ModifierScalarMultiply32(value, coefficient)
}

// ModifierDurabilityMultiplier4E0380 subtracts from one at 004E038A, then
// separately adds one at 004E0391. Each step rounds toward zero at PC53;
// replacing them with 2-coefficient changes small and large coefficients.
func ModifierDurabilityMultiplier4E0380(value, coefficient float32) float32 {
	difference := experienceAdd53TowardZero(1, -float64(coefficient))
	multiplier := experienceAdd53TowardZero(difference, 1)
	product := experienceMul53TowardZero(multiplier, float64(value))
	return experienceSpill32TowardZero(product)
}

// ModifierIntegerAmount32 returns the low 32 bits of the original signed-qword
// conversion after scaling at PC53. Invalid conversions produce zero.
func ModifierIntegerAmount32(value int32, coefficient float32) int32 {
	product := experienceMul53TowardZero(float64(value), float64(coefficient))
	return x87TruncSignedQwordLow566DCC(product)
}
