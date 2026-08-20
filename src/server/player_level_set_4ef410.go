package server

const (
	playerLevelSetMax4EF410       = int8(10)
	playerLevelSetXPTable4EF410   = "XPTable"
	playerLevelSetReadArg4EF410   = int32(0)
	playerLevelSetCoopFlag4EF410  = uint32(0x800)
	playerLevelSetWarrior4EF410   = uint8(0)
	playerLevelSetBookKind4EF410  = int32(3)
	playerLevelSetFirstAbil4EF410 = int32(1)
	playerLevelSetAbilEnd4EF410   = int32(6)
	playerLevelSetPauseArg4EF410  = int32(0)
)

type playerLevelSetHooks4EF410[O, U, P any] struct {
	loadLevelArg        func() uint8
	loadUnitArg         func() O
	loadUpdateData      func(O) U
	loadPlayer          func(U) P
	loadXPTable         func(string, int32) float64
	storeExperience     func(O, float32)
	loadExperienceToken func(P) uint32
	protectExperience   func(uint32, float32)
	reportExperience    func(O)
	loadLevelToken      func(P) uint32
	storeLevel          func(P, uint8)
	protectLevel        func(uint32, uint8)
	readValues          func(O, int32)
	gameFlag            func(uint32) int32
	loadPlayerClass     func(P) uint8
	loadAbilityLevel    func(P, int32) uint32
	bookAbility         func(int32, int32, int32)
	pauseFX             func(O, int32)
}

// GAME.EXE receives each balance value in x87 precision and performs an
// explicit binary32 spill before either storing Experience or calling its
// protection helper.
//
//go:noinline
func playerLevelSetSpill32_4EF410(value float64) float32 { return float32(value) }

// playerLevelSet4EF410 preserves GAME.EXE 004EF410. The argument's low byte
// is interpreted as signed only for the greater-than-ten clamp and XPTable
// index. Consequently 0x0b..0x7f clamp to ten while 0x80..0xff remain negative
// table indices and retain their original byte when stored as Player.Level.
// Unit.UpdateData and Player are cached before that clamp and there are no nil
// or class guards.
//
// XPTable is called twice with the cached signed index. The first independently
// rounded result is stored as Experience. The second callback precedes the
// cached Player's experience-token load and its separately rounded result is
// protected. Reporting then precedes the level-token load, Level store,
// protection call, and player-value recomputation.
//
// The two cooperative flag checks are independent. Only a nonzero first result
// reads the live class byte; class zero scans live ability levels one through
// five and refreshes each nonzero book entry. A nonzero second result invokes
// PauseFX even when the first result was zero or skipped the ability loop.
func playerLevelSet4EF410[O, U, P any](hooks playerLevelSetHooks4EF410[O, U, P]) {
	level := hooks.loadLevelArg()
	unit := hooks.loadUnitArg()
	update := hooks.loadUpdateData(unit)
	player := hooks.loadPlayer(update)

	if int8(level) > playerLevelSetMax4EF410 {
		level = uint8(playerLevelSetMax4EF410)
	}
	index := int32(int8(level))

	experience := hooks.loadXPTable(playerLevelSetXPTable4EF410, index)
	hooks.storeExperience(unit, playerLevelSetSpill32_4EF410(experience))

	protectedExperience := hooks.loadXPTable(playerLevelSetXPTable4EF410, index)
	experienceToken := hooks.loadExperienceToken(player)
	hooks.protectExperience(
		experienceToken,
		playerLevelSetSpill32_4EF410(protectedExperience),
	)
	hooks.reportExperience(unit)

	levelToken := hooks.loadLevelToken(player)
	hooks.storeLevel(player, level)
	hooks.protectLevel(levelToken, level)
	hooks.readValues(unit, playerLevelSetReadArg4EF410)

	if hooks.gameFlag(playerLevelSetCoopFlag4EF410) != 0 {
		if hooks.loadPlayerClass(player) == playerLevelSetWarrior4EF410 {
			for ability := playerLevelSetFirstAbil4EF410; ability < playerLevelSetAbilEnd4EF410; ability++ {
				if hooks.loadAbilityLevel(player, ability) != 0 {
					hooks.bookAbility(
						playerLevelSetBookKind4EF410,
						ability,
						ability-1,
					)
				}
			}
		}
	}
	if hooks.gameFlag(playerLevelSetCoopFlag4EF410) != 0 {
		hooks.pauseFX(unit, playerLevelSetPauseArg4EF410)
	}
}
