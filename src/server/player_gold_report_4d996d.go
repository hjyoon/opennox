package server

type playerGoldReportHooks4D996D[O, U, P any] struct {
	loadPlayer       func(U) P
	loadReportedGold func(P) uint32
	loadGold         func(P) uint32
	loadPlayerIndex  func(P) int32
	reportGold       func(int32, O)
	storeReported    func(P, uint32)
}

// playerGoldReport4D996D preserves the gold-delta block inside GAME.EXE's
// player-report routine. It compares the first live Player's reported cache
// with GoldVal, reports only a change, then reloads Player through the
// entry-cached update record and copies that second live GoldVal into the
// second Player's cache. The original has no nil guards in this block.
func playerGoldReport4D996D[O, U, P any](
	unit O,
	update U,
	hooks playerGoldReportHooks4D996D[O, U, P],
) {
	player := hooks.loadPlayer(update)
	reported := hooks.loadReportedGold(player)
	gold := hooks.loadGold(player)
	if reported == gold {
		return
	}
	index := hooks.loadPlayerIndex(player)
	hooks.reportGold(index, unit)
	player = hooks.loadPlayer(update)
	gold = hooks.loadGold(player)
	hooks.storeReported(player, gold)
}
