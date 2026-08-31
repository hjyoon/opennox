package server

const networkTryCollidePacketSize51BAD0 = 3

// networkTryCollideHooks51BAD0 exposes every ordered read and call in the
// MSG_TRY_COLLIDE branch at GAME.EXE 0051BFE1..0051C05F. Object and callback
// identities remain in their native pointer domains; only the packet fields
// use fixed-width integers.
type networkTryCollideHooks51BAD0[O comparable, U, P any, F comparable] struct {
	loadWireCode      func() uint16
	dynamicUnitCode   func(uint16) uint32
	netDebug          func() bool
	testHighBit       func(uint16)
	loadPlayer        func(U) P
	loadPlayerStatus  func(P) uint32
	loadTradeActive   func(U) bool
	loadDialogActive  func(U) bool
	objectFromNetCode func(uint32) O
	loadCollide       func(O) F
	callCollide       func(F, O, O)
}

// networkTryCollide51BAD0 preserves the original three-byte packet branch,
// including the status, trade, and dialog short-circuit order and the cached
// collide callback used for the final target/unit invocation.
func networkTryCollide51BAD0[O comparable, U, P any, F comparable](
	unit O,
	update U,
	hooks networkTryCollideHooks51BAD0[O, U, P, F],
) int32 {
	wireCode := hooks.loadWireCode()
	code := hooks.dynamicUnitCode(wireCode)
	if hooks.netDebug() {
		hooks.testHighBit(wireCode)
	}
	player := hooks.loadPlayer(update)
	if hooks.loadPlayerStatus(player)&0x3 != 0 ||
		hooks.loadTradeActive(update) ||
		hooks.loadDialogActive(update) {
		return networkTryCollidePacketSize51BAD0
	}
	target := hooks.objectFromNetCode(code)
	var zeroObject O
	if target == zeroObject {
		return networkTryCollidePacketSize51BAD0
	}
	callback := hooks.loadCollide(target)
	var zeroCallback F
	if callback == zeroCallback {
		return networkTryCollidePacketSize51BAD0
	}
	hooks.callCollide(callback, target, unit)
	return networkTryCollidePacketSize51BAD0
}
