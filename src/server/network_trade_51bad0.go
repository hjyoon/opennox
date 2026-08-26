package server

const networkTradeExitPacketSize51BAD0 = 2

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
