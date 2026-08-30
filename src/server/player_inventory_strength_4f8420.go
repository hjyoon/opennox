package server

const playerInventoryEquippedFlag4F8420 = uint32(0x100)

// playerInventoryStrengthHooks4F8420 exposes every observable field load and
// callback in GAME.EXE 004F8420 without assuming the width or layout of an
// Object pointer.
type playerInventoryStrengthHooks4F8420[O comparable] struct {
	loadInventoryHead func(O) O
	loadItemFlags     func(O) uint32
	checkStrength     func(O, O) int32
	forceDrop         func(O, O) int32
	loadInventoryNext func(O) O
}

// playerInventoryStrength4F8420 preserves GAME.EXE 004F8420. The inventory
// head is read unconditionally from player. Each item's full flags dword is
// read once and only bit 0x100 gates the strength callback. A zero strength
// result invokes force-drop; every nonzero 32-bit result skips it. Both
// callback results are otherwise discarded. Most importantly, the current
// item's live next link is loaded after either callback may have mutated it.
func playerInventoryStrength4F8420[O comparable](
	player O,
	hooks playerInventoryStrengthHooks4F8420[O],
) {
	item := hooks.loadInventoryHead(player)
	var zero O
	for item != zero {
		flags := hooks.loadItemFlags(item)
		if flags&playerInventoryEquippedFlag4F8420 != 0 {
			if hooks.checkStrength(player, item) == 0 {
				_ = hooks.forceDrop(player, item)
			}
		}
		item = hooks.loadInventoryNext(item)
	}
}
