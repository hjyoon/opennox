package server

const (
	experienceLevelCoopFlag4EF2E0    = uint32(0x800)
	experienceLevelReward4EF2E0      = int32(1)
	experienceLevelPauseMode4EF2E0   = int32(0)
	experienceLevelXPTable4EF2E0     = "XPTable"
	experienceLevelMessageKey4EF2E0  = "LevelUP"
	experienceLevelMessagePath4EF2E0 = `C:\NoxPost\src\Server\GameMech\explevel.c`
	experienceLevelMessageLine4EF2E0 = 253
	experienceLevelSound4EF2E0       = uint32(902)
	experienceLevelSoundKind4EF2E0   = int32(2)
)

type experienceLevelUpdateHooks4EF2E0[O, U, P, M any] struct {
	loadUnitArg     func() O
	loadUpdateData  func(O) U
	loadPlayer      func(U) P
	gameGet         func() int32
	gameSubActive   func() bool
	loadLevel       func(P) uint8
	loadXPTable     func(string, int32) float64
	loadExperience  func(O) float32
	storeLevel      func(P, uint8)
	loadLevelToken  func(P) uint32
	protectLevel    func(uint32, uint8)
	readValues      func(O, int32)
	gameFlag        func(uint32) int32
	pauseFX         func(O, int32)
	loadNetCode     func(O) uint32
	audio           func(uint32, O, int32, uint32)
	loadString      func(string, string, int) M
	sendLineMessage func(O, M)
}

// experienceLevelUpdate4EF2E0 preserves GAME.EXE 004EF2E0. Unit,
// UpdateData, and Player are cached before either game-state callback. Only
// the exact gameGet result one consults gameSubActive, and an active result
// returns before Level or Experience is read. XPTable uses signed int8(Level)
// plus one; its callback precedes the live binary32 Experience read. Only an
// ordered threshold greater than Experience returns, so equality and every
// unordered comparison level up.
//
// The level byte is reloaded after the table callback and incremented with
// uint8 wraparound. Its protection token is loaded after the store, followed
// by protection and value recomputation with reward one. A nonzero 0x800 flag
// calls PauseFX and returns without reading NetCode or strings. The zero branch
// reloads NetCode after that callback, emits sound 902/kind 2, loads LevelUP at
// the original path and line, and finally sends the line message.
func experienceLevelUpdate4EF2E0[O, U, P, M any](
	hooks experienceLevelUpdateHooks4EF2E0[O, U, P, M],
) {
	unit := hooks.loadUnitArg()
	update := hooks.loadUpdateData(unit)
	player := hooks.loadPlayer(update)
	if hooks.gameGet() == 1 && hooks.gameSubActive() {
		return
	}

	index := int32(int8(hooks.loadLevel(player))) + 1
	threshold := hooks.loadXPTable(experienceLevelXPTable4EF2E0, index)
	experience := hooks.loadExperience(unit)
	if threshold > float64(experience) {
		return
	}

	level := hooks.loadLevel(player)
	level++
	hooks.storeLevel(player, level)
	token := hooks.loadLevelToken(player)
	hooks.protectLevel(token, 1)
	hooks.readValues(unit, experienceLevelReward4EF2E0)
	if hooks.gameFlag(experienceLevelCoopFlag4EF2E0) != 0 {
		hooks.pauseFX(unit, experienceLevelPauseMode4EF2E0)
		return
	}

	netCode := hooks.loadNetCode(unit)
	hooks.audio(
		experienceLevelSound4EF2E0,
		unit,
		experienceLevelSoundKind4EF2E0,
		netCode,
	)
	message := hooks.loadString(
		experienceLevelMessageKey4EF2E0,
		experienceLevelMessagePath4EF2E0,
		experienceLevelMessageLine4EF2E0,
	)
	hooks.sendLineMessage(unit, message)
}
