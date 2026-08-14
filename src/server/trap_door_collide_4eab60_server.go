package server

import (
	"math"
	"unsafe"

	"github.com/opennox/libs/object"
	"github.com/opennox/libs/types"
)

// TrapDoorCollideData is the fixed-width 28-byte record registered for
// TrapDoorCollide. Its leading Script field is passed directly as the event
// block; no field depends on the host pointer width.
type TrapDoorCollideData struct {
	Script        ScriptCallback
	FallVelocityX int32
	FallVelocityY int32
	NextFrame     uint32
	Delay         uint16
	Reserved22    uint16
	Activated     uint32
}

type TrapDoorCollideScriptCaller4EAB60 func(
	block *ScriptCallback,
	caller *Object,
	trigger *Object,
	event ScriptEventType,
) unsafe.Pointer

type TrapDoorCollideRuntime4EAB60 struct {
	ScriptCallback TrapDoorCollideScriptCaller4EAB60
}

type trapDoorCollideNativeDeps4EAB60 struct {
	mapPointInBox  func(*types.Pointf, *Shape, *types.Pointf) bool
	abilityActive  func(*Object, int32) int32
	gameFrame      func() uint32
	scriptCallback TrapDoorCollideScriptCaller4EAB60
}

func trapDoorCollideNative4EAB60(
	source, target *Object,
	collision *types.Pointf,
	deps trapDoorCollideNativeDeps4EAB60,
) {
	trapDoorCollide4EAB60(
		source,
		target,
		collision,
		trapDoorCollideHooks4EAB60[*Object, *TrapDoorCollideData]{
			loadCollideData: func(obj *Object) *TrapDoorCollideData {
				return (*TrapDoorCollideData)(obj.CollideData)
			},
			loadClass: func(obj *Object) uint32 {
				return uint32(obj.ObjClass)
			},
			loadFlags: func(obj *Object) uint32 {
				return uint32(obj.ObjFlags)
			},
			loadShapeKind: func(obj *Object) uint32 {
				return uint32(obj.Shape.Kind)
			},
			loadBoxWidth: func(obj *Object) float32 {
				return obj.Shape.Box.W
			},
			loadBoxHeight: func(obj *Object) float32 {
				return obj.Shape.Box.H
			},
			loadCircleRadius: func(obj *Object) float32 {
				return obj.Shape.Circle.R
			},
			mapPointInBox: func(source, target *Object) bool {
				return deps.mapPointInBox(&source.PosVec, &source.Shape, &target.PosVec)
			},
			orFlags: func(obj *Object, flags uint32) {
				obj.ObjFlags |= object.Flags(flags)
			},
			loadFallVelocityX: func(data *TrapDoorCollideData) int32 {
				return data.FallVelocityX
			},
			storeFallVelocityX: func(obj *Object, value float32) {
				obj.Field41 = math.Float32bits(value)
			},
			loadFallVelocityY: func(data *TrapDoorCollideData) int32 {
				return data.FallVelocityY
			},
			storeFallVelocityY: func(obj *Object, value float32) {
				obj.Field42 = math.Float32bits(value)
			},
			loadPosX: func(obj *Object) float32 {
				return obj.PosVec.X
			},
			storeFallPosX: func(obj *Object, value float32) {
				obj.Pos39.X = value
			},
			loadPosY: func(obj *Object) float32 {
				return obj.PosVec.Y
			},
			storeFallPosY: func(obj *Object, value float32) {
				obj.Pos39.Y = value
			},
			loadActivated: func(data *TrapDoorCollideData) uint32 {
				return data.Activated
			},
			abilityActive: deps.abilityActive,
			loadDelay: func(data *TrapDoorCollideData) uint16 {
				return data.Delay
			},
			gameFrame: deps.gameFrame,
			storeNextFrame: func(data *TrapDoorCollideData, value uint32) {
				data.NextFrame = value
			},
			scriptCallback: func(
				data *TrapDoorCollideData,
				caller, trigger *Object,
				event ScriptEventType,
			) {
				_ = deps.scriptCallback(&data.Script, caller, trigger, event)
			},
			storeActivated: func(data *TrapDoorCollideData, value uint32) {
				data.Activated = value
			},
		},
	)
}

// TrapDoorCollide4EAB60 binds GAME.EXE's registered callback to native-width
// Object pointers while preserving the fixed-width collide record and existing
// script, frame and ability runtimes.
func (s *Server) TrapDoorCollide4EAB60(
	source, target *Object,
	collision *types.Pointf,
	runtime TrapDoorCollideRuntime4EAB60,
) {
	trapDoorCollideNative4EAB60(source, target, collision, trapDoorCollideNativeDeps4EAB60{
		mapPointInBox: MapPointInBox57B850,
		abilityActive: func(obj *Object, ability int32) int32 {
			if s.Abils.IsActive(obj, Ability(ability)) {
				return 1
			}
			return 0
		},
		gameFrame:      s.Frame,
		scriptCallback: runtime.ScriptCallback,
	})
}

var (
	_ = [1]struct{}{}[28-unsafe.Sizeof(TrapDoorCollideData{})]
	_ = [1]struct{}{}[0-unsafe.Offsetof(TrapDoorCollideData{}.Script)]
	_ = [1]struct{}{}[8-unsafe.Offsetof(TrapDoorCollideData{}.FallVelocityX)]
	_ = [1]struct{}{}[12-unsafe.Offsetof(TrapDoorCollideData{}.FallVelocityY)]
	_ = [1]struct{}{}[16-unsafe.Offsetof(TrapDoorCollideData{}.NextFrame)]
	_ = [1]struct{}{}[20-unsafe.Offsetof(TrapDoorCollideData{}.Delay)]
	_ = [1]struct{}{}[22-unsafe.Offsetof(TrapDoorCollideData{}.Reserved22)]
	_ = [1]struct{}{}[24-unsafe.Offsetof(TrapDoorCollideData{}.Activated)]
)
