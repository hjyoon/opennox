package server

// NetworkTradeExitPacketSize51BAD0 is the exact MSG_TRADE/0x12 packet width.
const NetworkTradeExitPacketSize51BAD0 = networkTradeExitPacketSize51BAD0

// NetworkTradeExit51BAD0 binds the original shop-exit packet contract to the
// native-width PlayerUpdateData and TradeSession layouts.
func NetworkTradeExit51BAD0(update *PlayerUpdateData, exitSession func(*TradeSession)) int32 {
	return networkTradeExit51BAD0(update, networkTradeExitHooks51BAD0[*PlayerUpdateData, TradeSession]{
		loadSession: func(update *PlayerUpdateData) *TradeSession {
			return update.Trade70
		},
		exitSession: exitSession,
	})
}
