package server

import (
	"math"
	"unsafe"

	"github.com/opennox/libs/noxnet/netmsg"
	"github.com/opennox/libs/types"

	"github.com/opennox/opennox/v1/common/sound"
)

// TeleportCollideData is the fixed-width eight-byte record registered for
// TeleportCollide. GAME.EXE interprets both coordinates as signed int32 before
// converting them to binary32 destination fields.
type TeleportCollideData struct {
	DestinationX int32
	DestinationY int32
}

type TeleportCollideRuntime4EACA0 struct {
	Teleport func(*Object, *types.Pointf)
}

type teleportCollideNativeDeps4EACA0 struct {
	pointFX  func(uint32, *types.Pointf)
	audio    func(uint32, *Object)
	teleport func(*Object, *types.Pointf)
}

func teleportCollideDestination4EACA0(obj *Object) *types.Pointf {
	return (*types.Pointf)(unsafe.Pointer(&obj.Field41))
}

func teleportCollideNative4EACA0(
	source, target *Object,
	collision *types.Pointf,
	deps teleportCollideNativeDeps4EACA0,
) {
	teleportCollide4EACA0(
		source,
		target,
		collision,
		teleportCollideHooks4EACA0[
			*Object,
			*TeleportCollideData,
			*types.Pointf,
			*types.Pointf,
		]{
			loadCollideData: func(obj *Object) *TeleportCollideData {
				return (*TeleportCollideData)(obj.CollideData)
			},
			loadClass: func(obj *Object) uint32 {
				return uint32(obj.ObjClass)
			},
			cachePosition: func(obj *Object) *types.Pointf {
				return &obj.PosVec
			},
			pointFX: deps.pointFX,
			audio:   deps.audio,
			loadDestinationX: func(data *TeleportCollideData) int32 {
				return data.DestinationX
			},
			cacheDestination: teleportCollideDestination4EACA0,
			storeDestinationX: func(obj *Object, value float32) {
				obj.Field41 = math.Float32bits(value)
			},
			loadDestinationY: func(data *TeleportCollideData) int32 {
				return data.DestinationY
			},
			storeDestinationY: func(obj *Object, value float32) {
				obj.Field42 = math.Float32bits(value)
			},
			teleport: deps.teleport,
		},
	)
}

func teleportCollideServerDeps4EACA0(
	s *Server,
	runtime TeleportCollideRuntime4EACA0,
) teleportCollideNativeDeps4EACA0 {
	return teleportCollideNativeDeps4EACA0{
		pointFX: func(id uint32, pos *types.Pointf) {
			s.Nox_xxx_netSendPointFx_522FF0(netmsg.Op(id), *pos)
		},
		audio: func(id uint32, obj *Object) {
			s.Audio.EventObj(sound.ID(id), obj, 0, 0)
		},
		teleport: runtime.Teleport,
	}
}

// TeleportCollide4EACA0 binds GAME.EXE's registered callback to native-width
// Object pointers while preserving its fixed-width destination record and the
// already restored 004E7190 teleport gate supplied by the legacy boundary.
func (s *Server) TeleportCollide4EACA0(
	source, target *Object,
	collision *types.Pointf,
	runtime TeleportCollideRuntime4EACA0,
) {
	teleportCollideNative4EACA0(
		source,
		target,
		collision,
		teleportCollideServerDeps4EACA0(s, runtime),
	)
}

var (
	_ = [1]struct{}{}[8-unsafe.Sizeof(TeleportCollideData{})]
	_ = [1]struct{}{}[0-unsafe.Offsetof(TeleportCollideData{}.DestinationX)]
	_ = [1]struct{}{}[4-unsafe.Offsetof(TeleportCollideData{}.DestinationY)]
	_ = [1]struct{}{}[8-unsafe.Sizeof(types.Pointf{})]
	_ = [1]struct{}{}[4-(unsafe.Offsetof(Object{}.Field42)-unsafe.Offsetof(Object{}.Field41))]
)
