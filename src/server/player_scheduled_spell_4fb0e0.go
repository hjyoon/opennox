package server

const playerScheduledSpellFizzleSound4FB0E0 = int32(231)

// playerScheduledSpellArg4FB0E0 is the layout-independent form of the
// 12-byte PE32 spell argument assembled on the stack by GAME.EXE. The native
// adapter converts it to SpellAcceptArg without narrowing Object pointers.
type playerScheduledSpellArg4FB0E0[O comparable] struct {
	target O
	posX   float32
	posY   float32
}

// playerScheduledSpellHooks4FB0E0 exposes each observable load, store, and
// call in GAME.EXE 004FB0E0 and 004FB1D0. In particular, TrapSpellsCnt is a
// byte access even though the native field is represented by a uint32.
type playerScheduledSpellHooks4FB0E0[O, U, P comparable] struct {
	loadUpdateData func(O) U
	loadCountLow   func(U) uint8
	loadSpell      func(U, int) uint32
	checkSpell     func(O, uint32, int32) int32
	loadPosX       func(U) int32
	loadPosY       func(U) int32
	loadPlayer     func(U) P
	loadPlayerInd  func(P) uint8
	informText     func(uint8, uint8, int32)
	audioEvent     func(int32, O, int32, int32)
	castSpell      func(uint32, O, playerScheduledSpellArg4FB0E0[O])
	storeSpell     func(U, int, uint32)
	storeCountLow  func(U, uint8)
}

// playerDoScheduledSpell4FB0E0 preserves GAME.EXE 004FB0E0. It consumes the
// oldest queued spell even when casting is rejected, shifts the remaining
// queue left, clears the final used slot, and changes only the low byte of the
// count. Loads repeated by the original machine code remain live here: the
// spell is reloaded after the rejection check and the count is reloaded while
// shifting and before its final decrement.
func playerDoScheduledSpell4FB0E0[O, U, P comparable](
	unit O,
	target O,
	h playerScheduledSpellHooks4FB0E0[O, U, P],
) int32 {
	update := h.loadUpdateData(unit)
	if h.loadCountLow(update) == 0 {
		return 0
	}

	rejection := h.checkSpell(unit, h.loadSpell(update, 0), 0)
	arg := playerScheduledSpellArg4FB0E0[O]{
		target: target,
		posX:   float32(h.loadPosX(update)),
		posY:   float32(h.loadPosY(update)),
	}
	if rejection != 0 {
		player := h.loadPlayer(update)
		h.informText(h.loadPlayerInd(player), 0, rejection)
		h.audioEvent(playerScheduledSpellFizzleSound4FB0E0, unit, 0, 0)
	} else {
		h.castSpell(h.loadSpell(update, 0), unit, arg)
	}

	clearOrdinal := 1
	if h.loadCountLow(update) > 1 {
		destination := 0
		for {
			next := h.loadSpell(update, destination+1)
			clearOrdinal++
			h.storeSpell(update, destination, next)
			destination++
			if clearOrdinal >= int(h.loadCountLow(update)) {
				break
			}
		}
	}
	h.storeSpell(update, clearOrdinal-1, 0)
	count := h.loadCountLow(update)
	h.storeCountLow(update, count-1)
	return 1
}

// playerDoScheduledSpellQueue4FB1D0 preserves GAME.EXE 004FB1D0. It consumes
// the newest queued spell without clearing its slot. On the successful path,
// both the count and selected spell are deliberately reloaded after the cast
// check; callbacks in the original executable can mutate either value.
func playerDoScheduledSpellQueue4FB1D0[O, U, P comparable](
	unit O,
	target O,
	h playerScheduledSpellHooks4FB0E0[O, U, P],
) int32 {
	update := h.loadUpdateData(unit)
	count := h.loadCountLow(update)
	if count == 0 {
		return 0
	}

	rejection := h.checkSpell(unit, h.loadSpell(update, int(count)-1), 0)
	arg := playerScheduledSpellArg4FB0E0[O]{
		target: target,
		posX:   float32(h.loadPosX(update)),
		posY:   float32(h.loadPosY(update)),
	}
	if rejection != 0 {
		player := h.loadPlayer(update)
		h.informText(h.loadPlayerInd(player), 0, rejection)
		h.audioEvent(playerScheduledSpellFizzleSound4FB0E0, unit, 0, 0)
	} else {
		count = h.loadCountLow(update)
		h.castSpell(h.loadSpell(update, int(count)-1), unit, arg)
	}

	count = h.loadCountLow(update)
	h.storeCountLow(update, count-1)
	return 1
}
