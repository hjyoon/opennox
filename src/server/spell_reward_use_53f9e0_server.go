package server

import (
	"unsafe"

	"github.com/opennox/libs/spell"
	"github.com/opennox/libs/strman"

	noxflags "github.com/opennox/opennox/v1/common/flags"
	"github.com/opennox/opennox/v1/common/sound"
)

// UseSpellRewardRuntime53F9E0 supplies the nested spell-grant service and
// delayed deletion operation still owned by the root/legacy runtime. Both
// Object arguments remain native-width throughout the call chain.
type UseSpellRewardRuntime53F9E0 struct {
	SpellGrant    SpellGrantRuntime4FB550
	DelayedDelete func(*Object)
}

type spellRewardUseNativeDeps53F9E0 struct {
	checkSpellClass   func(uint8, uint8) int32
	primaryMessage    func(*Object, string, uint8)
	audit             func(int32, *Object, int32, uint32)
	gameFlagsCheck    func(uint32) int32
	grantSpell        func(*Object, int32, int32, int32, int32) int32
	delayedDeleteItem func(*Object)
}

func useSpellRewardNative53F9E0(
	owner, item *Object,
	deps spellRewardUseNativeDeps53F9E0,
) int32 {
	return useSpellReward53F9E0(spellRewardUseHooks53F9E0[
		*Object,
		*PlayerUpdateData,
		*Player,
		*SpellRewardUseData,
	]{
		loadItemArg: func() *Object {
			return item
		},
		loadOwnerArg: func() *Object {
			return owner
		},
		loadUseData: func(item *Object) *SpellRewardUseData {
			return item.UseData.AsSpellReward()
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
		loadSpell: func(data *SpellRewardUseData) uint8 {
			return data.Spell
		},
		checkSpellClass: deps.checkSpellClass,
		primaryMessage:  deps.primaryMessage,
		loadNetCode: func(owner *Object) uint32 {
			return owner.NetCode
		},
		audit:          deps.audit,
		gameFlagsCheck: deps.gameFlagsCheck,
		loadSpellLevel: func(player *Player, spellID int32) uint32 {
			return player.SpellLvl[spellID]
		},
		grantSpell:        deps.grantSpell,
		delayedDeleteItem: deps.delayedDeleteItem,
	})
}

func spellRewardUseServerDeps53F9E0(
	s *Server,
	runtime UseSpellRewardRuntime53F9E0,
) spellRewardUseNativeDeps53F9E0 {
	return spellRewardUseNativeDeps53F9E0{
		checkSpellClass: func(class, spellID uint8) int32 {
			return rewardSpellClassCheck4F09F0(
				s.Spells.Flags(spell.ID(spellID)),
				uint32(class),
			)
		},
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
		grantSpell: func(owner *Object, spellID, notify, quest, override int32) int32 {
			return s.SpellGrantToPlayer4FB550(
				owner,
				spellID,
				notify,
				quest,
				override,
				runtime.SpellGrant,
			)
		},
		delayedDeleteItem: runtime.DelayedDelete,
	}
}

// UseSpellReward53F9E0 binds GAME.EXE 0053F9E0 to native-width owner, item,
// UseData, PlayerUpdateData, and Player pointers.
func (s *Server) UseSpellReward53F9E0(
	owner, item *Object,
	runtime UseSpellRewardRuntime53F9E0,
) int32 {
	return useSpellRewardNative53F9E0(
		owner,
		item,
		spellRewardUseServerDeps53F9E0(s, runtime),
	)
}

var (
	_ = [1]struct{}{}[unsafe.Sizeof(uintptr(0))-unsafe.Sizeof(Object{}.UseData)]
	_ = [1]struct{}{}[1-unsafe.Sizeof(SpellRewardUseData{})]
	_ = [1]struct{}{}[0-unsafe.Offsetof(SpellRewardUseData{}.Spell)]
	_ = [1]struct{}{}[4-unsafe.Sizeof(Player{}.SpellLvl[0])]
	_ = [1]struct{}{}[137-len(Player{}.SpellLvl)]
)
