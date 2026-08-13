package server

const questDedicatedHostIndex4E8F60 = uint8(31)

type questMaybeWarpHooks4E8F60[O comparable, U, P any] struct {
	currentQuestStage  func() uint32
	nextStageThreshold func(uint32) uint32
	firstUnit          func() O
	nextUnit           func(O) O
	loadUpdateData     func(O) U
	gameHost           func() int32
	noRendering        func() int32
	loadPlayer         func(U) P
	loadPlayerIndex    func(P) uint8
	loadQuestState     func(P) uint32
	loadQuestWarpGate  func(U) O
	loadQuestStage     func(P) uint32
}

// questMaybeWarp4E8F60 preserves GAME.EXE 004E8F60. The Quest stage
// threshold is computed before the player list is read. In a dedicated host,
// player index 31 is skipped; every other player with a nonzero Quest state
// must be standing in a warp gate, and at least one such player must have
// reached the unsigned stage threshold.
func questMaybeWarp4E8F60[O comparable, U, P any](
	hooks questMaybeWarpHooks4E8F60[O, U, P],
) int32 {
	stage := hooks.currentQuestStage()
	threshold := hooks.nextStageThreshold(stage)

	var zero O
	unit := hooks.firstUnit()
	if unit == zero {
		return 0
	}

	var count uint32
	var allowed int32
	for unit != zero {
		update := hooks.loadUpdateData(unit)

		skipDedicatedHost := false
		if hooks.gameHost() != 0 && hooks.noRendering() != 0 {
			player := hooks.loadPlayer(update)
			skipDedicatedHost = hooks.loadPlayerIndex(player) == questDedicatedHostIndex4E8F60
		}
		if !skipDedicatedHost {
			player := hooks.loadPlayer(update)
			if hooks.loadQuestState(player) != 0 {
				gate := hooks.loadQuestWarpGate(update)
				count++
				if gate == zero {
					return 0
				}
				if hooks.loadQuestStage(player) >= threshold {
					allowed = 1
				}
			}
		}
		unit = hooks.nextUnit(unit)
	}
	if count == 0 {
		return 0
	}
	return allowed
}
