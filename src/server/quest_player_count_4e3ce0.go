package server

type questPlayerCountHooks4E3CE0[O comparable, U, P any] struct {
	firstUnit       func() O
	nextUnit        func(O) O
	loadUpdate      func(O) U
	gameHost        func() bool
	noRendering     func() bool
	loadPlayer      func(U) P
	loadPlayerIndex func(P) uint8
	loadQuestState  func(P) uint32
}

// questPlayerCount4E3CE0 preserves GAME.EXE 004E3CE0. Unlike the admission
// check at 004E4100, only state exactly one counts. The update is cached before
// the live flag gates, but its player link is reloaded for the state read.
func questPlayerCount4E3CE0[O comparable, U, P any](hooks questPlayerCountHooks4E3CE0[O, U, P]) int32 {
	var zero O
	var count int32
	for unit := hooks.firstUnit(); unit != zero; unit = hooks.nextUnit(unit) {
		update := hooks.loadUpdate(unit)
		if hooks.gameHost() && hooks.noRendering() {
			player := hooks.loadPlayer(update)
			if hooks.loadPlayerIndex(player) == 31 {
				continue
			}
		}
		player := hooks.loadPlayer(update)
		if hooks.loadQuestState(player) == 1 {
			count++
		}
	}
	return count
}
