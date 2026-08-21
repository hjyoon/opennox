package server

const (
	spellAwardAllAdminMask4EFC80  = uint8(0x10)
	spellAwardAllQuestFlag4EFC80  = uint32(0x1000)
	spellAwardAllLevelCount4EFC80 = int32(137)
	spellAwardAllFirstIndex4EFC80 = int32(1)
	spellAwardAllAdminLevel4EFC80 = int32(3)
	spellAwardAllWizard4EFC80     = uint8(1)
	spellAwardAllConjurer4EFC80   = uint8(2)
	spellAwardAllFireball4EFC80   = int32(27)
	spellAwardAllCharm4EFC80      = int32(9)
	spellAwardAllLesserHeal4EFC80 = int32(41)
	spellAwardAllDefaultArg4EFC80 = int32(1)
)

type spellAwardAllHooks4EFC80[P comparable, O any] struct {
	loadPlayerArg   func() P
	loadProtection  func(P) uint32
	resetProtection func(uint32, int32)
	loadEngineFlags func() uint8
	storeSpellLevel func(P, int32, uint32)
	awardProtection func(uint32, int32, int32)
	gameFlagsCheck  func(uint32) int32
	loadPlayerClass func(P) uint8
	loadPlayerUnit  func(P) O
	grantSpell      func(O, int32, int32, int32, int32)
	applyProtection func(uint32, P, int32)
}

// spellAwardAll4EFC80 preserves GAME.EXE 004EFC80. The Player argument and
// initial protection token are loaded before protection reset. Engine flags
// are read only after reset, and the Admin decision remains fixed for the
// selected loop. Every index 1..136 is stored before a live protection-token
// reload and award callback.
//
// Only the disabled path queries Quest. It reads Player class after that
// callback and reloads PlayerUnit for each default-spell grant, including the
// second Conjurer grant. Both paths finish with one more live protection-token
// load followed by application to all 137 levels. There are no nil guards.
func spellAwardAll4EFC80[P comparable, O any](hooks spellAwardAllHooks4EFC80[P, O]) {
	player := hooks.loadPlayerArg()
	protection := hooks.loadProtection(player)
	hooks.resetProtection(protection, 0)

	flags := hooks.loadEngineFlags()
	level := int32(0)
	if flags&spellAwardAllAdminMask4EFC80 != 0 {
		level = spellAwardAllAdminLevel4EFC80
	}
	for index := spellAwardAllFirstIndex4EFC80; index < spellAwardAllLevelCount4EFC80; index++ {
		hooks.storeSpellLevel(player, index, uint32(level))
		protection = hooks.loadProtection(player)
		hooks.awardProtection(protection, index, level)
	}

	if flags&spellAwardAllAdminMask4EFC80 == 0 && hooks.gameFlagsCheck(spellAwardAllQuestFlag4EFC80) != 0 {
		switch hooks.loadPlayerClass(player) {
		case spellAwardAllWizard4EFC80:
			unit := hooks.loadPlayerUnit(player)
			hooks.grantSpell(
				unit,
				spellAwardAllFireball4EFC80,
				spellAwardAllDefaultArg4EFC80,
				spellAwardAllDefaultArg4EFC80,
				spellAwardAllDefaultArg4EFC80,
			)
		case spellAwardAllConjurer4EFC80:
			unit := hooks.loadPlayerUnit(player)
			hooks.grantSpell(
				unit,
				spellAwardAllCharm4EFC80,
				spellAwardAllDefaultArg4EFC80,
				spellAwardAllDefaultArg4EFC80,
				spellAwardAllDefaultArg4EFC80,
			)
			unit = hooks.loadPlayerUnit(player)
			hooks.grantSpell(
				unit,
				spellAwardAllLesserHeal4EFC80,
				spellAwardAllDefaultArg4EFC80,
				spellAwardAllDefaultArg4EFC80,
				spellAwardAllDefaultArg4EFC80,
			)
		}
	}

	protection = hooks.loadProtection(player)
	hooks.applyProtection(protection, player, spellAwardAllLevelCount4EFC80)
}
