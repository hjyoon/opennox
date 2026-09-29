package server

import (
	"github.com/opennox/libs/strman"

	"github.com/opennox/opennox/v1/common/sound"
)

// PickupOblivionRuntime53A9C0 supplies the root- and legacy-owned calls used
// after the server-owned class, subclass, player-state, message, and audio
// operations have been bound to native-width objects.
type PickupOblivionRuntime53A9C0 struct {
	WeaponPickup func(*Object, *Object, int32, int32) int32
	PauseFX      func(*Object, int32)
	TryEquip     func(*Object, *Object)
}

type pickupOblivionNativeDeps53A9C0 struct {
	weaponPickup    func(*Object, *Object, int32, int32) int32
	playerState     func(*Object) int32
	priorityMessage func(*Object, string, uint8)
	audio           func(uint32, *Object, int32, uint32)
	pauseFX         func(*Object, int32)
	tryEquip        func(*Object, *Object)
}

func pickupOblivionNative53A9C0(
	owner, item *Object,
	arg3, arg4 int32,
	deps pickupOblivionNativeDeps53A9C0,
) int32 {
	return pickupOblivion53A9C0(owner, item, arg3, arg4, pickupOblivionHooks53A9C0[*Object]{
		weaponPickup: deps.weaponPickup,
		loadOwnerClass: func(owner *Object) uint32 {
			return uint32(owner.ObjClass)
		},
		playerState: deps.playerState,
		loadItemSubclass: func(item *Object) uint32 {
			return uint32(item.ObjSubClass)
		},
		priorityMessage: deps.priorityMessage,
		audio:           deps.audio,
		pauseFX:         deps.pauseFX,
		tryEquip:        deps.tryEquip,
	})
}

func pickupOblivionServerDeps53A9C0(
	s *Server,
	runtime PickupOblivionRuntime53A9C0,
) pickupOblivionNativeDeps53A9C0 {
	return pickupOblivionNativeDeps53A9C0{
		weaponPickup: runtime.WeaponPickup,
		playerState: func(owner *Object) int32 {
			if s.Players.CheckXxx(owner) {
				return 1
			}
			return 0
		},
		priorityMessage: func(owner *Object, message string, value uint8) {
			s.NetPriMsgToPlayer(owner, strman.ID(message), value)
		},
		audio: func(id uint32, owner *Object, kind int32, code uint32) {
			s.Audio.EventObj(sound.ID(id), owner, int(kind), code)
		},
		pauseFX:  runtime.PauseFX,
		tryEquip: runtime.TryEquip,
	}
}

// PickupOblivion53A9C0 binds the registered OblivionPickup callback to
// native-width Object fields and server services.
func (s *Server) PickupOblivion53A9C0(
	owner, item *Object,
	arg3, arg4 int32,
	runtime PickupOblivionRuntime53A9C0,
) int32 {
	return pickupOblivionNative53A9C0(
		owner,
		item,
		arg3,
		arg4,
		pickupOblivionServerDeps53A9C0(s, runtime),
	)
}
