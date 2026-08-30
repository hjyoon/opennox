package server

const networkInfoBookPacketSize51BAD0 = 4

type networkInfoBookHooks51BAD0[O comparable, U any] struct {
	loadWireCode    func() uint16
	dynamicUnitCode func(uint16) uint32
	netDebug        func() bool
	testHighBit     func(uint16)
	findInventory   func(O, uint32) O
	findTrade       func(U, uint32) O
	findWorld       func(uint32) O
	unitCode        func(O) uint16
	loadKind        func() uint8
	loadDefaultInfo func(O) uint8
	loadGuideName   func(O) string
	guideID         func(string) uint8
	loadRecipient   func(U) uint8
	send            func(uint8, [networkInfoBookPacketSize51BAD0]byte)
}

// networkInfoBook51BAD0 preserves GAME.EXE 0051C47A..0051C572. Its three
// short-circuited object searches retain the original inventory, trade, world
// order, while every object and list link remains native pointer-width.
func networkInfoBook51BAD0[O comparable, U any](
	unit O,
	update U,
	hooks networkInfoBookHooks51BAD0[O, U],
) int32 {
	wireCode := hooks.loadWireCode()
	code := hooks.dynamicUnitCode(wireCode)
	if hooks.netDebug() {
		hooks.testHighBit(wireCode)
	}

	item := hooks.findInventory(unit, code)
	var zero O
	if item == zero {
		item = hooks.findTrade(update, code)
	}
	if item == zero {
		item = hooks.findWorld(code)
	}
	if item == zero {
		return networkInfoBookPacketSize51BAD0
	}

	unitCode := hooks.unitCode(item)
	response := [networkInfoBookPacketSize51BAD0]byte{
		0: 0xe2,
		1: byte(unitCode),
		2: byte(unitCode >> 8),
	}
	if hooks.loadKind() == 2 {
		response[3] = hooks.guideID(hooks.loadGuideName(item))
	} else {
		response[3] = hooks.loadDefaultInfo(item)
	}
	hooks.send(hooks.loadRecipient(update), response)
	return networkInfoBookPacketSize51BAD0
}
