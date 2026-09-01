package server

import (
	"unsafe"

	"github.com/opennox/libs/strman"

	noxflags "github.com/opennox/opennox/v1/common/flags"
)

// UseFieldGuideRuntime53F930 supplies the nested guide-award service and
// delayed deletion operation. Every object and data pointer remains
// native-width throughout the registered-use call chain.
type UseFieldGuideRuntime53F930 struct {
	BeastGuide    BeastGuideAwardRuntime4FAE80
	DelayedDelete func(*Object)
}

type fieldGuideUseNativeDeps53F930 struct {
	guideByName       func(string) int32
	gameFlagsCheck    func(uint32) int32
	primaryMessage    func(*Object, string, uint8)
	awardGuide        func(*Object, int32, int32) int32
	delayedDeleteItem func(*Object)
}

func useFieldGuideNative53F930(
	owner, item *Object,
	deps fieldGuideUseNativeDeps53F930,
) int32 {
	return useFieldGuide53F930(fieldGuideUseHooks53F930[
		*Object,
		*PlayerUpdateData,
		*Player,
		*FieldGuideUseData,
	]{
		loadOwnerArg: func() *Object {
			return owner
		},
		loadClassLow: func(owner *Object) uint8 {
			return uint8(owner.ObjClass)
		},
		loadItemArg: func() *Object {
			return item
		},
		loadUpdateData: func(owner *Object) *PlayerUpdateData {
			return (*PlayerUpdateData)(owner.UpdateData)
		},
		loadUseData: func(item *Object) *FieldGuideUseData {
			return item.UseData.AsFieldGuide()
		},
		loadCreature: func(data *FieldGuideUseData) string {
			return data.Creature()
		},
		guideByName:    deps.guideByName,
		gameFlagsCheck: deps.gameFlagsCheck,
		loadPlayer: func(update *PlayerUpdateData) *Player {
			return update.Player
		},
		loadPlayerClass: func(player *Player) uint8 {
			return player.info[66]
		},
		loadGuideLevel: func(player *Player, guide int32) uint32 {
			return player.BeastScrollLvl[guide]
		},
		primaryMessage:    deps.primaryMessage,
		awardGuide:        deps.awardGuide,
		delayedDeleteItem: deps.delayedDeleteItem,
	})
}

func fieldGuideUseServerDeps53F930(
	s *Server,
	runtime UseFieldGuideRuntime53F930,
) fieldGuideUseNativeDeps53F930 {
	return fieldGuideUseNativeDeps53F930{
		guideByName: func(name string) int32 {
			return int32(RewardFieldGuideID4F0D20(name))
		},
		gameFlagsCheck: func(mask uint32) int32 {
			if noxflags.HasGame(noxflags.GameFlag(mask)) {
				return 1
			}
			return 0
		},
		primaryMessage: func(owner *Object, message string, value uint8) {
			s.NetPriMsgToPlayer(owner, strman.ID(message), value)
		},
		awardGuide: func(owner *Object, guide, notify int32) int32 {
			return s.AwardBeastGuide4FAE80(owner, guide, notify, runtime.BeastGuide)
		},
		delayedDeleteItem: runtime.DelayedDelete,
	}
}

// UseFieldGuide53F930 binds GAME.EXE 0053F930 to native-width owner, item,
// UseData, PlayerUpdateData, and Player pointers.
func (s *Server) UseFieldGuide53F930(
	owner, item *Object,
	runtime UseFieldGuideRuntime53F930,
) int32 {
	return useFieldGuideNative53F930(
		owner,
		item,
		fieldGuideUseServerDeps53F930(s, runtime),
	)
}

var (
	_ = [1]struct{}{}[unsafe.Sizeof(uintptr(0))-unsafe.Sizeof(Object{}.UseData)]
	_ = [1]struct{}{}[64-unsafe.Sizeof(FieldGuideUseData{})]
	_ = [1]struct{}{}[0-unsafe.Offsetof(FieldGuideUseData{}.CreatureBuf)]
)
