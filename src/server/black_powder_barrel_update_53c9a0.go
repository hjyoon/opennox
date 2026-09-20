package server

import (
	"github.com/opennox/libs/object"
	"github.com/opennox/libs/types"
)

const (
	blackPowderBarrelDamageOuterRadius53C9A0 = float32(100)
	blackPowderBarrelDamageInnerRadius53C9A0 = float32(30)
	blackPowderBarrelDamage53C9A0            = 30
	blackPowderBarrelPushForce53C9A0         = float32(60)
	blackPowderBarrelFlameCount53C9A0        = 4
	blackPowderBarrelLifetimeSeconds53C9A0   = uint32(1)
)

var blackPowderBarrelFlameTypes53C9A0 = [...]string{"SmallFlame", "MediumFlame"}

// BlackPowderBarrelUpdateRuntime53C9A0 supplies the world effects reached by
// the native-width restoration of GAME.EXE 0053C9A0.
type BlackPowderBarrelUpdateRuntime53C9A0 struct {
	DamageUnitsAround func(types.Pointf, float32, float32, int, object.DamageType, *Object, Obj)
	PushUnitsAround   func(types.Pointf, float32, float32, float32)
	RandomInt         func(int, int) int
	RandomFloat       func(float32, float32) float32
	TraceRay          func(types.Pointf, types.Pointf) bool
	NewObjectByTypeID func(string) *Object
	CreateAt          func(*Object, *Object, types.Pointf)
	SetDecayTime      func(*Object, uint32)
	DelayedDelete     func(*Object)
}

type blackPowderBarrelUpdateDeps53C9A0 struct {
	frame             func() uint32
	fps               func() uint32
	damageUnitsAround func(types.Pointf, float32, float32, int, object.DamageType, *Object, Obj)
	pushUnitsAround   func(types.Pointf, float32, float32, float32)
	randomInt         func(int, int) int
	randomFloat       func(float32, float32) float32
	traceRay          func(types.Pointf, types.Pointf) bool
	newObjectByTypeID func(string) *Object
	createAt          func(*Object, *Object, types.Pointf)
	setDecayTime      func(*Object, uint32)
	delayedDelete     func(*Object)
}

func blackPowderBarrelUpdateNative53C9A0(source *Object, deps blackPowderBarrelUpdateDeps53C9A0) {
	fuseFrame := source.Field34
	creationFrame := source.Field32
	if fuseFrame == creationFrame {
		source.Field34 = deps.frame() + uint32(deps.randomInt(1, 5))
		return
	}
	if fuseFrame == deps.frame() {
		deps.damageUnitsAround(
			source.PosVec,
			blackPowderBarrelDamageOuterRadius53C9A0,
			blackPowderBarrelDamageInnerRadius53C9A0,
			blackPowderBarrelDamage53C9A0,
			object.DamageExplosion,
			source,
			nil,
		)
		deps.pushUnitsAround(
			source.PosVec,
			blackPowderBarrelDamageOuterRadius53C9A0,
			blackPowderBarrelDamageInnerRadius53C9A0,
			blackPowderBarrelPushForce53C9A0,
		)
		for range blackPowderBarrelFlameCount53C9A0 {
			flameType := blackPowderBarrelFlameTypes53C9A0[deps.randomInt(0, 1)]
			distance := deps.randomFloat(0, 15) + 10
			direction := byte(deps.randomInt(0, 255))
			cosine, sine := SinCosDir(direction)
			position := types.Ptf(
				source.PosVec.X+distance*cosine,
				source.PosVec.Y+distance*sine,
			)
			if !deps.traceRay(source.PosVec, position) {
				continue
			}
			flame := deps.newObjectByTypeID(flameType)
			if flame == nil {
				continue
			}
			deps.createAt(flame, nil, position)
			deps.setDecayTime(flame, deps.fps()*uint32(deps.randomInt(5, 20)))
		}
		return
	}
	if deps.frame()-creationFrame >= blackPowderBarrelLifetimeSeconds53C9A0*deps.fps() {
		deps.delayedDelete(source)
	}
}

// BlackPowderBarrelUpdate53C9A0 restores GAME.EXE 0053C9A0 without reading a
// native-width Object through PE32 field offsets.
func (s *Server) BlackPowderBarrelUpdate53C9A0(source *Object, runtime BlackPowderBarrelUpdateRuntime53C9A0) {
	if source == nil {
		return
	}
	blackPowderBarrelUpdateNative53C9A0(source, blackPowderBarrelUpdateDeps53C9A0{
		frame:             s.Frame,
		fps:               s.TickRate,
		damageUnitsAround: runtime.DamageUnitsAround,
		pushUnitsAround:   runtime.PushUnitsAround,
		randomInt:         runtime.RandomInt,
		randomFloat:       runtime.RandomFloat,
		traceRay:          runtime.TraceRay,
		newObjectByTypeID: runtime.NewObjectByTypeID,
		createAt:          runtime.CreateAt,
		setDecayTime:      runtime.SetDecayTime,
		delayedDelete:     runtime.DelayedDelete,
	})
}
