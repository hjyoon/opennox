package server

const (
	warriorAbilityAwardAllWarrior4EFE10    = uint8(0)
	warriorAbilityAwardAllAdminMask4EFE10  = uint8(0x10)
	warriorAbilityAwardAllFirstIndex4EFE10 = int32(1)
	warriorAbilityAwardAllEndIndex4EFE10   = int32(6)
	warriorAbilityAwardAllAdminLevel4EFE10 = int32(5)
)

type warriorAbilityAwardAllHooks4EFE10[P comparable] struct {
	loadPlayerArg     func() P
	loadPlayerClass   func(P) uint8
	loadEngineFlags   func() uint8
	storeAbilityLevel func(P, int32, uint32)
	loadProtection    func(P) uint32
	awardProtection   func(uint32, int32, int32)
}

// warriorAbilityAwardAll4EFE10 preserves GAME.EXE 004EFE10. It loads the
// Player class before every other field or global and returns immediately for
// every non-warrior value. A warrior reads the low engine-flags byte once and
// fixes the selected level at five for Admin or zero otherwise.
//
// Each ability index 1..5 is stored before a live protection-token reload and
// award callback. The Player argument and Admin decision remain cached across
// callbacks. Level zero is never written, protection is neither reset nor
// applied here, and there is no nil guard.
func warriorAbilityAwardAll4EFE10[P comparable](hooks warriorAbilityAwardAllHooks4EFE10[P]) {
	player := hooks.loadPlayerArg()
	if hooks.loadPlayerClass(player) != warriorAbilityAwardAllWarrior4EFE10 {
		return
	}

	flags := hooks.loadEngineFlags()
	level := int32(0)
	if flags&warriorAbilityAwardAllAdminMask4EFE10 != 0 {
		level = warriorAbilityAwardAllAdminLevel4EFE10
	}
	for index := warriorAbilityAwardAllFirstIndex4EFE10; index < warriorAbilityAwardAllEndIndex4EFE10; index++ {
		hooks.storeAbilityLevel(player, index, uint32(level))
		protection := hooks.loadProtection(player)
		hooks.awardProtection(protection, index, level)
	}
}
