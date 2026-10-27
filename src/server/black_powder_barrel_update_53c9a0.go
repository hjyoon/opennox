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
	DamageUnitsAround    func(types.Pointf, float32, float32, int, object.DamageType, *Object, Obj)
	PushUnitsAround      func(types.Pointf, float32, float32, float32)
	RandomInt            func(int, int) int
	RandomFloat          func(float32, float32) float64
	TraceRay             func(types.Pointf, types.Pointf) bool
	LoadFlameType        func(int) int
	StoreFlameType       func(int, int)
	TypeIndexByID        func(string) int
	NewObjectByTypeIndex func(int) *Object
	CreateAt             func(*Object, *Object, types.Pointf)
	SetDecayTime         func(*Object, uint32)
	DelayedDelete        func(*Object)
}

type blackPowderBarrelUpdateDeps53C9A0 struct {
	frame                func() uint32
	fps                  func() uint32
	damageUnitsAround    func(types.Pointf, float32, float32, int, object.DamageType, *Object, Obj)
	pushUnitsAround      func(types.Pointf, float32, float32, float32)
	randomInt            func(int, int) int
	randomFloat          func(float32, float32) float64
	traceRay             func(types.Pointf, types.Pointf) bool
	loadFlameType        func(int) int
	storeFlameType       func(int, int)
	typeIndexByID        func(string) int
	newObjectByTypeIndex func(int) *Object
	createAt             func(*Object, *Object, types.Pointf)
	setDecayTime         func(*Object, uint32)
	delayedDelete        func(*Object)
}

func blackPowderBarrelUpdateNative53C9A0(source *Object, deps blackPowderBarrelUpdateDeps53C9A0) {
	fuseFrame := source.Field34
	creationFrame := source.Field32
	if fuseFrame == creationFrame {
		delay := uint32(deps.randomInt(1, 5))
		source.Field34 = deps.frame() + delay
		return
	}
	if fuseFrame == deps.frame() {
		if deps.loadFlameType(0) == 0 {
			for index, name := range blackPowderBarrelFlameTypes53C9A0 {
				deps.storeFlameType(index, deps.typeIndexByID(name))
			}
		}
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
		// The ray origin is captured after push; each destination reloads the
		// live source after RNG callbacks, including on later flame iterations.
		traceOrigin := source.PosVec
		for range blackPowderBarrelFlameCount53C9A0 {
			flameType := deps.randomInt(0, 1)
			// The RNG result remains wide through +10, then FSTP stores binary32.
			distance := experienceSpill32TowardZero(objectForceAdd53_52DF80(deps.randomFloat(0, 15), 10))
			direction := byte(deps.randomInt(0, 255))
			cosine, sine := SinCosDir(direction)
			position := types.Ptf(
				experienceSpill32TowardZero(objectForceAdd53_52DF80(objectForceMul53_52DF80(float64(distance), float64(cosine)), objectForceMemory32_52DF80(source.PosVec.X))),
				experienceSpill32TowardZero(objectForceAdd53_52DF80(objectForceMul53_52DF80(float64(distance), float64(sine)), objectForceMemory32_52DF80(source.PosVec.Y))),
			)
			if !deps.traceRay(traceOrigin, position) {
				continue
			}
			flame := deps.newObjectByTypeIndex(deps.loadFlameType(flameType))
			if flame == nil {
				continue
			}
			deps.createAt(flame, nil, position)
			seconds := uint32(deps.randomInt(5, 20))
			deps.setDecayTime(flame, deps.fps()*seconds)
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
		frame:                s.Frame,
		fps:                  s.TickRate,
		damageUnitsAround:    runtime.DamageUnitsAround,
		pushUnitsAround:      runtime.PushUnitsAround,
		randomInt:            runtime.RandomInt,
		randomFloat:          runtime.RandomFloat,
		traceRay:             runtime.TraceRay,
		loadFlameType:        runtime.LoadFlameType,
		storeFlameType:       runtime.StoreFlameType,
		typeIndexByID:        runtime.TypeIndexByID,
		newObjectByTypeIndex: runtime.NewObjectByTypeIndex,
		createAt:             runtime.CreateAt,
		setDecayTime:         runtime.SetDecayTime,
		delayedDelete:        runtime.DelayedDelete,
	})
}
