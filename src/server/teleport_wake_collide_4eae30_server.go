package server

import (
	"unsafe"

	"github.com/opennox/libs/noxnet/netmsg"
	"github.com/opennox/libs/types"

	noxflags "github.com/opennox/opennox/v1/common/flags"
	"github.com/opennox/opennox/v1/common/sound"
)

// TeleportWakeCollideData is the fixed-width two-float destination record
// initialized by the TeleportWake spell producer and registered for
// TeleportWakeCollide.
type TeleportWakeCollideData struct {
	Destination types.Pointf
}

type TeleportWakeCollideRuntime4EAE30 struct {
	Teleport func(*Object, *types.Pointf)
}

type teleportWakeCollideNativeDeps4EAE30 struct {
	hasEnchant func(*Object, EnchantID) bool
	questMode  func() bool
	pointFX    func(uint32, *types.Pointf)
	audio      func(uint32, *Object)
	teleport   func(*Object, *types.Pointf)
}

func teleportWakeCollideNative4EAE30(
	source, target *Object,
	collision *types.Pointf,
	deps teleportWakeCollideNativeDeps4EAE30,
) {
	teleportWakeCollide4EAE30(
		source,
		target,
		collision,
		teleportWakeCollideHooks4EAE30[*Object, *TeleportWakeCollideData, *types.Pointf]{
			loadCollideData: func(obj *Object) *TeleportWakeCollideData {
				return (*TeleportWakeCollideData)(obj.CollideData)
			},
			hasEnchant: func(obj *Object, enchant uint32) bool {
				return deps.hasEnchant(obj, EnchantID(enchant))
			},
			questMode: deps.questMode,
			loadOwner: func(obj *Object) *Object {
				return obj.ObjOwner
			},
			loadOwnerClassLo: func(obj *Object) uint8 {
				return uint8(obj.ObjClass)
			},
			loadTargetClass: func(obj *Object) uint32 {
				return uint32(obj.ObjClass)
			},
			position: func(obj *Object) *types.Pointf {
				return &obj.PosVec
			},
			pointFX: deps.pointFX,
			audio:   deps.audio,
			teleport: func(obj *Object, destination *TeleportWakeCollideData) {
				deps.teleport(obj, (*types.Pointf)(unsafe.Pointer(destination)))
			},
		},
	)
}

func teleportWakeCollideServerDeps4EAE30(
	s *Server,
	runtime TeleportWakeCollideRuntime4EAE30,
) teleportWakeCollideNativeDeps4EAE30 {
	return teleportWakeCollideNativeDeps4EAE30{
		hasEnchant: (*Object).HasEnchant,
		questMode: func() bool {
			return noxflags.HasGame(noxflags.GameModeQuest)
		},
		pointFX: func(id uint32, pos *types.Pointf) {
			s.Nox_xxx_netSendPointFx_522FF0(netmsg.Op(id), *pos)
		},
		audio: func(id uint32, obj *Object) {
			s.Audio.EventObj(sound.ID(id), obj, 0, 0)
		},
		teleport: runtime.Teleport,
	}
}

// TeleportWakeCollide4EAE30 binds GAME.EXE's registered callback to
// native-width Object and Pointf pointers. Runtime Teleport supplies the
// already restored 004E7190 gate at the legacy boundary.
func (s *Server) TeleportWakeCollide4EAE30(
	source, target *Object,
	collision *types.Pointf,
	runtime TeleportWakeCollideRuntime4EAE30,
) {
	teleportWakeCollideNative4EAE30(
		source,
		target,
		collision,
		teleportWakeCollideServerDeps4EAE30(s, runtime),
	)
}

var (
	_ = [1]struct{}{}[8-unsafe.Sizeof(TeleportWakeCollideData{})]
	_ = [1]struct{}{}[0-unsafe.Offsetof(TeleportWakeCollideData{}.Destination)]
	_ = [1]struct{}{}[8-unsafe.Sizeof(types.Pointf{})]
	_ = [1]struct{}{}[0-unsafe.Offsetof(types.Pointf{}.X)]
	_ = [1]struct{}{}[4-unsafe.Offsetof(types.Pointf{}.Y)]
)
