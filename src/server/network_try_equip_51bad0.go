package server

const networkTryEquipPacketSize51BAD0 = 3

// networkTryEquipHooks51BAD0 exposes the reads and calls made by the
// MSG_TRY_EQUIP branch at GAME.EXE 0051BCCD..0051BD42. The hooks stay lazy so
// every early return preserves the original read and callback order.
type networkTryEquipHooks51BAD0[O comparable, U, P any] struct {
	loadWireCode     func() uint16
	dynamicUnitCode  func(uint16) uint32
	netDebug         func() bool
	testHighBit      func(uint16)
	gameBlocked      func() bool
	loadPlayer       func(U) P
	loadPlayerStatus func(P) uint32
	findItemByCode   func(O, uint32) O
	tryEquip         func(O, O)
}

// networkTryEquip51BAD0 preserves the three-byte MSG_TRY_EQUIP branch from
// GAME.EXE. Object identities remain native-width values throughout.
func networkTryEquip51BAD0[O comparable, U, P any](
	unit O,
	update U,
	hooks networkTryEquipHooks51BAD0[O, U, P],
) int32 {
	wireCode := hooks.loadWireCode()
	code := hooks.dynamicUnitCode(wireCode)
	if hooks.netDebug() {
		hooks.testHighBit(wireCode)
	}
	if hooks.gameBlocked() {
		return networkTryEquipPacketSize51BAD0
	}

	player := hooks.loadPlayer(update)
	if hooks.loadPlayerStatus(player)&0x3 != 0 {
		return networkTryEquipPacketSize51BAD0
	}
	item := hooks.findItemByCode(unit, code)
	var zero O
	if item != zero {
		hooks.tryEquip(unit, item)
	}
	return networkTryEquipPacketSize51BAD0
}
