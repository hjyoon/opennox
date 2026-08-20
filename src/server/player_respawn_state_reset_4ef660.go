package server

const playerRespawnStateResetCoopFlag4EF660 uint32 = 0x800

type playerRespawnStateResetHooks4EF660[O, U, P any] struct {
	loadUpdateData       func(O) U
	storePendingObject   func(U, int, O)
	storeSoulGate        func(U, O)
	loadPlayer           func(U) P
	storeQuestAnkh       func(P, int, O)
	gameFlag             func(uint32) int32
	countGlyphs          func(O) int32
	storeCurTrapsLowByte func(U, uint8)
	storeField66         func(U, uint32)
	storeAttribution     func(O, O)
	resetPlayerMarkers   func(P) P
}

// playerRespawnStateReset4EF660 preserves GAME.EXE 004EF660. The unit's
// UpdateData pointer is cached once and is never guarded. Four pending-object
// pointers and SoulGate are cleared through that cache. In contrast, Player is
// reloaded from the cached UpdateData before every one of the five Quest-Ankh
// stores, so a mutation after one store is observed by the next iteration.
//
// The cooperative flag is read only after those ten Player operations. A zero
// result counts Glyphs in the unit's live inventory and stores only the low
// result byte into CurTraps; a nonzero result preserves the whole CurTraps
// word. Field66 and the unit attribution pointer are then cleared on every
// path. Finally, Player is reloaded once more from the cached UpdateData and
// forwarded to the marker reset helper, whose exact return is propagated.
func playerRespawnStateReset4EF660[O, U, P any](
	unit O,
	hooks playerRespawnStateResetHooks4EF660[O, U, P],
) P {
	var zeroObject O
	update := hooks.loadUpdateData(unit)
	for index := 0; index < 4; index++ {
		hooks.storePendingObject(update, index, zeroObject)
	}
	hooks.storeSoulGate(update, zeroObject)

	for index := 0; index < 5; index++ {
		player := hooks.loadPlayer(update)
		hooks.storeQuestAnkh(player, index, zeroObject)
	}

	if hooks.gameFlag(playerRespawnStateResetCoopFlag4EF660) == 0 {
		count := hooks.countGlyphs(unit)
		hooks.storeCurTrapsLowByte(update, uint8(count))
	}
	hooks.storeField66(update, 0)
	hooks.storeAttribution(unit, zeroObject)

	player := hooks.loadPlayer(update)
	return hooks.resetPlayerMarkers(player)
}
