package server

import (
	noxflags "github.com/opennox/opennox/v1/common/flags"
	"github.com/opennox/opennox/v1/common/sound"
)

// PickupAbilityBookRuntime4F3CE0 contains the root-owned object-list operations
// used by AbilityBookPickup's nested DefaultPickup call.
type PickupAbilityBookRuntime4F3CE0 struct {
	DefaultPickup PickupDefaultRuntime4F31E0
}

type pickupAbilityBookNativeDeps4F3CE0 struct {
	gameFlagsCheck func(uint32) int32
	useByNetCode   func(*Object, *Object)
	defaultPickup  func(*Object, *Object, int32, int32) int32
	audio          func(uint32, *Object, int32, uint32)
}

func pickupAbilityBookNative4F3CE0(
	owner, item *Object,
	arg3, arg4 int32,
	deps pickupAbilityBookNativeDeps4F3CE0,
) int32 {
	return pickupAbilityBook4F3CE0(owner, item, pickupAbilityBookHooks4F3CE0[*Object]{
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
		audio:         deps.audio,
	})
}

func pickupAbilityBookServerDeps4F3CE0(
	s *Server,
	runtime PickupAbilityBookRuntime4F3CE0,
) pickupAbilityBookNativeDeps4F3CE0 {
	return pickupAbilityBookNativeDeps4F3CE0{
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

// PickupAbilityBook4F3CE0 binds GAME.EXE's registered four-argument
// AbilityBookPickup callback to native-width Object pointers.
func (s *Server) PickupAbilityBook4F3CE0(
	owner, item *Object,
	arg3, arg4 int32,
	runtime PickupAbilityBookRuntime4F3CE0,
) int32 {
	return pickupAbilityBookNative4F3CE0(
		owner,
		item,
		arg3,
		arg4,
		pickupAbilityBookServerDeps4F3CE0(s, runtime),
	)
}
