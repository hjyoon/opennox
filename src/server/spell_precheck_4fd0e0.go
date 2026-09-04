package server

const (
	spellPrecheckPlayerClass4FD0E0 = uint8(0x04)
	spellPrecheckBadSkill4FD0E0    = int32(9)
	spellPrecheckIllegal4FD0E0     = int32(10)
)

type spellPrecheckHooks4FD0E0[Object, UpdateData, Player any] struct {
	loadSpellArg          func() int32
	spellFlags            func(int32) uint32
	loadUnitArg           func() Object
	findParentPlayer      func(Object) Object
	spellEnabled          func(int32) int32
	loadUnitClassLow      func(Object) uint8
	loadUpdateData        func(Object) UpdateData
	loadPlayer            func(UpdateData) Player
	loadPlayerClass       func(Player) uint8
	checkPlayerSpellClass func(uint8, int32) int32
	summonAllowed         func(int32, Object) int32
}

// spellPrecheck4FD0E0 preserves GAME.EXE 004FD0E0's observable load and
// callback order. The spell argument is cached before the first, discarded
// flags lookup. The unit argument is loaded only after that callback, while
// its Player owner is cached before the enablement callback. Unit class and
// Player data remain live loads after the enablement gate.
//
// The Player branch returns the class check verbatim. The non-Player branch
// canonicalizes every nonzero summon result to zero and zero to ten, exactly
// matching the original neg/sbb/and/add sequence. Deliberately omitted nil
// guards preserve the original disabled-spell and enabled-spell fault
// boundaries.
func spellPrecheck4FD0E0[Object, UpdateData, Player any](
	hooks spellPrecheckHooks4FD0E0[Object, UpdateData, Player],
) int32 {
	spellID := hooks.loadSpellArg()
	_ = hooks.spellFlags(spellID)
	unit := hooks.loadUnitArg()
	owner := hooks.findParentPlayer(unit)
	if hooks.spellEnabled(spellID) == 0 {
		return spellPrecheckIllegal4FD0E0
	}
	if hooks.loadUnitClassLow(unit)&spellPrecheckPlayerClass4FD0E0 != 0 {
		update := hooks.loadUpdateData(unit)
		player := hooks.loadPlayer(update)
		class := hooks.loadPlayerClass(player)
		return hooks.checkPlayerSpellClass(class, spellID)
	}
	if hooks.summonAllowed(spellID, owner) != 0 {
		return 0
	}
	return spellPrecheckIllegal4FD0E0
}
