package server

import (
	"math"

	"github.com/opennox/libs/types"
)

// SparkUpdateRaw53ADC0 retains the C entry's required object and update record.
func SparkUpdateRaw53ADC0(source *Object, delayedDelete func(*Object)) {
	data := (*SparkUpdateData)(source.UpdateData)
	if int32(data.LifetimeRemaining) <= 0 {
		delayedDelete(source)
		return
	}
	data.LifetimeRemaining--
	if data.Kind == 4 {
		source.Float28 = 1
	} else {
		source.Float28 = math.Float32frombits(1064514355)
	}
}

// ProjectileTrailRawRuntime53AEC0 preserves wide RNG results and the actual
// last spark pointer. DirectionFloat receives a signed DWORD table index.
type ProjectileTrailRawRuntime53AEC0 struct {
	DirectionFloat func(int32) float32
	RandomFloat    func(float32, float32) float64
	CreateSpark    func(types.Pointf, int, int, types.Pointf, float32, *Object) *Object
}

// ProjectileTrailUpdateRaw53AEC0 reproduces the original startup PC53/ToZero
// operations and binary32 spills locally, without changing host FP state.
func ProjectileTrailUpdateRaw53AEC0(source *Object, h ProjectileTrailRawRuntime53AEC0) *Object {
	direction := int32(int16(source.Direction1)) * 2
	x := objectForceMul53_52DF80(float64(h.DirectionFloat(direction)), objectForceMemory32_52DF80(source.SpeedCur))
	x = objectForceMul53_52DF80(x, 0.25)
	source.ForceVec.X = experienceSpill32TowardZero(objectForceAdd53_52DF80(x, objectForceMemory32_52DF80(source.ForceVec.X)))
	y := objectForceMul53_52DF80(float64(h.DirectionFloat(direction+1)), objectForceMemory32_52DF80(source.SpeedCur))
	y = objectForceMul53_52DF80(y, 0.25)
	source.ForceVec.Y = experienceSpill32TowardZero(objectForceAdd53_52DF80(y, objectForceMemory32_52DF80(source.ForceVec.Y)))
	stepX := experienceSpill32TowardZero(objectForceMul53_52DF80(objectForceSub53_52DF80(float64(source.PosVec.X), objectForceMemory32_52DF80(source.PrevPos.X)), 0.25))
	stepY := experienceSpill32TowardZero(objectForceMul53_52DF80(objectForceSub53_52DF80(float64(source.PosVec.Y), objectForceMemory32_52DF80(source.PrevPos.Y)), 0.25))
	var result *Object
	for i := 0; i < 4; i++ {
		// Steps remain cached, but the original reloads PrevPos for each pair.
		px := experienceSpill32TowardZero(objectForceAdd53_52DF80(objectForceMul53_52DF80(float64(i), objectForceMemory32_52DF80(stepX)), objectForceMemory32_52DF80(source.PrevPos.X)))
		py := experienceSpill32TowardZero(objectForceAdd53_52DF80(objectForceMul53_52DF80(float64(i), objectForceMemory32_52DF80(stepY)), objectForceMemory32_52DF80(source.PrevPos.Y)))
		baseX := experienceSpill32TowardZero(float64(netFXTruncateSignedDWORD419A70(px)))
		baseY := experienceSpill32TowardZero(float64(netFXTruncateSignedDWORD419A70(py)))
		for j := 0; j < 2; j++ {
			vy := experienceSpill32TowardZero(h.RandomFloat(-2, 2))
			vx := experienceSpill32TowardZero(h.RandomFloat(-2, 2))
			y := experienceSpill32TowardZero(objectForceAdd53_52DF80(h.RandomFloat(-4, 4), objectForceMemory32_52DF80(baseY)))
			x := experienceSpill32TowardZero(objectForceAdd53_52DF80(h.RandomFloat(-4, 4), objectForceMemory32_52DF80(baseX)))
			result = h.CreateSpark(types.Ptf(x, y), 1, 6, types.Ptf(vx, vy), 0, source)
		}
	}
	return result
}

// PushUpdateRaw53B030 keeps the position pointer live through the radial push,
// and reads Force before Radius from the required fixed-width record.
func PushUpdateRaw53B030(source *Object, push func(*types.Pointf, float32, float32, float32)) {
	data := (*PushUpdateData53B030)(source.UpdateData)
	force := data.Force
	radius := data.Radius
	push(&source.PosVec, radius, 0, force)
}
