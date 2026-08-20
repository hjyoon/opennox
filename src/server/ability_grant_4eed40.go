package server

type abilityGivePlayerAllHooks4EED40[O comparable, U, P any] struct {
	loadUnitArg       func() O
	loadUpdateData    func(O) U
	loadCountLow      func() int8
	loadPlayer        func(U) P
	loadAbilityID     func(int32) uint32
	gameFlagsCheck    func(uint32) int32
	isQuest           func() int32
	questMode         func() int32
	loadRewardArg     func() int32
	rewardAbility     func(O, int32, int32)
	storeAbilityLevel func(P, int32, uint32)
}

// abilityGivePlayerAll4EED40 preserves GAME.EXE 004EED40. A non-nil unit
// loads UpdateData, the signed low count byte, and Player in that exact order;
// even a non-positive count therefore performs both pointer-bearing loads.
// Each iteration reads its live ability-table entry. Zero entries skip every
// callback and store. Nonzero entries short-circuit three mode queries; any
// nonzero result clears the same-index ability slot through the Player cached
// before the loop. Otherwise the reward argument is loaded, the ability ID is
// reloaded, and the reward callback receives the cached unit. There are no
// UpdateData or Player nil guards.
func abilityGivePlayerAll4EED40[O comparable, U, P any](
	hooks abilityGivePlayerAllHooks4EED40[O, U, P],
) {
	unit := hooks.loadUnitArg()
	var nilObject O
	if unit == nilObject {
		return
	}
	update := hooks.loadUpdateData(unit)
	count := hooks.loadCountLow()
	player := hooks.loadPlayer(update)
	if count <= 0 {
		return
	}
	for index := int32(0); index < int32(count); index++ {
		if hooks.loadAbilityID(index) == 0 {
			continue
		}
		if hooks.gameFlagsCheck(0x1000) != 0 || hooks.isQuest() != 0 || hooks.questMode() != 0 {
			hooks.storeAbilityLevel(player, index, 0)
			continue
		}
		rewardArg := hooks.loadRewardArg()
		ability := hooks.loadAbilityID(index)
		hooks.rewardAbility(unit, int32(ability), rewardArg)
	}
}
