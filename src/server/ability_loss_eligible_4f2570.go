package server

// randomAbilityLossEligible4F2570 reconstructs GAME.EXE 004F2570. The
// original uses signed 32-bit comparisons and accepts exactly the five
// nonzero Warrior ability IDs below six.
func randomAbilityLossEligible4F2570(abilityID int32) int32 {
	if abilityID <= 0 || abilityID >= 6 {
		return 0
	}
	return 1
}

// RandomAbilityLossEligible4F2570 exposes the exact fixed-width predicate
// used by random Warrior ability loss.
//
//go:noinline
func RandomAbilityLossEligible4F2570(abilityID int32) int32 {
	return randomAbilityLossEligible4F2570(abilityID)
}
