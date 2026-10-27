package server

// 004E1EE8 multiplies cached absorption by one half at PC53 before the
// subtraction, damage multiplication and binary32 store, all toward zero.
func playerDamageCrushAmount4E17B0(armor float32, damage int32) float32 {
	absorption := experienceMul53TowardZero(float64(armor), .5)
	factor := experienceAdd53TowardZero(1, -absorption)
	return experienceSpill32TowardZero(experienceMul53TowardZero(factor, float64(damage)))
}

// Compute 004E20F0 without storing so callers can keep their existing
// missing-service admission before mutation. The integer remains exact until
// the final carry subtraction and binary32 spill.
func playerDamagePhysicalCarry4E17B0(amount, carry float32) (int32, float32) {
	accumulated := experienceSpill32TowardZero(experienceAdd53TowardZero(float64(amount), float64(carry)))
	effective := netFXTruncateSignedDWORD419A70(accumulated)
	remaining := experienceSpill32TowardZero(experienceAdd53TowardZero(float64(accumulated), -float64(effective)))
	return effective, remaining
}

func playerDamageFullArmorAmount4E17B0(armor float32, damage int32) float32 {
	factor := experienceAdd53TowardZero(1, -float64(armor))
	return experienceSpill32TowardZero(experienceMul53TowardZero(factor, float64(damage)))
}
