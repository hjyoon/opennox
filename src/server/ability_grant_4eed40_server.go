package server

import (
	"unsafe"

	noxflags "github.com/opennox/opennox/v1/common/flags"
)

// AbilityGivePlayerAllRuntime4EED40 supplies the legacy-owned ability table and
// services used by GAME.EXE 004EED40. Object, PlayerUpdateData, and Player
// access remains native-width.
type AbilityGivePlayerAllRuntime4EED40 struct {
	LoadAbilityID func(int32) uint32
	IsQuest       func() int32
	QuestMode     func() int32
	RewardAbility func(*Object, int32, int32)
}

type abilityGivePlayerAllNativeDeps4EED40 struct {
	loadAbilityID  func(int32) uint32
	gameFlagsCheck func(uint32) int32
	isQuest        func() int32
	questMode      func() int32
	rewardAbility  func(*Object, int32, int32)
}

func abilityGivePlayerAllNative4EED40(
	unit *Object,
	count int8,
	rewardArg int32,
	deps abilityGivePlayerAllNativeDeps4EED40,
) {
	abilityGivePlayerAll4EED40(abilityGivePlayerAllHooks4EED40[
		*Object,
		*PlayerUpdateData,
		*Player,
	]{
		loadUnitArg: func() *Object {
			return unit
		},
		loadUpdateData: func(obj *Object) *PlayerUpdateData {
			// Do not use UpdateDataPlayer: 004EED40 has no class check.
			return (*PlayerUpdateData)(obj.UpdateData)
		},
		loadCountLow: func() int8 {
			return count
		},
		loadPlayer: func(update *PlayerUpdateData) *Player {
			return update.Player
		},
		loadAbilityID:  deps.loadAbilityID,
		gameFlagsCheck: deps.gameFlagsCheck,
		isQuest:        deps.isQuest,
		questMode:      deps.questMode,
		loadRewardArg: func() int32 {
			return rewardArg
		},
		rewardAbility: deps.rewardAbility,
		storeAbilityLevel: func(player *Player, index int32, value uint32) {
			player.SpellLvl[index] = value
		},
	})
}

func abilityGivePlayerAllServerDeps4EED40(
	runtime AbilityGivePlayerAllRuntime4EED40,
) abilityGivePlayerAllNativeDeps4EED40 {
	return abilityGivePlayerAllNativeDeps4EED40{
		loadAbilityID: runtime.LoadAbilityID,
		gameFlagsCheck: func(mask uint32) int32 {
			if noxflags.HasGame(noxflags.GameFlag(mask)) {
				return 1
			}
			return 0
		},
		isQuest:       runtime.IsQuest,
		questMode:     runtime.QuestMode,
		rewardAbility: runtime.RewardAbility,
	}
}

// AbilityGivePlayerAll4EED40 binds GAME.EXE 004EED40 to native-width Object,
// PlayerUpdateData, and Player pointers. count is intentionally int8: the
// original routine sign-extends only its low byte before entering the loop.
func (s *Server) AbilityGivePlayerAll4EED40(
	unit *Object,
	count int8,
	rewardArg int32,
	runtime AbilityGivePlayerAllRuntime4EED40,
) {
	abilityGivePlayerAllNative4EED40(unit, count, rewardArg, abilityGivePlayerAllServerDeps4EED40(runtime))
}

var (
	_ = [1]struct{}{}[4-unsafe.Sizeof(Player{}.SpellLvl[0])]
	_ = [1]struct{}{}[137-len(Player{}.SpellLvl)]
)
