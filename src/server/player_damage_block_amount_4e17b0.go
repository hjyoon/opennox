package server

// Original shield/weapon blocking multiplies the balance double by signed
// damage at startup PC53, then spills binary32 before the wear callback.
func playerDamageBlockAmount4E17B0(percent float64, damage int32) float32 {
	return experienceSpill32TowardZero(experienceMul53TowardZero(percent, float64(damage)))
}
