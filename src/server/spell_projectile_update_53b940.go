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
	update := missile.UpdateDataSpellProjectile()
	lifetimeKey := "UnTargetedSpellLifetime"
	if update.Target != nil {
		lifetimeKey = "TargetedSpellLifetime"
	}
	lifetime := float32(runtime.Lifetime(lifetimeKey))
	// nox_float2int uses x87 FISTP with round-to-nearest-even. An invalid
	// conversion yields INT32_MIN, which the original comparison casts to u32.
	var lifetimeFrames int32
	if math.IsNaN(float64(lifetime)) || lifetime >= 2147483648 || lifetime < -2147483648 {
		lifetimeFrames = math.MinInt32
	} else {
		lifetimeFrames = int32(math.RoundToEven(float64(lifetime)))
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

	var delta types.Pointf
	if update.Target != nil {
		delta = update.Target.PosVec.Sub(missile.PosVec)
	} else {
		delta = missile.VelVec
	}
	distance := math.Hypot(float64(delta.X), float64(delta.Y)) + 0.10000000149011612
	missile.Float28 = math.Float32frombits(1063675494)
	missile.ForceVec = types.Ptf(
		float32(float64(delta.X)*float64(missile.SpeedCur)/distance),
		float32(float64(delta.Y)*float64(missile.SpeedCur)/distance),
	)
}
