package server

import (
	"unsafe"

	"github.com/opennox/libs/strman"

	noxflags "github.com/opennox/opennox/v1/common/flags"
	"github.com/opennox/opennox/v1/common/sound"
)

// UseAbilityRewardRuntime53FAE0 supplies the nested reward service and delayed
// deletion operation still owned by the root/legacy runtime. Both Object
// arguments remain native-width throughout the call chain.
type UseAbilityRewardRuntime53FAE0 struct {
	AbilityReward AbilityRewardRuntime4FB9C0
	DelayedDelete func(*Object)
}

type abilityRewardUseNativeDeps53FAE0 struct {
	primaryMessage func(*Object, string, uint8)
	audit          func(int32, *Object, int32, uint32)
	gameFlagsCheck func(uint32) int32
	rewardAbility  func(*Object, int32, int32) int32
	delayedDelete  func(*Object)
}

func useAbilityRewardNative53FAE0(
	owner, item *Object,
	deps abilityRewardUseNativeDeps53FAE0,
) int32 {
	return useAbilityReward53FAE0(abilityRewardUseHooks53FAE0[
		*Object,
		*PlayerUpdateData,
		*Player,
		*AbilityRewardUseData,
	]{
		loadItemArg: func() *Object {
			return item
		},
		loadOwnerArg: func() *Object {
			return owner
		},
		loadUseData: func(item *Object) *AbilityRewardUseData {
			return item.UseData.AsAbilityReward()
		},
		loadClassLow: func(owner *Object) uint8 {
			return uint8(owner.ObjClass)
		},
		loadUpdateData: func(owner *Object) *PlayerUpdateData {
			return (*PlayerUpdateData)(owner.UpdateData)
		},
		loadPlayer: func(update *PlayerUpdateData) *Player {
			return update.Player
		},
		loadPlayerClass: func(player *Player) uint8 {
			// Direct indexing deliberately faults on nil, matching GAME.EXE.
			return player.info[66]
		},
		primaryMessage: deps.primaryMessage,
		loadNetCode: func(owner *Object) uint32 {
			return owner.NetCode
		},
		audit:          deps.audit,
		gameFlagsCheck: deps.gameFlagsCheck,
		loadAbility: func(data *AbilityRewardUseData) uint8 {
			return data.Ability
		},
		loadAbilityLevel: func(player *Player, ability int32) uint32 {
			return player.SpellLvl[ability]
		},
		rewardAbility: deps.rewardAbility,
		delayedDelete: deps.delayedDelete,
	})
}

func abilityRewardUseServerDeps53FAE0(
	s *Server,
	runtime UseAbilityRewardRuntime53FAE0,
) abilityRewardUseNativeDeps53FAE0 {
	return abilityRewardUseNativeDeps53FAE0{
		primaryMessage: func(owner *Object, message string, value uint8) {
			s.NetPriMsgToPlayer(owner, strman.ID(message), value)
		},
		audit: func(id int32, owner *Object, kind int32, code uint32) {
			s.Audio.EventObj(sound.ID(id), owner, int(kind), code)
		},
		gameFlagsCheck: func(mask uint32) int32 {
			if noxflags.HasGame(noxflags.GameFlag(mask)) {
				return 1
			}
			return 0
		},
		rewardAbility: func(owner *Object, ability, rewardArg int32) int32 {
			return s.AbilityRewardServ4FB9C0(owner, ability, rewardArg, runtime.AbilityReward)
		},
		delayedDelete: runtime.DelayedDelete,
	}
}

// UseAbilityReward53FAE0 binds GAME.EXE 0053FAE0 to native-width owner, item,
// UseData, PlayerUpdateData, and Player pointers.
func (s *Server) UseAbilityReward53FAE0(
	owner, item *Object,
	runtime UseAbilityRewardRuntime53FAE0,
) int32 {
	return useAbilityRewardNative53FAE0(
		owner,
		item,
		abilityRewardUseServerDeps53FAE0(s, runtime),
	)
}

var (
	_ = [1]struct{}{}[unsafe.Sizeof(uintptr(0))-unsafe.Sizeof(Object{}.UseData)]
	_ = [1]struct{}{}[1-unsafe.Sizeof(AbilityRewardUseData{})]
	_ = [1]struct{}{}[0-unsafe.Offsetof(AbilityRewardUseData{}.Ability)]
)
