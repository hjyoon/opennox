package server

import (
	"image"
	"unsafe"

	"github.com/opennox/libs/object"
	"github.com/opennox/libs/types"
)

const (
	meteorShowerLifetimeSeconds53D5A0 = uint32(5)
	meteorShowerDelayMinimum53D5A0    = 4
	meteorShowerDelayMaximum53D5A0    = 8
	meteorShowerRadiusMinimum53D5A0   = float64(4)
	meteorShowerRadiusMaximum53D5A0   = float64(12)
	meteorSpawnHeight53D5A0           = float32(255)
	meteorSpawnVelocity53D5A0         = float32(-8)
	meteorFallingXStatus53D5A0        = uint32(0x20)

	meteorImpactAudio53D6E0            = uint32(87)
	meteorImpactScorch53D6E0           = 2
	meteorImpactEarthquake53D6E0       = 10
	meteorImpactOuterRadius53D6E0      = float32(80)
	meteorImpactInnerRadius53D6E0      = float32(30)
	meteorImpactDamageType53D6E0       = object.DamageExplosion
	meteorImpactWallGridStep53D6E0     = int32(23)
	meteorImpactExplosionTypeID53D6E0  = "MeteorExplode"
	meteorShowerProjectileTypeID53D5A0 = "Meteor"
)

// MeteorUpdateData is the fixed-width damage record shared by
// MeteorShowerUpdate and MeteorUpdate.
type MeteorUpdateData struct {
	Damage int32
}

func (obj *Object) UpdateDataMeteor() *MeteorUpdateData {
	return (*MeteorUpdateData)(obj.UpdateData)
}

var (
	_ = [1]struct{}{}[4-unsafe.Sizeof(MeteorUpdateData{})]
	_ = [1]struct{}{}[0-unsafe.Offsetof(MeteorUpdateData{}.Damage)]
)

// MeteorShowerUpdateRuntime53D5A0 supplies the world operations reached by
// the native-width restoration of GAME.EXE 0053D5A0.
type MeteorShowerUpdateRuntime53D5A0 struct {
	RandomInt         func(int, int) int
	RandomFloat       func(float64, float64) float64
	TraceRay          func(types.Pointf, types.Pointf) bool
	NewObjectByTypeID func(string) *Object
	CreateAt          func(*Object, *Object, types.Pointf)
	Raise             func(*Object, float32)
	DelayedDelete     func(*Object)
}

type meteorShowerUpdateDeps53D5A0 struct {
	frame             func() uint32
	fps               func() uint32
	randomInt         func(int, int) int
	randomFloat       func(float64, float64) float64
	traceRay          func(types.Pointf, types.Pointf) bool
	newObjectByTypeID func(string) *Object
	createAt          func(*Object, *Object, types.Pointf)
	raise             func(*Object, float32)
	delayedDelete     func(*Object)
}

func meteorShowerUpdateNative53D5A0(source *Object, deps meteorShowerUpdateDeps53D5A0) {
	MeteorShowerUpdateRaw53D5A0(source, MeteorShowerRawRuntime53D5A0{
		Frame: deps.frame, FPS: deps.fps, RandomInt: deps.randomInt, RandomFloat: deps.randomFloat,
		DirectionFloat: func(index int32) float32 {
			cosine, sine := SinCosDir(byte(index >> 1))
			if index&1 == 0 {
				return cosine
			}
			return sine
		},
		TraceRay: func(r *types.Rectf, _ *Object) int32 {
			if deps.traceRay(r.Min, r.Max) {
				return 1
			}
			return 0
		},
		NewObjectByTypeID: deps.newObjectByTypeID, CreateAt: deps.createAt, Raise: deps.raise, DelayedDelete: deps.delayedDelete,
	})
}

