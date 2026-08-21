package server

const (
	playerUnitInitQuestFlag4EFE80     = uint32(0x1000)
	playerUnitInitReadReward4EFE80    = int32(0)
	playerUnitInitRestoreStats4EFE80  = int32(1)
	playerUnitInitKeepItems4EFE80     = int32(0)
	playerUnitInitExtraLivesKey4EFE80 = "QuestGameStartingExtraLives"
)

type playerUnitInitHooks4EFE80[O, U, P comparable] struct {
	loadUnitArg           func() O
	loadUpdateData        func(O) U
	getGold               func(O) uint32
	subGold               func(O, uint32)
	syncLevel             func(O)
	loadPlayer            func(U) P
	awardBeastScrolls     func(P)
	awardSpells           func(P)
	readValues            func(O, int32)
	awardWarriorAbilities func(P)
	gameFlag              func(uint32) int32
	balanceFloat          func(string) float32
	floatToInt            func(float32) int32
	storeExtraLives       func(U, int32)
	makeDefaultItems      func(O, int32, int32) uint8
}

// playerUnitInit4EFE80 preserves GAME.EXE 004EFE80's exact observable order.
// The unit and its UpdateData pointer are cached at entry before the gold
// callbacks. Player is instead reloaded from that cached UpdateData before
// each of the three award callbacks, including after read-values processing.
//
// QuestGameStartingExtraLives is read only for a nonzero Quest flag. The
// balance result is explicitly binary32 before the float-to-int callback, and
// the converted value is stored through the cached UpdateData. The final
// default-item callback receives literal arguments 1 and 0, and its low byte
// is the function result. No object, UpdateData, or Player nil guard is added.
func playerUnitInit4EFE80[O, U, P comparable](
	h playerUnitInitHooks4EFE80[O, U, P],
) uint8 {
	unit := h.loadUnitArg()
	update := h.loadUpdateData(unit)
	gold := h.getGold(unit)
	h.subGold(unit, gold)
	h.syncLevel(unit)

	player := h.loadPlayer(update)
	h.awardBeastScrolls(player)
	player = h.loadPlayer(update)
	h.awardSpells(player)
	h.readValues(unit, playerUnitInitReadReward4EFE80)
	player = h.loadPlayer(update)
	h.awardWarriorAbilities(player)

	if h.gameFlag(playerUnitInitQuestFlag4EFE80) != 0 {
		extraLives := h.balanceFloat(playerUnitInitExtraLivesKey4EFE80)
		h.storeExtraLives(update, h.floatToInt(extraLives))
	}
	return h.makeDefaultItems(
		unit,
		playerUnitInitRestoreStats4EFE80,
		playerUnitInitKeepItems4EFE80,
	)
}
