package server

import (
	"unsafe"

	noxflags "github.com/opennox/opennox/v1/common/flags"
)

// BeastScrollAwardAllRuntime4EFD80 supplies the protection services still
// owned by the legacy runtime. Player pointers remain native-width; every
// scalar that crosses this boundary keeps the original 32-bit width.
type BeastScrollAwardAllRuntime4EFD80 struct {
	ResetProtection func(uint32, int32)
	AwardProtection func(uint32, int32, int32)
	ApplyProtection func(uint32, *[41]uint32, int32)
}

type beastScrollAwardAllNativeDeps4EFD80 struct {
	resetProtection func(uint32, int32)
	loadEngineFlags func() uint8
	awardProtection func(uint32, int32, int32)
	applyProtection func(uint32, *[41]uint32, int32)
}

func beastScrollAwardAllNative4EFD80(player *Player, deps beastScrollAwardAllNativeDeps4EFD80) {
	beastScrollAwardAll4EFD80(beastScrollAwardAllHooks4EFD80[*Player]{
		loadPlayerArg: func() *Player {
			return player
		},
		loadProtection: func(player *Player) uint32 {
			return player.Prot4640
		},
		resetProtection: deps.resetProtection,
		loadEngineFlags: deps.loadEngineFlags,
		storeScrollLevel: func(player *Player, index int32, value uint32) {
			player.BeastScrollLvl[index] = value
		},
		awardProtection: deps.awardProtection,
		applyProtection: func(protection uint32, player *Player, count int32) {
			deps.applyProtection(protection, &player.BeastScrollLvl, count)
		},
	})
}

func beastScrollAwardAllServerDeps4EFD80(
	runtime BeastScrollAwardAllRuntime4EFD80,
) beastScrollAwardAllNativeDeps4EFD80 {
	return beastScrollAwardAllNativeDeps4EFD80{
		resetProtection: runtime.ResetProtection,
		loadEngineFlags: func() uint8 {
			// GAME.EXE reads only the low byte at 0x85B7A0.
			return uint8(noxflags.GetEngine())
		},
		awardProtection: runtime.AwardProtection,
		applyProtection: runtime.ApplyProtection,
	}
}

// BeastScrollAwardAll4EFD80 binds GAME.EXE 004EFD80 to native-width Player
// storage. Engine flags remain a live global observation at the exact point
// selected by the restored control flow.
func (s *Server) BeastScrollAwardAll4EFD80(player *Player, runtime BeastScrollAwardAllRuntime4EFD80) {
	beastScrollAwardAllNative4EFD80(player, beastScrollAwardAllServerDeps4EFD80(runtime))
}

var (
	_ = [1]struct{}{}[4-unsafe.Sizeof(Player{}.BeastScrollLvl[0])]
	_ = [1]struct{}{}[41-len(Player{}.BeastScrollLvl)]
	_ = [1]struct{}{}[4-unsafe.Sizeof(Player{}.Prot4640)]
)