// MeteorShowerUpdate53D5A0 restores GAME.EXE 0053D5A0 without passing the
// source, spawned Meteor, or their update-data pointers through PE32 C code.
func (s *Server) MeteorShowerUpdate53D5A0(source *Object, runtime MeteorShowerUpdateRuntime53D5A0) {
	if source == nil {
		return
	}
	meteorShowerUpdateNative53D5A0(source, meteorShowerUpdateDeps53D5A0{
		frame:             s.Frame,
		fps:               s.TickRate,
		randomInt:         runtime.RandomInt,
		randomFloat:       runtime.RandomFloat,
		traceRay:          runtime.TraceRay,
		newObjectByTypeID: runtime.NewObjectByTypeID,
		createAt:          runtime.CreateAt,
		raise:             runtime.Raise,
		delayedDelete:     runtime.DelayedDelete,
	})
}

// MeteorUpdateRuntime53D6E0 supplies the world effects reached by the
// native-width restoration of GAME.EXE 0053D6E0.
type MeteorUpdateRuntime53D6E0 struct {
	AudioEvent            func(uint32, *Object)
	MakeScorch            func(types.Pointf, int)
	Earthquake            func(types.Pointf, int)
	NewObjectByTypeID     func(string) *Object
	CreateAt              func(*Object, *Object, types.Pointf)
	FindParentChainPlayer func(*Object) *Object
	DamageUnitsAround     func(types.Pointf, float32, float32, int, object.DamageType, *Object, Obj, bool)
	DamageWalls           func(image.Rectangle, types.Pointf, float32, int, object.DamageType, *Object)
	DamageSource          func() bool
	DelayedDelete         func(*Object)
}

type meteorUpdateDeps53D6E0 struct {
	audioEvent            func(uint32, *Object)
	makeScorch            func(types.Pointf, int)
	earthquake            func(types.Pointf, int)
	newObjectByTypeID     func(string) *Object
	createAt              func(*Object, *Object, types.Pointf)
	findParentChainPlayer func(*Object) *Object
	damageUnitsAround     func(types.Pointf, float32, float32, int, object.DamageType, *Object, Obj, bool)
	damageWalls           func(image.Rectangle, types.Pointf, float32, int, object.DamageType, *Object)
	damageSource          func() bool
	delayedDelete         func(*Object)
}

func meteorFloatToInt53D6E0(value float32) int32 {
	return netFXTruncateSignedDWORD419A70(value)
}

func meteorUpdateNative53D6E0(source *Object, deps meteorUpdateDeps53D6E0) {
	MeteorUpdateRaw53D6E0(source, MeteorRawRuntime53D6E0{
		AudioEvent: deps.audioEvent, MakeScorch: func(p *types.Pointf, k int) { deps.makeScorch(*p, k) }, Earthquake: func(p *types.Pointf, k int) { deps.earthquake(*p, k) },
		NewObjectByTypeID: deps.newObjectByTypeID, CreateAt: deps.createAt, FindParentChainPlayer: deps.findParentChainPlayer,
		DamageUnitsAround: func(p *types.Pointf, a, b float32, d int, typ object.DamageType, o *Object, x Obj) {
			deps.damageUnitsAround(*p, a, b, d, typ, o, x, deps.damageSource())
		},
		DamageWalls: func(r image.Rectangle, p *types.Pointf, a float32, d int, typ object.DamageType, o *Object) {
			deps.damageWalls(r, *p, a, d, typ, o)
		},
		DelayedDelete: deps.delayedDelete,
	})
}

// MeteorUpdate53D6E0 restores GAME.EXE 0053D6E0 without reading the native
// Object and update-data records at their PE32 offsets.
func (s *Server) MeteorUpdate53D6E0(source *Object, runtime MeteorUpdateRuntime53D6E0) {
	if source == nil {
		return
	}
	meteorUpdateNative53D6E0(source, meteorUpdateDeps53D6E0{
		audioEvent:            runtime.AudioEvent,
		makeScorch:            runtime.MakeScorch,
		earthquake:            runtime.Earthquake,
		newObjectByTypeID:     runtime.NewObjectByTypeID,
		createAt:              runtime.CreateAt,
		findParentChainPlayer: runtime.FindParentChainPlayer,
		damageUnitsAround:     runtime.DamageUnitsAround,
		damageWalls:           runtime.DamageWalls,
		damageSource:          runtime.DamageSource,
		delayedDelete:         runtime.DelayedDelete,
	})
}
