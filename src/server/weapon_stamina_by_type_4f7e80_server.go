package server

// WeaponStaminaByType4F7E80 exposes the native GAME.EXE 004F7E80 contract.
// Both the input flag word and return value stay exactly 32 bits at every Go
// and C boundary.
func WeaponStaminaByType4F7E80(flags uint32) int32 {
	return weaponStaminaByType4F7E80(flags)
}
