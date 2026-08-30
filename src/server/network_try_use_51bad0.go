package server

const networkTryUsePacketSize51BAD0 = 3

type networkTryUseHooks51BAD0[O comparable, U, P any] struct {
	loadWireCode     func() uint16
	dynamicUnitCode  func(uint16) uint32
	netDebug         func() bool
	testHighBit      func(uint16)
	gameBlocked      func() bool
	loadPlayer       func(U) P
	loadPlayerStatus func(P) uint32
	findItemByCode   func(O, uint32) O
	useByNetCode     func(O, O)
}

// networkTryUse51BAD0 preserves GAME.EXE 0051BF08..0051BF7C. Packet decoding
// and debug observation precede the game/player gates; the item identity stays
// pointer-width and a rejected request still consumes exactly three bytes.
func networkTryUse51BAD0[O comparable, U, P any](
	unit O,
	update U,
	hooks networkTryUseHooks51BAD0[O, U, P],
) int32 {
	wireCode := hooks.loadWireCode()
	code := hooks.dynamicUnitCode(wireCode)
	if hooks.netDebug() {
		hooks.testHighBit(wireCode)
	}
	if hooks.gameBlocked() {
		return networkTryUsePacketSize51BAD0
	}
	player := hooks.loadPlayer(update)
	if hooks.loadPlayerStatus(player)&0x3 != 0 {
		return networkTryUsePacketSize51BAD0
	}
	item := hooks.findItemByCode(unit, code)
	var zero O
	if item != zero {
		hooks.useByNetCode(unit, item)
	}
	return networkTryUsePacketSize51BAD0
}
