package server

import (
	"math"

	"github.com/opennox/libs/object"
	"github.com/opennox/libs/types"
)

// IntegrateObjectMotion51B810 advances the object using GAME.EXE's
// precision-53, round-toward-zero arithmetic and refreshes its live bounds.
func IntegrateObjectMotion51B810(obj *Object) {
	if obj == nil {
		return
	}
	add := pointOnLineAddChop53_57C8A0
	mul := pointOnLineMulChop53_57C8A0
	spill := pointOnLineSpill57C8A0

	obj.PosVec = obj.NewPos
	drag := float64(obj.Float28)
	vx := mul(add(float64(obj.ForceVec.X), float64(obj.VelVec.X)), drag)
	vy := mul(add(float64(obj.ForceVec.Y), float64(obj.VelVec.Y)), drag)
	obj.VelVec = types.Ptf(spill(vx), spill(vy))
	// 0051B83D still has the full X velocity in the FPU. Y reloads
	// the binary32 velocity stored on the stack by 0051B836.
	obj.NewPos.X = spill(add(vx, float64(obj.NewPos.X)))
	obj.NewPos.Y = spill(add(float64(obj.VelVec.Y), float64(obj.NewPos.Y)))
	obj.Nox_xxx_objectUnkUpdateCoords_4E7290()
}

type ObjectFallRuntime51B870 struct {
	Raise        func(*Object, float32)
	Move         func(*Object, types.Pointf)
	Enqueue      func(*Object)
	LandingSound func(*Object)
}

// UpdateObjectFall51B870 updates the original fall speed and flags. A raise
// or collision callback can change the object, so later operations use the
// fields at the same points where GAME.EXE reloads them.
func UpdateObjectFall51B870(obj *Object, rt ObjectFallRuntime51B870) {
	if obj == nil {
		return
	}
	add := pointOnLineAddChop53_57C8A0
	mul := pointOnLineMulChop53_57C8A0
	div := pointOnLineDivChop53_57C8A0
	spill := pointOnLineSpill57C8A0

	flags, z := obj.ObjFlags, obj.ZVal
	if flags.Has(object.FlagInHole) {
		rt.Raise(obj, spill(add(float64(z), float64(obj.Field27))))
		vel := spill(add(float64(obj.Field27), -1))
		obj.ForceVec = types.Pointf{}
		obj.VelVec = types.Pointf{}
		obj.Field27 = vel

		dx := add(float64(obj.PosVec.X), -float64(obj.Pos39.X))
		dy := add(float64(obj.PosVec.Y), -float64(obj.Pos39.Y))
		storedY := float64(spill(dy))
		dist := monsterMoveForceSqrtChop53_50D50C(add(mul(dy, storedY), mul(dx, dx)))
		if dist > 0 {
			storedDist := float64(spill(dist))
			obj.ForceVec.X = spill(div(mul(dx, -3), storedDist))
			obj.ForceVec.Y = spill(div(mul(storedY, -3), storedDist))
		}
		// Tests of x87 C0 alone include unordered values.
		if !(obj.ZVal >= -50) {
			rt.Raise(obj, 90)
			obj.ObjFlags &^= object.FlagInHole
			rt.Move(obj, types.Ptf(math.Float32frombits(obj.Field41), math.Float32frombits(obj.Field42)))
		}
		return
	}

	vz := obj.Field27
	// Tests of x87 C3 treat an unordered value like zero on this path.
	if !(z < 0 || z > 0 || vz < 0 || vz > 0) {
		return
	}
	if flags.Has(object.FlagBouncy) {
		rt.Raise(obj, spill(add(float64(z), float64(vz))))
		if obj.ZVal >= 0 {
			obj.Field27 = spill(add(float64(obj.Field27), -.5))
			return
		}
		rt.Raise(obj, 0)
		rebound := mul(mul(-float64(obj.Field27), float64(math.Float32frombits(obj.Field29))), float64(float32(.1)))
		obj.Field27 = spill(rebound)
		if !(rebound >= 2) {
			rt.Raise(obj, 0)
			obj.Field27 = 0
		}
		return
	}
	if flags.Has(object.FlagOnObject) {
		return
	}
	if z > 0 {
		if !(vz > 0) {
			obj.ObjFlags = flags | object.FlagFalling
		}
		rt.Raise(obj, spill(add(float64(z), float64(vz))))
		obj.Field27 = spill(add(float64(obj.Field27), -1))
	}
	if obj.ZVal > 0 {
		return
	}
	vz = obj.Field27
	obj.ObjFlags &^= object.FlagFalling
	if !(vz >= 0) && !obj.ObjClass.Has(object.ClassMissile) {
		rt.Enqueue(obj)
		if !(obj.Field27 >= -10) && obj.ObjClass.Has(object.ClassPlayer) {
			rt.LandingSound(obj)
		}
	}
	rt.Raise(obj, 0)
	obj.Field27 = 0
}
