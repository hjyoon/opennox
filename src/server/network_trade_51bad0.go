package server

const networkTradeExitPacketSize51BAD0 = 2

const networkTradeStartPacketSize51BAD0 = 4

const networkTradeBuyPacketSize51BAD0 = 4

type networkTradeStartHooks51BAD0[O comparable, U, P any] struct {
	gameBlocked            func() bool
	loadPlayer             func(U) P
	loadPlayerStatus       func(P) uint32
	loadWireCode           func() uint16
	dynamicUnitCode        func(uint16) uint32
	objectFromNetCode      func(uint32) O
	loadMonsterSubclassLow func(O) uint8
	startShop              func(O, O)
}

// networkTradeStart51BAD0 preserves the MSG_TRADE/0x15 branch from
// GAME.EXE 0051CBA6..0051CBEE. In particular, the game/player gates run
// before the packet code is decoded, and every rejected request still
// consumes the exact four-byte packet.
func networkTradeStart51BAD0[O comparable, U, P any](
	unit O,
	update U,
	hooks networkTradeStartHooks51BAD0[O, U, P],
) int32 {
	if hooks.gameBlocked() {
		return networkTradeStartPacketSize51BAD0
	}
	player := hooks.loadPlayer(update)
	if hooks.loadPlayerStatus(player)&0x3 != 0 {
		return networkTradeStartPacketSize51BAD0
	}
	wireCode := hooks.loadWireCode()
	code := hooks.dynamicUnitCode(wireCode)
	merchant := hooks.objectFromNetCode(code)
	var zero O
	if merchant == zero || hooks.loadMonsterSubclassLow(merchant)&0x8 == 0 {
		return networkTradeStartPacketSize51BAD0
	}
	hooks.startShop(unit, merchant)
	return networkTradeStartPacketSize51BAD0
}

type networkTradeExitHooks51BAD0[U, T any] struct {
	loadSession func(U) *T
	exitSession func(*T)
}

// networkTradeExit51BAD0 preserves the MSG_TRADE/0x12 branch from
// GAME.EXE 0051CCE1..0051CCF9 and its two-byte shared tail at 0051CE3E
// without indexing PlayerUpdateData as PE32 ints. The original consumes the
// packet even when no session is active.
func networkTradeExit51BAD0[U, T any](update U, hooks networkTradeExitHooks51BAD0[U, T]) int32 {
	session := hooks.loadSession(update)
	if session != nil {
		hooks.exitSession(session)
	}
	return networkTradeExitPacketSize51BAD0
}

type networkTradeBuyHooks51BAD0[U, T any] struct {
	loadSession func(U) *T
	loadNetCode func() uint16
	buy         func(*T, uint16)
}

// networkTradeBuy51BAD0 preserves the MSG_TRADE/0x16 decoder branch from
// GAME.EXE 0051CCFA..0051CD1B. A missing session still consumes the exact
// four-byte request and never decodes or dispatches its item code.
func networkTradeBuy51BAD0[U, T any](update U, hooks networkTradeBuyHooks51BAD0[U, T]) int32 {
	session := hooks.loadSession(update)
	if session != nil {
		hooks.buy(session, hooks.loadNetCode())
	}
	return networkTradeBuyPacketSize51BAD0
}

func NetworkTradeBuy51BAD0(update *PlayerUpdateData, packet *[networkTradeBuyPacketSize51BAD0]byte, buy func(*TradeSession, uint16)) int32 {
	return networkTradeBuy51BAD0(update, networkTradeBuyHooks51BAD0[*PlayerUpdateData, TradeSession]{
		loadSession: func(update *PlayerUpdateData) *TradeSession {
			return update.Trade70
		},
		loadNetCode: func() uint16 {
			return uint16(packet[2]) | uint16(packet[3])<<8
		},
		buy: buy,
	})
}
