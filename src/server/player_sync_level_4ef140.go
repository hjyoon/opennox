package server

const (
	playerSyncLevelCoopFlag4EF140 = uint32(0x2000)
	playerSyncLevelMaximum4EF140  = int32(10)
)

type playerSyncLevelHooks4EF140[O, U, P, R any] struct {
	loadUnitArg         func() O
	loadUpdateData      func(O) U
	loadPlayer          func(U) P
	gameFlagsCheck      func(uint32) int32
	loadXPTable         func(int32) float64
	loadExperience      func(O) float32
	loadLevelProtection func(P) uint32
	storeLevel          func(P, uint8)
	protectLevel        func(uint32, uint8)
	readValues          func(O, int32) R
}

// playerSyncLevel4EF140 preserves GAME.EXE 004EF140. The unit, UpdateData,
// and Player are cached before the game-mode callback, without nil or class
// gates. Cooperative mode stores level ten. Normal mode asks for XPTable
// indices zero through ten, reloads the unit's binary32 experience after each
// callback, and stops only for an ordered table > experience comparison.
// Unordered comparisons therefore continue. A stop at index zero deliberately
// stores and protects the low byte 0xff. Both paths load the cached Player's
// protection token before the level store and finish by recomputing values with
// reward zero; the recomputation result is the function result.
func playerSyncLevel4EF140[O, U, P, R any](
	hooks playerSyncLevelHooks4EF140[O, U, P, R],
) R {
	unit := hooks.loadUnitArg()
	update := hooks.loadUpdateData(unit)
	player := hooks.loadPlayer(update)
	if hooks.gameFlagsCheck(playerSyncLevelCoopFlag4EF140) != 0 {
		token := hooks.loadLevelProtection(player)
		level := uint8(playerSyncLevelMaximum4EF140)
		hooks.storeLevel(player, level)
		hooks.protectLevel(token, level)
		return hooks.readValues(unit, 0)
	}

	level := playerSyncLevelMaximum4EF140
	for index := int32(0); index <= playerSyncLevelMaximum4EF140; index++ {
		threshold := hooks.loadXPTable(index)
		experience := hooks.loadExperience(unit)
		if threshold > float64(experience) {
			level = index - 1
			break
		}
	}
	token := hooks.loadLevelProtection(player)
	levelByte := uint8(level)
	hooks.storeLevel(player, levelByte)
	hooks.protectLevel(token, levelByte)
	return hooks.readValues(unit, 0)
}
