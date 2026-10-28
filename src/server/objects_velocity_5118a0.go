package server

import (
	"math"

	"github.com/opennox/libs/types"
)

// ObjectsVelocityRuntime5118A0 retains the original two enumeration passes
// and the live object/point arguments at each external service boundary.
// Cache is the original ten DWORDs, adjacent to the script object cache.
type ObjectsVelocityRuntime5118A0 struct {
	Cache     *[10]uint32
	Lookup    func(string) int32
	First     func() *Object
	Next      func(*Object) *Object
	Prepare   func(*Object)
	Springs   func()
	Action    func(*Object, int32) int32
	Buff      func(*Object, int32) int32
	Trace     func(*types.Rectf, byte) int32
	Tile      func(*types.Pointf) int32
	Collision func(*Object, uintptr, *types.Pointf)
	Sync      func(*Object)
	Coords    func(*Object)
	Move      func(*Object)
}

// UpdateObjectsVelocity5118A0 preserves GAME.EXE's PC53/ToZero integration
// without changing the host floating-point environment. In particular, X's
// target/delta and synchronization difference stay wide where Y is spilled.
func UpdateObjectsVelocity5118A0(step float32, r ObjectsVelocityRuntime5118A0) int32 {
	if r.Cache[0] == 0 {
		// The repeated small flame names are also repeated in GAME.EXE.
		for i, name := range [...]string{
			"SmallFlameCleanse", "SmallFlameCleanse", "MediumFlameCleanse", "FlameCleanse", "LargeFlameCleanse",
			"SmallBlueFlameCleanse", "SmallBlueFlameCleanse", "MediumBlueFlameCleanse", "BlueFlameCleanse", "LargeBlueFlameCleanse",
		} {
			r.Cache[i] = uint32(r.Lookup(name))
		}
	}
	for obj := r.First(); obj != nil; obj = r.Next(obj) {
		r.Prepare(obj)
	}
	r.Springs()
	for obj := r.First(); obj != nil; obj = r.Next(obj) {
		if obj.ObjFlags&2 != 0 || obj.ObjClass&2 != 0 && r.Action(obj, 67) != 0 {
			obj.Pos24.Y = 0
			obj.Pos24.X = 0
			obj.VelVec.Y = 0
			obj.VelVec.X = 0
			continue
		}
		var targetX float64
		var targetY float32
		if r.Buff(obj, 5) != 0 || r.Buff(obj, 25) != 0 || r.Buff(obj, 28) != 0 {
			targetY = obj.Pos24.Y // MOV preserves the raw word, including signaling NaNs.
			targetX = float64(obj.Pos24.X)
		} else {
			targetX = objectForceAdd53_52DF80(float64(obj.Pos24.X), objectForceMemory32_52DF80(obj.ForceVec.X))
			targetY = experienceSpill32TowardZero(objectForceAdd53_52DF80(float64(obj.Pos24.Y), objectForceMemory32_52DF80(obj.ForceVec.Y)))
		}
		dragX := objectForceMul53_52DF80(float64(obj.Float28), objectForceMemory32_52DF80(obj.VelVec.X))
		oldNewX := obj.NewPos.X
		deltaX := objectForceSub53_52DF80(targetX, dragX)
		// These FLDS/FMUL operands are reversed for Y in the original, which
		// matters when a signaling and a quiet NaN have different payloads.
		dragY := objectForceMul53_52DF80(float64(obj.VelVec.Y), objectForceMemory32_52DF80(obj.Float28))
		deltaY := experienceSpill32TowardZero(objectForceSub53_52DF80(objectForceMemory32_52DF80(targetY), dragY))
		obj.VelVec.X = experienceSpill32TowardZero(objectForceAdd53_52DF80(
			objectForceMul53_52DF80(deltaX, objectForceMemory32_52DF80(step)), objectForceMemory32_52DF80(obj.VelVec.X)))
		obj.VelVec.Y = experienceSpill32TowardZero(objectForceAdd53_52DF80(
			objectForceMul53_52DF80(float64(deltaY), objectForceMemory32_52DF80(step)), objectForceMemory32_52DF80(obj.VelVec.Y)))
		ray := types.Rectf{Min: types.Ptf(oldNewX, obj.NewPos.Y)}
		ray.Max.X = experienceSpill32TowardZero(objectForceAdd53_52DF80(
			objectForceMul53_52DF80(float64(step), objectForceMemory32_52DF80(obj.VelVec.X)), objectForceMemory32_52DF80(obj.NewPos.X)))
		ray.Max.Y = experienceSpill32TowardZero(objectForceAdd53_52DF80(
			objectForceMul53_52DF80(float64(step), objectForceMemory32_52DF80(obj.VelVec.Y)), objectForceMemory32_52DF80(obj.NewPos.Y)))
		flags := byte((uint32(obj.ObjFlags)>>12)&4 | 1)
		typ := uint32(obj.TypeInd)
		for _, cleanse := range r.Cache {
			if typ == cleanse {
				flags |= 0x40
				break
			}
		}
		if r.Trace(&ray, flags) != 0 {
			obj.NewPos.X = ray.Max.X
			obj.NewPos.Y = ray.Max.Y
		}
		if obj.ObjFlags&0x4000 == 0 && obj.HealthData != nil && r.Tile(&obj.NewPos) == 6 {
			normal := types.Pointf{}
			r.Collision(obj, 6, &normal)
		}
		dx := math.Abs(objectForceSub53_52DF80(float64(obj.NewPos.X), objectForceMemory32_52DF80(obj.PosVec.X)))
		dy := experienceSpill32TowardZero(math.Abs(objectForceSub53_52DF80(float64(obj.NewPos.Y), objectForceMemory32_52DF80(obj.PosVec.Y))))
		if dx > float64(float32(0.01)) || dy > float32(0.01) {
			r.Sync(obj)
			r.Coords(obj)
			r.Move(obj)
		}
	}
	return 0
}
