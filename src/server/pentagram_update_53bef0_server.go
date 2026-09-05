package server

import (
	"unsafe"

	"github.com/opennox/libs/noxnet/netmsg"
	"github.com/opennox/libs/object"
	"github.com/opennox/libs/types"

	"github.com/opennox/opennox/v1/common/sound"
)

// PentagramUpdateRuntime53BEF0 supplies the already restored 004E7190
// teleport gate at the legacy boundary. All other dependencies are native
// Server operations.
type PentagramUpdateRuntime53BEF0 struct {
	Teleport func(*Object, *types.Pointf)
}

type pentagramUpdateNativeDeps53BEF0 struct {
	eachInRect func(types.Rectf, func(*Object))
	pointFX    func(uint32, *types.Pointf)
	audio      func(uint32, *Object)
	teleport   func(*Object, *types.Pointf)
}

func pentagramUpdateNativeHooks53BEF0(
	s *Server,
	deps pentagramUpdateNativeDeps53BEF0,
) pentagramUpdateHooks53BEF0[*Object, *PentagramUpdateData, *types.Pointf] {
	visibleHooks := pentagramTeleportHooks53C060[*Object, *types.Pointf]{
		loadClass: func(obj *Object) uint32 {
			return uint32(obj.ObjClass)
		},
		cachePosition: func(obj *Object) *types.Pointf {
			return &obj.PosVec
		},
		pointFX:  deps.pointFX,
		audio:    deps.audio,
		teleport: deps.teleport,
	}
	return pentagramUpdateHooks53BEF0[*Object, *PentagramUpdateData, *types.Pointf]{
		loadUpdate: func(obj *Object) *PentagramUpdateData {
			return obj.UpdateDataPentagram()
		},
		loadState: func(data *PentagramUpdateData) uint8 {
			return data.State
		},
		storeState: func(data *PentagramUpdateData, value uint8) {
			data.State = value
		},
		loadTriggered: func(data *PentagramUpdateData) uint32 {
			return data.Triggered
		},
		storeTriggered: func(data *PentagramUpdateData, value uint32) {
			data.Triggered = value
		},
		loadAnimationFrame: func(data *PentagramUpdateData) uint8 {
			return data.AnimationFrame
		},
		storeAnimationFrame: func(data *PentagramUpdateData, value uint8) {
			data.AnimationFrame = value
		},
		loadAnimationTick: func(data *PentagramUpdateData) uint8 {
			return data.AnimationTick
		},
		storeAnimationTick: func(data *PentagramUpdateData, value uint8) {
			data.AnimationTick = value
		},
		loadAnimationStep: func(data *PentagramUpdateData) uint8 {
			return data.AnimationStep
		},
		storeAnimationStep: func(data *PentagramUpdateData, value uint8) {
			data.AnimationStep = value
		},
		needSync: (*Object).NeedSync,
		loadDestination: func(obj *Object, data *PentagramUpdateData) *Object {
			return obj.PentagramDestinationFor(data)
		},
		loadRadius: func(obj *Object) float32 {
			return obj.Shape.Circle.R
		},
		loadPosX: func(obj *Object) float32 {
			return obj.PosVec.X
		},
		loadPosY: func(obj *Object) float32 {
			return obj.PosVec.Y
		},
		cachePosition: func(obj *Object) *types.Pointf {
			return &obj.PosVec
		},
		eachInRect: deps.eachInRect,
		teleportVisible: func(unit *Object, destination *types.Pointf) {
			pentagramTeleport53C060(unit, destination, visibleHooks)
		},
		teleportInvisible: func(unit *Object, destination *types.Pointf) {
			pentagramTeleportInvisible53C140(
				unit,
				destination,
				func(obj *Object) uint32 { return uint32(obj.ObjClass) },
				deps.teleport,
			)
		},
		isEnabled: func(obj *Object) bool {
			return obj.ObjFlags.Has(object.FlagEnabled)
		},
		frame: s.Frame,
		storeField34: func(obj *Object, value uint32) {
			obj.Field34 = value
		},
	}
}

func pentagramUpdateServerDeps53BEF0(
	s *Server,
	runtime PentagramUpdateRuntime53BEF0,
) pentagramUpdateNativeDeps53BEF0 {
	return pentagramUpdateNativeDeps53BEF0{
		eachInRect: func(rect types.Rectf, callback func(*Object)) {
			s.Map.EachObjInRect(rect, func(obj *Object) bool {
				callback(obj)
				return true
			})
		},
		pointFX: func(id uint32, position *types.Pointf) {
			s.Nox_xxx_netSendPointFx_522FF0(netmsg.Op(id), *position)
		},
		audio: func(id uint32, obj *Object) {
			s.Audio.EventObj(sound.ID(id), obj, 0, 0)
		},
		teleport: runtime.Teleport,
	}
}

// TeleportPentagramUpdate53BEF0 binds the fixed PE32 update record to the
// native destination sidecar and the restored visible Pentagram semantics.
func (s *Server) TeleportPentagramUpdate53BEF0(
	pentagram *Object,
	runtime PentagramUpdateRuntime53BEF0,
) int32 {
	return pentagramUpdate53BEF0(
		pentagram,
		pentagramUpdateNativeHooks53BEF0(
			s,
			pentagramUpdateServerDeps53BEF0(s, runtime),
		),
	)
}

// InvisiblePentagramUpdate53C0C0 binds the same fixed update record and
// destination sidecar to the effect-free Pentagram semantics.
func (s *Server) InvisiblePentagramUpdate53C0C0(
	pentagram *Object,
	runtime PentagramUpdateRuntime53BEF0,
) int32 {
	return pentagramInvisibleUpdate53C0C0(
		pentagram,
		pentagramUpdateNativeHooks53BEF0(
			s,
			pentagramUpdateServerDeps53BEF0(s, runtime),
		),
	)
}

// PentagramTeleportUnit53C060 exposes the original visible enumeration
// callback with native-width object and point aliases.
func (s *Server) PentagramTeleportUnit53C060(
	unit *Object,
	destination *types.Pointf,
	runtime PentagramUpdateRuntime53BEF0,
) {
	deps := pentagramUpdateServerDeps53BEF0(s, runtime)
	pentagramTeleport53C060(
		unit,
		destination,
		pentagramTeleportHooks53C060[*Object, *types.Pointf]{
			loadClass: func(obj *Object) uint32 {
				return uint32(obj.ObjClass)
			},
			cachePosition: func(obj *Object) *types.Pointf {
				return &obj.PosVec
			},
			pointFX:  deps.pointFX,
			audio:    deps.audio,
			teleport: deps.teleport,
		},
	)
}

// PentagramTeleportUnitInvisible53C140 exposes the original effect-free
// enumeration callback with native-width object and point aliases.
func (s *Server) PentagramTeleportUnitInvisible53C140(
	unit *Object,
	destination *types.Pointf,
	runtime PentagramUpdateRuntime53BEF0,
) {
	pentagramTeleportInvisible53C140(
		unit,
		destination,
		func(obj *Object) uint32 { return uint32(obj.ObjClass) },
		runtime.Teleport,
	)
}

var (
	_ = [1]struct{}{}[24-unsafe.Sizeof(PentagramUpdateData{})]
	_ = [1]struct{}{}[12-unsafe.Offsetof(PentagramUpdateData{}.DestinationPE32)]
	_ = [1]struct{}{}[16-unsafe.Offsetof(PentagramUpdateData{}.DestinationExtent)]
	_ = [1]struct{}{}[20-unsafe.Offsetof(PentagramUpdateData{}.AnimationStep)]
)
