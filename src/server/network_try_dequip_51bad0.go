package server

const networkTryDequipPacketSize51BAD0 = 3

// networkTryDequipHooks51BAD0 exposes every ordered read and call in the
// MSG_TRY_DEQUIP branch at GAME.EXE 0051BD43..0051BDC8. Object identities stay
// in their native pointer domain while only the three-byte packet is fixed.
type networkTryDequipHooks51BAD0[O comparable, U, P any] struct {
	loadWireCode     func() uint16
	dynamicUnitCode  func(uint16) uint32
	netDebug         func() bool
	testHighBit      func(uint16)
	loadPlayer       func(U) P
	loadPlayerStatus func(P) uint32
	findItemByCode   func(O, uint32) O
	loadState        func(U) uint8
	loadItemClass    func(O) uint32
	loadItemSubclass func(O) uint32
	tryDequip        func(O, O)
}

// networkTryDequip51BAD0 preserves the original three-byte packet branch. The
// state-one gate protects quest items whose class has 0x01000000 and whose
// subclass low byte has bit 0x08. Its class and subclass reads are deliberately
// short-circuited in the same order as GAME.EXE.
func networkTryDequip51BAD0[O comparable, U, P any](
	unit O,
	update U,
	hooks networkTryDequipHooks51BAD0[O, U, P],
) int32 {
	wireCode := hooks.loadWireCode()
	code := hooks.dynamicUnitCode(wireCode)
	if hooks.netDebug() {
		hooks.testHighBit(wireCode)
	}
	player := hooks.loadPlayer(update)
	if hooks.loadPlayerStatus(player)&0x3 != 0 {
		return networkTryDequipPacketSize51BAD0
	}
	item := hooks.findItemByCode(unit, code)
	var zero O
	if item == zero {
		return networkTryDequipPacketSize51BAD0
	}
	if hooks.loadState(update) == 1 &&
		hooks.loadItemClass(item)&0x01000000 != 0 &&
		uint8(hooks.loadItemSubclass(item))&0x08 != 0 {
		return networkTryDequipPacketSize51BAD0
	}
	hooks.tryDequip(unit, item)
	return networkTryDequipPacketSize51BAD0
}
