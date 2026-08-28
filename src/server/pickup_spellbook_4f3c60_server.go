package server

import (
	noxflags "github.com/opennox/opennox/v1/common/flags"
	"github.com/opennox/opennox/v1/common/sound"
)

// PickupSpellBookRuntime4F3C60 contains the root-owned object-list operations
// used by SpellBookPickup's nested DefaultPickup call.
type PickupSpellBookRuntime4F3C60 struct {
	DefaultPickup PickupDefaultRuntime4F31E0
}

type pickupSpellBookNativeDeps4F3C60 struct {
	gameFlagsCheck func(uint32) int32
	useByNetCode   func(*Object, *Object)
	defaultPickup  func(*Object, *Object, int32, int32) int32
	audio          func(uint32, *Object, int32, uint32)
}

func pickupSpellBookNative4F3C60(
	owner, item *Object,
	arg3, arg4 int32,
	deps pickupSpellBookNativeDeps4F3C60,
) int32 {
	return pickupSpellBook4F3C60(owner, item, pickupSpellBookHooks4F3C60[*Object]{
		gameFlagsCheck: deps.gameFlagsCheck,
		useByNetCode:   deps.useByNetCode,
		loadItemFlagsLow: func(item *Object) uint8 {
			return uint8(item.ObjFlags)
		},
		loadArg4: func() int32 {
			return arg4
		},
		loadArg3: func() int32 {
			return arg3
		},
		defaultPickup: deps.defaultPickup,
		loadItemSubClassLow: func(item *Object) uint8 {
			return uint8(item.ObjSubClass)
		},
		audio: deps.audio,
	})
}

func pickupSpellBookServerDeps4F3C60(
	s *Server,
	runtime PickupSpellBookRuntime4F3C60,
) pickupSpellBookNativeDeps4F3C60 {
	return pickupSpellBookNativeDeps4F3C60{
		gameFlagsCheck: func(mask uint32) int32 {
			if noxflags.HasGame(noxflags.GameFlag(mask)) {
				return 1
			}
			return 0
		},
		useByNetCode: func(owner, item *Object) {
			_ = s.pickupUseByNetCode4F34D0(owner, item)
		},
		defaultPickup: func(owner, item *Object, arg3, arg4 int32) int32 {
			return s.PickupDefault4F31E0(owner, item, arg3, arg4, runtime.DefaultPickup)
		},
		audio: func(id uint32, owner *Object, kind int32, code uint32) {
			s.Audio.EventObj(sound.ID(id), owner, int(kind), code)
		},
	}
}

// PickupSpellBook4F3C60 binds GAME.EXE's registered four-argument
// SpellBookPickup callback to native-width Object pointers.
func (s *Server) PickupSpellBook4F3C60(
	owner, item *Object,
	arg3, arg4 int32,
	runtime PickupSpellBookRuntime4F3C60,
) int32 {
	return pickupSpellBookNative4F3C60(
		owner,
		item,
		arg3,
		arg4,
		pickupSpellBookServerDeps4F3C60(s, runtime),
	)
}
