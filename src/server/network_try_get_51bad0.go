package server

const networkTryGetPacketSize51BAD0 = 3

// networkTryGetHooks51BAD0 exposes the reads and calls made by the
// MSG_TRY_GET branch at GAME.EXE 0051BE6D..0051BF6A. The hooks stay lazy so every
// early return keeps the original fault and callback order.
type networkTryGetHooks51BAD0[O comparable, U, P any] struct {
	loadWireCode       func() uint16
	dynamicUnitCode    func(uint16) uint32
	netDebug           func() bool
	testHighBit        func(uint16)
	gameBlocked        func() bool
	loadPlayer         func(U) P
	loadPlayerStatus   func(P) uint32
	loadTradeActive    func(U) bool
	loadDialogActive   func(U) bool
	loadUnitFlagsLow   func(O) uint8
	objectFromNetCode  func(uint32) O
	loadInventoryFirst func(O) O
	loadInventoryNext  func(O) O
	loadWeight         func(O) uint8
	loadCarryCapacity  func(O) uint16
	pickup             func(O, O)
	carryingTooMuch    func(O)
}

// networkTryGet51BAD0 preserves the three-byte MSG_TRY_GET branch from
// GAME.EXE 0051BE6D..0051BF6A. Object identities remain native-width values
// throughout.
func networkTryGet51BAD0[O comparable, U, P any](
	unit O,
	update U,
	hooks networkTryGetHooks51BAD0[O, U, P],
) int32 {
	wireCode := hooks.loadWireCode()
	code := hooks.dynamicUnitCode(wireCode)
	if hooks.netDebug() {
		hooks.testHighBit(wireCode)
	}
	if hooks.gameBlocked() {
		return networkTryGetPacketSize51BAD0
	}

	player := hooks.loadPlayer(update)
	if hooks.loadPlayerStatus(player)&0x3 != 0 ||
		hooks.loadTradeActive(update) ||
		hooks.loadDialogActive(update) ||
		hooks.loadUnitFlagsLow(unit)&0x2 != 0 {
		return networkTryGetPacketSize51BAD0
	}

	item := hooks.objectFromNetCode(code)
	var zero O
	if item == zero {
		return networkTryGetPacketSize51BAD0
	}
	weight := uint32(0)
	for it := hooks.loadInventoryFirst(unit); it != zero; it = hooks.loadInventoryNext(it) {
		weight += uint32(hooks.loadWeight(it))
	}
	if weight+uint32(hooks.loadWeight(item)) <= uint32(hooks.loadCarryCapacity(unit)) {
		hooks.pickup(unit, item)
	} else {
		hooks.carryingTooMuch(unit)
	}
	return networkTryGetPacketSize51BAD0
}
