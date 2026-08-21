package server

import (
	"unsafe"

	noxflags "github.com/opennox/opennox/v1/common/flags"
)

// WarriorAbilityAwardAllRuntime4EFE10 supplies the protection-award service
// still owned by the legacy runtime. Player pointers remain native-width;
// every scalar crossing the boundary keeps the original 32-bit width.
type WarriorAbilityAwardAllRuntime4EFE10 struct {
	AwardProtection func(uint32, int32, int32)
}

type warriorAbilityAwardAllNativeDeps4EFE10 struct {
	loadEngineFlags func() uint8
	awardProtection func(uint32, int32, int32)
}

func warriorAbilityAwardAllNative4EFE10(player *Player, deps warriorAbilityAwardAllNativeDeps4EFE10) {
	warriorAbilityAwardAll4EFE10(warriorAbilityAwardAllHooks4EFE10[*Player]{
		loadPlayerArg: func() *Player {
			return player
		},
		loadPlayerClass: func(player *Player) uint8 {
			return player.info[66]
		},
		loadEngineFlags: deps.loadEngineFlags,
		storeAbilityLevel: func(player *Player, index int32, value uint32) {
			player.SpellLvl[index] = value
		},
		loadProtection: func(player *Player) uint32 {
			return player.Prot4636
		},
		awardProtection: deps.awardProtection,
	})
}

func warriorAbilityAwardAllServerDeps4EFE10(
	runtime WarriorAbilityAwardAllRuntime4EFE10,
) warriorAbilityAwardAllNativeDeps4EFE10 {
	return warriorAbilityAwardAllNativeDeps4EFE10{
		loadEngineFlags: func() uint8 {
			// GAME.EXE reads only the low byte at 0x85B7A0.
			return uint8(noxflags.GetEngine())
		},
		awardProtection: runtime.AwardProtection,
	}
}

// WarriorAbilityAwardAll4EFE10 binds GAME.EXE 004EFE10 to native-width
// Player, PlayerInfo, ability-level, and protection-token storage. Engine
// flags remain a live global observation after the class gate.
func (s *Server) WarriorAbilityAwardAll4EFE10(
	player *Player,
	runtime WarriorAbilityAwardAllRuntime4EFE10,
) {
	warriorAbilityAwardAllNative4EFE10(player, warriorAbilityAwardAllServerDeps4EFE10(runtime))
}

var (
	_ = [1]struct{}{}[4-unsafe.Sizeof(Player{}.SpellLvl[0])]
	_ = [1]struct{}{}[137-len(Player{}.SpellLvl)]
	_ = [1]struct{}{}[4-unsafe.Sizeof(Player{}.Prot4636)]
)
