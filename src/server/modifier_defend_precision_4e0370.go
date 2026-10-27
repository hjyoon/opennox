package server

// ModifierArmorMultiplier4E0370 retains the complete original leaf's PC53
// product and binary32 store under the startup rounding mode, locally.
func ModifierArmorMultiplier4E0370(value, coefficient float32) float32 {
	product := experienceMul53TowardZero(float64(coefficient), float64(value))
	return experienceSpill32TowardZero(product)
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
