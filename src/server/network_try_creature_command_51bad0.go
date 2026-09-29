package server

const networkTryCreatureCommandPacketSize51BAD0 = 4

// networkTryCreatureCommandHooks51BAD0 exposes every ordered read and call in
// the MSG_TRY_CREATURE_COMMAND branch at GAME.EXE 0051BED7..0051BF66. Object
// identities remain native-width handles; only packet fields use fixed-width
// integers.
type networkTryCreatureCommandHooks51BAD0[O comparable, U, P any] struct {
	loadWireCode      func() uint16
	dynamicUnitCode   func(uint16) uint32
	netDebug          func() bool
	testHighBit       func(uint16)
	loadPlayer        func(U) P
	loadPlayerStatus  func(P) uint32
	objectFromNetCode func(uint32) O
	loadOrder         func() uint8
	orderUnit         func(O, O, uint32)
}

// networkTryCreatureCommand51BAD0 preserves the original four-byte packet
// branch, including its lookup and short-circuit order. A zero wire code
// applies the order to all owned creatures by dispatching a zero creature.
func networkTryCreatureCommand51BAD0[O comparable, U, P any](
	unit O,
	update U,
	hooks networkTryCreatureCommandHooks51BAD0[O, U, P],
) int32 {
	wireCode := hooks.loadWireCode()
	code := hooks.dynamicUnitCode(wireCode)
	if hooks.netDebug() {
		hooks.testHighBit(wireCode)
	}
	player := hooks.loadPlayer(update)
	if hooks.loadPlayerStatus(player)&0x1 != 0 {
		return networkTryCreatureCommandPacketSize51BAD0
	}
	if wireCode == 0 {
		var zero O
		hooks.orderUnit(unit, zero, uint32(hooks.loadOrder()))
		return networkTryCreatureCommandPacketSize51BAD0
	}
	creature := hooks.objectFromNetCode(code)
	var zero O
	if creature != zero {
		hooks.orderUnit(unit, creature, uint32(hooks.loadOrder()))
	}
	return networkTryCreatureCommandPacketSize51BAD0
}
