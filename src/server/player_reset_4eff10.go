package server

const (
	playerResetObjectFlagMask4EFF10 = uint32(0xffeb3fe7)
	playerResetReadReward4EFF10     = int32(0)
	playerResetLevel4EFF10          = uint8(1)
	playerResetState4EFF10          = PlayerState13
	playerResetMarker4EFF10         = uint32(0xdeadface)
	playerResetResult4EFF10         = int32(-559023410)
)

type playerResetHooks4EFF10[O, U, P comparable] struct {
	loadUnitArg           func() O
	loadUpdateData        func(O) U
	loadPlayer            func(U) P
	awardBeastScrolls     func(P)
	awardSpells           func(P)
	storePlayerLevel      func(P, uint8)
	cancelAbilities       func(O)
	readValues            func(O, int32)
	awardWarriorAbilities func(P)
	loadManaMaximum       func(U) uint16
	storeManaCurrent      func(U, uint16)
	storeManaPrevious     func(U, uint16)
	loadManaToken         func(P) uint32
	protectMana           func(uint32, uint16)
	storeTrapSpell        func(U, int, uint32)
	storeTrapCountLow     func(U, uint8)
	setHealthMaximum      func(O)
	loadObjectFlags       func(O) uint32
	storeObjectField541   func(O, uint8)
	storeObjectFlags      func(O, uint32)
	setPlayerState        func(O, PlayerState)
	clearBuffs            func(O)
	cancelSpells          func(O)
	removePoison          func(O)
	resetPlayerRuntime    func(O)
	loadPlayerIndex       func(P) uint8
	reportTotalHealth     func(uint8, O)
	reportTotalMana       func(uint8, O)
	storeObject130        func(O, O)
	storePlayerMarker3664 func(P, uint32)
	storePlayerMarker3660 func(P, uint32)
}

// playerReset4EFF10 preserves GAME.EXE 004EFF10's exact observable order.
// The unit and its UpdateData pointer are cached at entry. Player is instead
// reloaded from that cached UpdateData before every original Player access:
// four award/level operations, mana protection, two reports, and the two
// final marker stores all receive their own live load.
//
// ManaMax is loaded before the fifth Player. ManaCur and ManaPrev are stored
// before that cached Player's protection token is read. TrapSpellsCnt and the
// unit's field 541 are low-byte stores, object flags use the exact 32-bit
// mask, and Obj130 is a native-width nil store. No nil guard is added.
func playerReset4EFF10[O, U, P comparable](
	h playerResetHooks4EFF10[O, U, P],
) int32 {
	unit := h.loadUnitArg()
	update := h.loadUpdateData(unit)

	player := h.loadPlayer(update)
	h.awardBeastScrolls(player)
	player = h.loadPlayer(update)
	h.awardSpells(player)
	player = h.loadPlayer(update)
	h.storePlayerLevel(player, playerResetLevel4EFF10)
	h.cancelAbilities(unit)
	h.readValues(unit, playerResetReadReward4EFF10)
	player = h.loadPlayer(update)
	h.awardWarriorAbilities(player)

	manaMaximum := h.loadManaMaximum(update)
	player = h.loadPlayer(update)
	h.storeManaCurrent(update, manaMaximum)
	h.storeManaPrevious(update, manaMaximum)
	manaToken := h.loadManaToken(player)
	h.protectMana(manaToken, manaMaximum)

	for index := 0; index < 5; index++ {
		h.storeTrapSpell(update, index, 0)
	}
	h.storeTrapCountLow(update, 0)
	h.setHealthMaximum(unit)

	flags := h.loadObjectFlags(unit)
	h.storeObjectField541(unit, 0)
	h.storeObjectFlags(unit, flags&playerResetObjectFlagMask4EFF10)
	h.setPlayerState(unit, playerResetState4EFF10)
	h.clearBuffs(unit)
	h.cancelSpells(unit)
	h.removePoison(unit)
	h.resetPlayerRuntime(unit)

	player = h.loadPlayer(update)
	h.reportTotalHealth(h.loadPlayerIndex(player), unit)
	player = h.loadPlayer(update)
	h.reportTotalMana(h.loadPlayerIndex(player), unit)

	var zeroObject O
	h.storeObject130(unit, zeroObject)
	player = h.loadPlayer(update)
	h.storePlayerMarker3664(player, playerResetMarker4EFF10)
	player = h.loadPlayer(update)
	h.storePlayerMarker3660(player, playerResetMarker4EFF10)
	return playerResetResult4EFF10
}
