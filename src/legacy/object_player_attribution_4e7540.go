package legacy

const playerClassBit4E7540 uint8 = 4

type playerAttributionHooks4E7540[O comparable, U, P any] struct {
	class          func(O) uint32
	updateData     func(O) U
	player         func(U) P
	playerIndex    func(P) byte
	setPlayerIndex func(P, uint32)
	frame          func() uint32
	setFrame       func(P, uint32)
	setPending     func(P, uint32)
}

// recordPlayerAttribution4E7540 is the pointer-width-independent contract for
// GAME.EXE 004E7540. The target player pointer is deliberately reloaded before
// each store, and the frame is read after the second reload.
func recordPlayerAttribution4E7540[O comparable, U, P any](
	source, target O,
	h playerAttributionHooks4E7540[O, U, P],
) {
	var nilObject O
	if source == nilObject || target == nilObject {
		return
	}
	if uint8(h.class(source))&playerClassBit4E7540 == 0 {
		return
	}
	if uint8(h.class(target))&playerClassBit4E7540 == 0 {
		return
	}
	if source == target {
		return
	}

	sourceUpdate := h.updateData(source)
	targetUpdate := h.updateData(target)
	sourcePlayer := h.player(sourceUpdate)
	index := uint32(h.playerIndex(sourcePlayer))

	targetPlayer := h.player(targetUpdate)
	h.setPlayerIndex(targetPlayer, index)
	targetPlayer = h.player(targetUpdate)
	frame := h.frame()
	h.setFrame(targetPlayer, frame)
	targetPlayer = h.player(targetUpdate)
	h.setPending(targetPlayer, 1)
}
