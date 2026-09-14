package server

import (
	"math"

	"github.com/opennox/libs/object"
)

// MagicMissileUpdateRuntime53BDA0 supplies the world operations used by the
// thing.bin update callback. The missile, owner, and target stay native-width.
type MagicMissileUpdateRuntime53BDA0 struct {
	Frame        func() uint32
	FPS          func() uint32
	SearchTarget func(missile, owner *Object) *Object
	Collide      func(missile *Object)
}

// MagicMissileUpdate53BDA0 follows GAME.EXE 0053BDA0 without reading PE32
// object or update-data pointer offsets from native-width records.
func MagicMissileUpdate53BDA0(missile *Object, runtime MagicMissileUpdateRuntime53BDA0) {
	if missile == nil || missile.UpdateData == nil {
		return
	}
	update := missile.UpdateDataMissile()
	owner := update.Owner
	if owner == nil || owner.ObjFlags.Has(object.FlagDestroyed) {
		runtime.Collide(missile)
		return
	}
	frame := runtime.Frame()
	if frame-missile.Field32 > 3*runtime.FPS() {
		runtime.Collide(missile)
		return
	}

	if target := update.Target; target != nil && target.ObjFlags.HasAny(object.FlagDestroyed|object.FlagDead) {
		update.Target = nil
	}
	if update.Target == nil && frame&7 == 0 {
		update.Target = runtime.SearchTarget(missile, owner)
		if update.Target == missile.ObjOwner {
			update.Target = nil
		}
	}
	if target := update.Target; target != nil {
		cos, sin := SinCosDir(byte(missile.Direction2))
		delta := target.PosVec.Sub(missile.PosVec)
		if float64(delta.Y)*float64(cos)-float64(delta.X)*float64(sin) >= 0 {
			missile.Direction2 = RoundDir(int(missile.Direction2) + 42)
		} else {
			missile.Direction2 = RoundDir(int(missile.Direction2) - 42)
		}
	}
	force := missile.Direction2.Vec().Mul(missile.SpeedCur)
	missile.ForceVec.X = force.X
	missile.Float28 = math.Float32frombits(1061997773)
	missile.ForceVec.Y = force.Y
}
