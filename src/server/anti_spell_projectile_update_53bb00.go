package server

import (
	"math"

	"github.com/opennox/libs/object"
	"github.com/opennox/libs/types"
)

// AntiSpellProjectileRuntime53BB00 supplies the world effects for the
// native-width AntiSpellProjectileUpdate callback.
type AntiSpellProjectileRuntime53BB00 struct {
	Frame         func() uint32
	FPS           func() uint32
	EachMissile   func(types.Pointf, float32, func(*Object) bool)
	CanInteract   func(*Object, *Object) bool
	DelayedDelete func(*Object)
	Audio         func(*Object)
	RandomFloat   func(float32, float32) float32
	RandomInt     func(int, int) int
	CreateSpark   func(types.Pointf, int, int, types.Pointf, float32, *Object)
}

func antiSpellFindTarget53BD10(missile *Object, runtime AntiSpellProjectileRuntime53BB00) *Object {
	var target *Object
	bestDistance := float64(100_000_000)
	runtime.EachMissile(missile.PosVec, 600, func(candidate *Object) bool {
		if candidate == nil || candidate == missile || candidate.ObjFlags.Has(object.FlagDestroyed) ||
			(uint8(candidate.ObjClass)&1 != 0 && uint8(candidate.ObjSubClass)&2 == 0) ||
			(missile.ObjOwner != nil && candidate.HasOwner(missile.ObjOwner)) ||
			!runtime.CanInteract(missile, candidate) {
			return true
		}
		delta := missile.PosVec.Sub(candidate.PosVec)
		distance := float64(delta.X)*float64(delta.X) + float64(delta.Y)*float64(delta.Y)
		if distance < bestDistance {
			bestDistance = distance
			target = candidate
		}
		return true
	})
	return target
}

// AntiSpellProjectileUpdate53BB00 restores GAME.EXE 0053BB00 without reading
// the PE32 pointer in the missile's update record or its target object.
func AntiSpellProjectileUpdate53BB00(missile *Object, runtime AntiSpellProjectileRuntime53BB00) {
	if missile == nil || missile.UpdateData == nil {
		return
	}
	frame := runtime.Frame()
	fps := runtime.FPS()
	if frame-missile.Field32 > 5*fps {
		runtime.DelayedDelete(missile)
		return
	}
	update := missile.UpdateDataMissile()
	if update.Target != nil && update.Target.ObjFlags.Has(object.FlagDestroyed) {
		update.Target = nil
	}
	if update.Target == nil && frame-missile.Field34 > fps>>2 {
		update.Target = antiSpellFindTarget53BD10(missile, runtime)
		missile.Field34 = runtime.Frame()
	}
	if target := update.Target; target != nil {
		delta := target.PosVec.Sub(missile.PosVec)
		distance := math.Hypot(float64(delta.X), float64(delta.Y)) + 0.1
		missile.ForceVec = types.Ptf(
			float32(float64(delta.X)*float64(missile.SpeedCur)/distance),
			float32(float64(delta.Y)*float64(missile.SpeedCur)/distance),
		)
		if distance < 10 {
			runtime.Audio(missile)
			runtime.DelayedDelete(missile)
			runtime.DelayedDelete(target)
		}
	} else {
		velocity := missile.VelVec
		length := math.Hypot(float64(velocity.X), float64(velocity.Y)) + 0.1
		missile.ForceVec = types.Ptf(
			float32(float64(velocity.X)*float64(missile.SpeedCur)/length),
			float32(float64(velocity.Y)*float64(missile.SpeedCur)/length),
		)
	}
	velocityY := runtime.RandomFloat(-2, 2)
	velocityX := runtime.RandomFloat(-2, 2)
	lifetime := runtime.RandomInt(15, 30)
	posY := runtime.RandomFloat(-4, 4) + missile.PosVec.Y
	posX := runtime.RandomFloat(-4, 4) + missile.PosVec.X
	runtime.CreateSpark(types.Ptf(posX, posY), 3, lifetime, types.Ptf(velocityX, velocityY), 0, missile)
}
