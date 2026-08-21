package server

import (
	"unsafe"

	noxflags "github.com/opennox/opennox/v1/common/flags"
)

// SpellAwardAllRuntime4EFC80 supplies the protection and spell services still
// owned by the legacy runtime. Player and Object pointers remain native-width;
// every scalar that crosses this boundary keeps the original 32-bit width.
type SpellAwardAllRuntime4EFC80 struct {
	ResetProtection func(uint32, int32)
	AwardProtection func(uint32, int32, int32)
	GrantSpell      func(*Object, int32, int32, int32, int32)
	ApplyProtection func(uint32, *[137]uint32, int32)
}

type spellAwardAllNativeDeps4EFC80 struct {
	resetProtection func(uint32, int32)
	loadEngineFlags func() uint8
	awardProtection func(uint32, int32, int32)
	gameFlagsCheck  func(uint32) int32
	grantSpell      func(*Object, int32, int32, int32, int32)
	applyProtection func(uint32, *[137]uint32, int32)
}

func spellAwardAllNative4EFC80(player *Player, deps spellAwardAllNativeDeps4EFC80) {
	spellAwardAll4EFC80(spellAwardAllHooks4EFC80[*Player, *Object]{
		loadPlayerArg: func() *Player {
			return player
		},
		loadProtection: func(player *Player) uint32 {
			return player.Prot4636
		},
		resetProtection: deps.resetProtection,
		loadEngineFlags: deps.loadEngineFlags,
		storeSpellLevel: func(player *Player, index int32, value uint32) {
			player.SpellLvl[index] = value
		},
		awardProtection: deps.awardProtection,
		gameFlagsCheck:  deps.gameFlagsCheck,
		loadPlayerClass: func(player *Player) uint8 {
			return player.info[66]
		},
		loadPlayerUnit: func(player *Player) *Object {
			return player.PlayerUnit
		},
		grantSpell: deps.grantSpell,
		applyProtection: func(protection uint32, player *Player, count int32) {
			deps.applyProtection(protection, &player.SpellLvl, count)
		},
	})
}

func spellAwardAllServerDeps4EFC80(
	runtime SpellAwardAllRuntime4EFC80,
) spellAwardAllNativeDeps4EFC80 {
	return spellAwardAllNativeDeps4EFC80{
		resetProtection: runtime.ResetProtection,
		loadEngineFlags: func() uint8 {
			// GAME.EXE reads only the low byte at 0x85B7A0.
			return uint8(noxflags.GetEngine())
		},
		awardProtection: runtime.AwardProtection,
		gameFlagsCheck: func(mask uint32) int32 {
			if noxflags.HasGame(noxflags.GameFlag(mask)) {
				return 1
			}
			return 0
		},
		grantSpell:      runtime.GrantSpell,
		applyProtection: runtime.ApplyProtection,
	}
}

// SpellAwardAll4EFC80 binds GAME.EXE 004EFC80 to native-width Player and
// Object storage. Engine and game flags remain live global observations at the
// exact points selected by the restored control flow.
func (s *Server) SpellAwardAll4EFC80(player *Player, runtime SpellAwardAllRuntime4EFC80) {
	spellAwardAllNative4EFC80(player, spellAwardAllServerDeps4EFC80(runtime))
}

var (
	_ = [1]struct{}{}[4-unsafe.Sizeof(Player{}.SpellLvl[0])]
	_ = [1]struct{}{}[137-len(Player{}.SpellLvl)]
	_ = [1]struct{}{}[4-unsafe.Sizeof(Player{}.Prot4636)]
)
