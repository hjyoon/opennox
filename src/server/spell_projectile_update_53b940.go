package server

import (
	"math"

	"github.com/opennox/libs/object"
	"github.com/opennox/libs/types"
)

// SpellProjectileUpdateRuntime53B940 supplies the game state and effects for
// the native-width SpellProjectileUpdate callback.
type SpellProjectileUpdateRuntime53B940 struct {
	Frame        func() uint32
	TickRate     func() uint32
	Lifetime     func(string) float64
	SearchTarget func(missile, owner *Object, spellID uint32) *Object
	SendPointFX  func(types.Pointf)
	Expire       func(*Object)
}

// SpellProjectileUpdate53B940 follows GAME.EXE 0053B940 without interpreting
// the missile or its update record as PE32 pointer-sized words.
func SpellProjectileUpdate53B940(missile *Object, runtime SpellProjectileUpdateRuntime53B940) {
	if missile == nil || missile.UpdateData == nil {
		return
	}
	spellProjectileUpdate53B940(missile, missile.UpdateDataSpellProjectile(), runtime)
}

// SpellProjectileUpdateRaw53B940 retains the original C entry's required
// object/update pointers. The registered Go dispatcher keeps its nil guard.
func SpellProjectileUpdateRaw53B940(missile *Object, runtime SpellProjectileUpdateRuntime53B940) {
	spellProjectileUpdate53B940(missile, (*SpellProjectileUpdateData)(missile.UpdateData), runtime)
}

func spellProjectileUpdate53B940(missile *Object, update *SpellProjectileUpdateData, runtime SpellProjectileUpdateRuntime53B940) {
	lifetimeKey := "UnTargetedSpellLifetime"
	if update.Target != nil {
		lifetimeKey = "TargetedSpellLifetime"
	}
	// GAME.EXE spills the balance double to binary32, then FISTP uses the
	// gameplay PC53/ToZero control word. Do not alter the host FP environment.
	lifetime := experienceSpill32TowardZero(runtime.Lifetime(lifetimeKey))
	var lifetimeFrames int32
	if math.IsNaN(float64(lifetime)) || lifetime >= 2147483648 || lifetime < -2147483648 {
		lifetimeFrames = math.MinInt32
	} else {
		lifetimeFrames = int32(lifetime)
	}
	frame := runtime.Frame()
	if frame-missile.Field32 > uint32(lifetimeFrames) {
		runtime.SendPointFX(missile.PosVec)
		runtime.Expire(missile)
		return
	}

	if update.Field0 != nil && update.Field0.ObjFlags.Has(object.FlagDestroyed) {
		update.Field0 = nil
	}
	if update.Field8 != nil && update.Field8.ObjFlags.Has(object.FlagDestroyed) {
		update.Field8 = nil
	}
	if update.Target != nil && update.Target.ObjFlags.HasAny(object.FlagDestroyed|object.FlagDead) {
		update.Target = nil
	}
	if update.Target == nil && (frame-missile.Field34 > runtime.TickRate()>>2 || missile.Field34 == missile.Field32) {
		update.Target = runtime.SearchTarget(missile, update.Field0, update.Spell12)
		missile.Field34 = runtime.Frame()
	}

	var dx, dy float32
	var square float64
	if update.Target != nil {
		dx = experienceSpill32TowardZero(objectForceSub53_52DF80(float64(update.Target.PosVec.X), objectForceMemory32_52DF80(missile.PosVec.X)))
		dyWide := objectForceSub53_52DF80(float64(update.Target.PosVec.Y), objectForceMemory32_52DF80(missile.PosVec.Y))
		dy = experienceSpill32TowardZero(dyWide)
		// FST at 0053BA52 retains the unspilled Y difference for FMUL;
		// X and the second Y operand have already been spilled to binary32.
		square = objectForceAdd53_52DF80(objectForceMul53_52DF80(dyWide, float64(dy)), objectForceMul53_52DF80(float64(dx), float64(dx)))
	} else {
		dx, dy = missile.VelVec.X, missile.VelVec.Y
		square = objectForceAdd53_52DF80(objectForceMul53_52DF80(float64(dy), float64(dy)), objectForceMul53_52DF80(float64(dx), float64(dx)))
	}
	// The qword at 00583C10 is widened binary32 0.1. The complete sum is
	// passed through the actual double-to-float helper before both divisions.
	distance := experienceSpill32TowardZero(objectForceAdd53_52DF80(monsterMoveForceSqrtChop53_50D50C(square), 0.10000000149011612))
	missile.Float28 = math.Float32frombits(1063675494)
	missile.ForceVec.X = experienceSpill32TowardZero(objectForceDiv53_52DF80(objectForceMul53_52DF80(float64(dx), objectForceMemory32_52DF80(missile.SpeedCur)), float64(distance)))
	missile.ForceVec.Y = experienceSpill32TowardZero(objectForceDiv53_52DF80(objectForceMul53_52DF80(float64(dy), objectForceMemory32_52DF80(missile.SpeedCur)), float64(distance)))
}
