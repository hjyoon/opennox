package server

import "github.com/opennox/libs/types"

const networkInventoryFailPacketSize51BAD0 = 3

type networkInventoryFailHooks51BAD0[O comparable] struct {
	loadCode      func() uint16
	findItem      func(O, uint32) O
	loadPosition  func(O) *types.Pointf
	drop          func(O, O, *types.Pointf)
	carryingHeavy func(O)
}

// networkInventoryFail51BAD0 preserves GAME.EXE 0051C7EA..0051C837. The
// packet uses its raw sixteen-bit object code rather than the dynamic-code
// decoder, and both the inventory result and drop arguments remain native
// pointer-width values.
func networkInventoryFail51BAD0[O comparable](
	unit O,
	hooks networkInventoryFailHooks51BAD0[O],
) int32 {
	item := hooks.findItem(unit, uint32(hooks.loadCode()))
	var zero O
	if item == zero {
		return networkInventoryFailPacketSize51BAD0
	}
	hooks.drop(unit, item, hooks.loadPosition(unit))
	hooks.carryingHeavy(unit)
	return networkInventoryFailPacketSize51BAD0
}
