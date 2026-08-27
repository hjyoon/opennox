package server

import "github.com/opennox/opennox/v1/common/sound"

// PickupTrapRuntime4F3510 supplies the root-owned object-list operations used
// by TrapPickup's nested DefaultPickup call.
type PickupTrapRuntime4F3510 struct {
	DefaultPickup PickupDefaultRuntime4F31E0
}

type pickupTrapNativeDeps4F3510 struct {
	hasOwner      func(*Object, *Object) int32
	defaultPickup func(*Object, *Object, int32, int32) int32
	audio         func(uint32, *Object, int32, uint32)
}

func pickupTrapNative4F3510(
	owner, item *Object,
	arg3, arg4 int32,
	deps pickupTrapNativeDeps4F3510,
) int32 {
	return pickupTrap4F3510(
		owner,
		item,
		pickupTrapHooks4F3510[*Object]{
			hasOwner: deps.hasOwner,
			loadArg4: func() int32 {
				return arg4
			},
			loadArg3: func() int32 {
				return arg3
			},
			defaultPickup: deps.defaultPickup,
			loadOwnerClassLow: func(owner *Object) uint8 {
				return uint8(owner.ObjClass)
			},
			loadOwnerNetCode: func(owner *Object) uint32 {
				return owner.NetCode
			},
			audio: deps.audio,
		},
	)
}

func pickupTrapServerDeps4F3510(
	s *Server,
	runtime PickupTrapRuntime4F3510,
) pickupTrapNativeDeps4F3510 {
	return pickupTrapNativeDeps4F3510{
		hasOwner: func(item, owner *Object) int32 {
			if item.HasOwner(owner) {
				return 1
			}
			return 0
		},
		defaultPickup: func(owner, item *Object, arg3, arg4 int32) int32 {
			return s.PickupDefault4F31E0(owner, item, arg3, arg4, runtime.DefaultPickup)
		},
		audio: func(id uint32, owner *Object, kind int32, code uint32) {
			s.Audio.EventObj(sound.ID(id), owner, int(kind), code)
		},
	}
}

// PickupTrap4F3510 binds GAME.EXE's registered four-argument TrapPickup
// callback to native-width Object owner links, class, NetCode, and audio paths.
func (s *Server) PickupTrap4F3510(
	owner, item *Object,
	arg3, arg4 int32,
	runtime PickupTrapRuntime4F3510,
) int32 {
	return pickupTrapNative4F3510(
		owner,
		item,
		arg3,
		arg4,
		pickupTrapServerDeps4F3510(s, runtime),
	)
}
