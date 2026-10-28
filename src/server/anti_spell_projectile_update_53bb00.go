package server

import (
	"math"

	"github.com/opennox/libs/object"
	"github.com/opennox/libs/types"
)

// AntiSpellTargetState53BD10 owns the original shared search result at native
// pointer width. The best squared distance remains an original binary32 word.
type AntiSpellTargetState53BD10 struct {
	Target       *Object
	BestDistance float32
}

// AntiSpellProjectileRuntime53BB00 supplies the world effects for the
// native-width AntiSpellProjectileUpdate callback.
type AntiSpellProjectileRuntime53BB00 struct {
	TargetState     *AntiSpellTargetState53BD10
	Frame           func() uint32
	FPS             func() uint32
	EachMissile     func(types.Pointf, float32, func(*Object) bool)
	EachMissilePtr  func(*types.Pointf, float32, func(*Object) bool)
	CanInteract     func(*Object, *Object) bool
	DelayedDelete   func(*Object)
	Audio           func(*Object)
	RandomFloat     func(float32, float32) float32
	RandomFloatWide func(float32, float32) float64
	RandomInt       func(int, int) int
	CreateSpark     func(types.Pointf, int, int, types.Pointf, float32, *Object)
}

func antiSpellFindTarget53BD10(missile *Object, runtime AntiSpellProjectileRuntime53BB00) *Object {
	state := runtime.TargetState
	state.Target, state.BestDistance = nil, 100_000_000
	visit := func(candidate *Object) bool {
		AntiSpellCandidateRaw53BD10(candidate, missile, runtime)
		return true
	}
	if runtime.EachMissilePtr != nil {
		runtime.EachMissilePtr(&missile.PosVec, 600, visit)
	} else {
		runtime.EachMissile(missile.PosVec, 600, visit)
	}
	return state.Target
}

// AntiSpellCandidateRaw53BD10 retains the raw C callback's byte gates,
// owner-chain check and live shared search state, including reentrant queries.
func AntiSpellCandidateRaw53BD10(candidate, missile *Object, runtime AntiSpellProjectileRuntime53BB00) {
	if candidate.ObjFlags.Has(object.FlagDestroyed) ||
		uint8(candidate.ObjClass)&1 != 0 && uint8(candidate.ObjSubClass)&2 == 0 || candidate == missile {
		return
	}
	owner := missile.ObjOwner
	if owner != nil && candidate.HasOwner(owner) || !runtime.CanInteract(missile, candidate) {
		return
	}
	dx := objectForceSub53_52DF80(float64(missile.PosVec.X), objectForceMemory32_52DF80(candidate.PosVec.X))
	dy := objectForceSub53_52DF80(float64(missile.PosVec.Y), objectForceMemory32_52DF80(candidate.PosVec.Y))
	distance := objectForceAdd53_52DF80(objectForceMul53_52DF80(dy, dy), objectForceMul53_52DF80(dx, dx))
	state := runtime.TargetState
	// FCOM tests C0 alone: less and unordered both select the candidate.
	if !(distance >= float64(state.BestDistance)) {
		state.BestDistance = experienceSpill32TowardZero(distance)
		state.Target = candidate
	}
}

// AntiSpellProjectileUpdate53BB00 restores GAME.EXE 0053BB00 without reading
// the PE32 pointer in the missile's update record or its target object.
func AntiSpellProjectileUpdate53BB00(missile *Object, runtime AntiSpellProjectileRuntime53BB00) {
	if missile == nil || missile.UpdateData == nil {
		return
	}
	AntiSpellProjectileRaw53BB00(missile, runtime)
}

// AntiSpellProjectileRaw53BB00 preserves the C entry's required pointer reads.
// Native game callers supply their server's shared search state; independent
// Go API callers can omit it for a single non-reentrant search.
func AntiSpellProjectileRaw53BB00(missile *Object, runtime AntiSpellProjectileRuntime53BB00) {
	if runtime.TargetState == nil {
		runtime.TargetState = new(AntiSpellTargetState53BD10)
	}
	frame := runtime.Frame()
	created := missile.Field32
	update := missile.UpdateDataMissile()
	fps := runtime.FPS()
	if frame-created > 5*fps {
		runtime.DelayedDelete(missile)
		return
	}
	if update.Target != nil && update.Target.ObjFlags.Has(object.FlagDestroyed) {
		update.Target = nil
		frame, fps = runtime.Frame(), runtime.FPS()
	}
	if update.Target == nil && frame-missile.Field34 > fps>>2 {
		update.Target = antiSpellFindTarget53BD10(missile, runtime)
		missile.Field34 = runtime.Frame()
	}
	if target := update.Target; target != nil {
		dx := objectForceSub53_52DF80(float64(target.PosVec.X), objectForceMemory32_52DF80(missile.PosVec.X))
		dy := objectForceSub53_52DF80(float64(target.PosVec.Y), objectForceMemory32_52DF80(missile.PosVec.Y))
		distance := antiSpellDistance53BB00(dx, dy)
		antiSpellForce53BB00(missile, dx, dy, distance)
		if !(distance >= 10) {
			runtime.Audio(missile)
			runtime.DelayedDelete(missile)
			// Audio/deletion may replace Target, but the update pointer is cached.
			runtime.DelayedDelete(update.Target)
		}
	} else {
		velocity := missile.VelVec
		dx, dy := float64(velocity.X), float64(velocity.Y)
		antiSpellForce53BB00(missile, dx, dy, antiSpellDistance53BB00(dx, dy))
	}
	position := missile.PosVec
	randomFloat := runtime.RandomFloatWide
	if randomFloat == nil {
		randomFloat = func(min, max float32) float64 { return float64(runtime.RandomFloat(min, max)) }
	}
	velocityY := experienceSpill32TowardZero(randomFloat(-2, 2))
	velocityX := experienceSpill32TowardZero(randomFloat(-2, 2))
	lifetime := runtime.RandomInt(15, 30)
	posY := experienceSpill32TowardZero(objectForceAdd53_52DF80(randomFloat(-4, 4), objectForceMemory32_52DF80(position.Y)))
	posX := experienceSpill32TowardZero(objectForceAdd53_52DF80(randomFloat(-4, 4), objectForceMemory32_52DF80(position.X)))
	runtime.CreateSpark(types.Ptf(posX, posY), 3, lifetime, types.Ptf(velocityX, velocityY), 0, missile)
}

func antiSpellDistance53BB00(dx, dy float64) float64 {
	square := objectForceAdd53_52DF80(objectForceMul53_52DF80(dy, dy), objectForceMul53_52DF80(dx, dx))
	return objectForceAdd53_52DF80(monsterMoveForceSqrtChop53_50D50C(square), float64(math.Float32frombits(0x3dcccccd)))
}

func antiSpellForce53BB00(missile *Object, dx, dy, distance float64) {
	missile.ForceVec.X = experienceSpill32TowardZero(objectForceDiv53_52DF80(objectForceMul53_52DF80(dx, objectForceMemory32_52DF80(missile.SpeedCur)), distance))
	missile.ForceVec.Y = experienceSpill32TowardZero(objectForceDiv53_52DF80(objectForceMul53_52DF80(dy, objectForceMemory32_52DF80(missile.SpeedCur)), distance))
}

func (s *Server) AntiSpellTargetState53BD10() *AntiSpellTargetState53BD10 {
	return &s.antiSpellTarget53BD10
}
