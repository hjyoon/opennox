package server

import (
	"image"
	"math"

	"github.com/opennox/libs/object"
	"github.com/opennox/libs/types"
)

// MeteorShowerRawRuntime53D5A0 preserves native pointer identities and the
// original trace result's low-byte contract at world service boundaries.
type MeteorShowerRawRuntime53D5A0 struct {
	Frame, FPS        func() uint32
	RandomInt         func(int, int) int
	RandomFloat       func(float64, float64) float64
	DirectionFloat    func(int32) float32
	TraceRay          func(*types.Rectf, *Object) int32
	NewObjectByTypeID func(string) *Object
	CreateAt          func(*Object, *Object, types.Pointf)
	Raise             func(*Object, float32)
	DelayedDelete     func(*Object)
}

// MeteorRawRuntime53D6E0 keeps the source's position pointer live across the
// impact services. The original damage record is cached separately.
type MeteorRawRuntime53D6E0 struct {
	AudioEvent            func(uint32, *Object)
	MakeScorch            func(*types.Pointf, int)
	Earthquake            func(*types.Pointf, int)
	NewObjectByTypeID     func(string) *Object
	CreateAt              func(*Object, *Object, types.Pointf)
	FindParentChainPlayer func(*Object) *Object
	DamageUnitsAround     func(*types.Pointf, float32, float32, int, object.DamageType, *Object, Obj)
	DamageWalls           func(image.Rectangle, *types.Pointf, float32, int, object.DamageType, *Object)
	DelayedDelete         func(*Object)
}

// MeteorShowerUpdateRaw53D5A0 preserves the required-pointer C entry contract.
func MeteorShowerUpdateRaw53D5A0(source *Object, h MeteorShowerRawRuntime53D5A0) {
	frame, fps := h.Frame(), h.FPS()
	if frame-source.Field32 >= meteorShowerLifetimeSeconds53D5A0*fps {
		h.DelayedDelete(source)
		return
	}
	if source.Field34 > frame {
		return
	}
	data := source.UpdateDataMeteor()
	direction := int32(h.RandomInt(0, 255))
	radius := h.RandomFloat(meteorShowerRadiusMinimum53D5A0, meteorShowerRadiusMaximum53D5A0)
	square := objectForceMul53_52DF80(radius, radius)
	from := source.PosVec
	x := objectForceMul53_52DF80(square, objectForceMemory32_52DF80(h.DirectionFloat(direction*2)))
	x = objectForceAdd53_52DF80(x, objectForceMemory32_52DF80(source.PosVec.X))
	destination := types.Ptf(experienceSpill32TowardZero(x), 0)
	y := objectForceMul53_52DF80(square, objectForceMemory32_52DF80(h.DirectionFloat(direction*2+1)))
	y = objectForceAdd53_52DF80(y, objectForceMemory32_52DF80(source.PosVec.Y))
	destination.Y = experienceSpill32TowardZero(y)
	ray := types.Rectf{Min: from, Max: destination}
	if uint8(h.TraceRay(&ray, source)) != 0 {
		if meteor := h.NewObjectByTypeID(meteorShowerProjectileTypeID53D5A0); meteor != nil {
			spawnData := meteor.UpdateDataMeteor()
			spawnData.Damage = data.Damage
			// Trace can write its rectangle; CreateAt uses the separately retained
			// pre-trace destination words in the original routine.
			h.CreateAt(meteor, source, destination)
			meteor.Field5 |= meteorFallingXStatus53D5A0
			h.Raise(meteor, meteorSpawnHeight53D5A0)
			meteor.Field27 = meteorSpawnVelocity53D5A0
		}
	}
	delay := uint32(h.RandomInt(meteorShowerDelayMinimum53D5A0, meteorShowerDelayMaximum53D5A0))
	source.Field34 = h.Frame() + delay
}

// MeteorUpdateRaw53D6E0 keeps the original unordered impact comparison,
// cached damage-record pointer and live reads between world effects.
func MeteorUpdateRaw53D6E0(source *Object, h MeteorRawRuntime53D6E0) {
	if !(source.ZVal <= 0 || math.IsNaN(float64(source.ZVal))) {
		return
	}
	data := source.UpdateDataMeteor()
	h.AudioEvent(meteorImpactAudio53D6E0, source)
	h.MakeScorch(&source.PosVec, meteorImpactScorch53D6E0)
	h.Earthquake(&source.PosVec, meteorImpactEarthquake53D6E0)
	if explosion := h.NewObjectByTypeID(meteorImpactExplosionTypeID53D6E0); explosion != nil {
		h.CreateAt(explosion, nil, source.PosVec)
	}
	player := h.FindParentChainPlayer(source)
	h.DamageUnitsAround(&source.PosVec, meteorImpactOuterRadius53D6E0, meteorImpactInnerRadius53D6E0, int(data.Damage), meteorImpactDamageType53D6E0, player, nil)
	// All four binary32 spills precede the actual FISTP conversions. Preserve
	// raw Min/Max order even when conversion overflow reverses the interval.
	x1 := experienceSpill32TowardZero(objectForceSub53_52DF80(float64(source.PosVec.X), float64(meteorImpactOuterRadius53D6E0)))
	y1 := experienceSpill32TowardZero(objectForceSub53_52DF80(float64(source.PosVec.Y), float64(meteorImpactOuterRadius53D6E0)))
	x2 := experienceSpill32TowardZero(objectForceAdd53_52DF80(float64(source.PosVec.X), float64(meteorImpactOuterRadius53D6E0)))
	y2 := experienceSpill32TowardZero(objectForceAdd53_52DF80(float64(source.PosVec.Y), float64(meteorImpactOuterRadius53D6E0)))
	rect := image.Rectangle{
		Min: image.Pt(int(meteorFloatToInt53D6E0(x1)/meteorImpactWallGridStep53D6E0), int(meteorFloatToInt53D6E0(y1)/meteorImpactWallGridStep53D6E0)),
		Max: image.Pt(int(meteorFloatToInt53D6E0(x2)/meteorImpactWallGridStep53D6E0), int(meteorFloatToInt53D6E0(y2)/meteorImpactWallGridStep53D6E0)),
	}
	h.DamageWalls(rect, &source.PosVec, meteorImpactOuterRadius53D6E0, int(data.Damage), meteorImpactDamageType53D6E0, source)
	h.DelayedDelete(source)
}
