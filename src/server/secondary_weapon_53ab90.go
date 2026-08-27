package server

import "github.com/opennox/libs/player"

type secondaryWeaponHooks53AB90[O comparable, U, P any] struct {
	zero            O
	loadUpdate      func(O) U
	loadPlayer      func(U) P
	loadPlayerClass func(P) player.Class
	classCanUseItem func(O, player.Class) bool
	checkStrength   func(O, O) bool
	loadPlayerIndex func(P) byte
	clearClient     func(byte)
	store           func(U, O)
}

// secondaryWeaponReport53AB90 reconstructs GAME.EXE 0053AB90. The owner
// update pointer is cached before the item nil check. Invalid items reload the
// player pointer for the clear report, and the reported item is stored only
// after both validation callbacks and any clear report have completed.
func secondaryWeaponReport53AB90[O comparable, U, P any](
	owner, item O,
	h secondaryWeaponHooks53AB90[O, U, P],
) {
	if owner == h.zero {
		return
	}
	update := h.loadUpdate(owner)
	if item != h.zero {
		playerValue := h.loadPlayer(update)
		class := h.loadPlayerClass(playerValue)
		if !h.classCanUseItem(item, class) || !h.checkStrength(owner, item) {
			playerValue = h.loadPlayer(update)
			h.clearClient(h.loadPlayerIndex(playerValue))
		}
	}
	h.store(update, item)
}
