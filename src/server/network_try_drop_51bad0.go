package server

import "github.com/opennox/libs/types"

const networkTryDropPacketSize51BAD0 = 7

// networkTryDropHooks51BAD0 exposes the reads and calls made by the
// MSG_TRY_DROP branch inside GAME.EXE 0051BAD0. Keeping these operations lazy
// preserves the original gate and packet-read order for oracle tests.
type networkTryDropHooks51BAD0[O comparable, U, P any] struct {
	loadWireCode     func() uint16
	dynamicUnitCode  func(uint16) uint32
	netDebug         func() bool
	testHighBit      func(uint16)
	loadPlayer       func(U) P
	loadPlayerStatus func(P) uint32
	loadTradeActive  func(U) bool
	loadDialogActive func(U) bool
	loadUnitFlagsLow func(O) uint8
	findItemByCode   func(O, uint32) O
	loadDestinationX func() uint16
	loadDestinationY func() uint16
	drop             func(O, O, *types.Pointf)
}

// networkTryDrop51BAD0 preserves the seven-byte MSG_TRY_DROP branch from
// GAME.EXE 0051BDC9..0051BE6C. Object identities are never converted to the
// original PE32 integer slots.
func networkTryDrop51BAD0[O comparable, U, P any](
	unit O,
	update U,
	hooks networkTryDropHooks51BAD0[O, U, P],
) int32 {
	wireCode := hooks.loadWireCode()
	code := hooks.dynamicUnitCode(wireCode)
	if hooks.netDebug() {
		hooks.testHighBit(wireCode)
	}

	player := hooks.loadPlayer(update)
	if hooks.loadPlayerStatus(player)&0x3 != 0 ||
		hooks.loadTradeActive(update) ||
		hooks.loadDialogActive(update) ||
		hooks.loadUnitFlagsLow(unit)&0x2 != 0 {
		return networkTryDropPacketSize51BAD0
	}

	item := hooks.findItemByCode(unit, code)
	var zero O
	if item == zero {
		return networkTryDropPacketSize51BAD0
	}
	point := types.Pointf{
		X: float32(hooks.loadDestinationX()),
		Y: float32(hooks.loadDestinationY()),
	}
	hooks.drop(unit, item, &point)
	return networkTryDropPacketSize51BAD0
}
