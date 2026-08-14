package server

const questDedicatedHostIndex4E9010 = uint8(31)

type questAllPlayersExitedHooks4E9010[O comparable, U, P any] struct {
	firstUnit       func() O
	nextUnit        func(O) O
	loadUpdateData  func(O) U
	gameHost        func() int32
	noRendering     func() int32
	loadPlayer      func(U) P
	loadPlayerIndex func(P) uint8
	loadQuestState  func(P) uint32
	loadQuestExit   func(U) O
}

// questAllPlayersExited4E9010 preserves GAME.EXE 004E9010. In a dedicated
// host, player index 31 is skipped. At least one other player must have a
// nonzero Quest state, and every such player must be standing in an exit.
func questAllPlayersExited4E9010[O comparable, U, P any](
	hooks questAllPlayersExitedHooks4E9010[O, U, P],
) int32 {
	var zero O
	unit := hooks.firstUnit()
	if unit == zero {
		return 0
	}

	var count uint32
	for unit != zero {
		update := hooks.loadUpdateData(unit)

		skipDedicatedHost := false
		if hooks.gameHost() != 0 && hooks.noRendering() != 0 {
			player := hooks.loadPlayer(update)
			skipDedicatedHost = hooks.loadPlayerIndex(player) == questDedicatedHostIndex4E9010
		}
		if !skipDedicatedHost {
			player := hooks.loadPlayer(update)
			if hooks.loadQuestState(player) != 0 {
				exit := hooks.loadQuestExit(update)
				count++
				if exit == zero {
					return 0
				}
			}
		}
		unit = hooks.nextUnit(unit)
	}
	if count == 0 {
		return 0
	}
	return 1
}
