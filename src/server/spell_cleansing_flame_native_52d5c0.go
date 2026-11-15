package server

import (
	"github.com/opennox/libs/object"
	"github.com/opennox/libs/types"
	"github.com/opennox/opennox/v1/common/sound"
	"unsafe"
)

// CleansingFlameNativeRuntime52D5C0 supplies the original external services.
// Cache, timing and direction reads remain live, in their original order.
type CleansingFlameNativeRuntime52D5C0 struct {
	LoadCache       func(int32) uint32
	StoreCache      func(int32, uint32)
	LookupType      func(string) uint32
	PriorityMsg     func(*Object, string, byte)
	HasGameFlag     func(uint32) bool
	Random          func(int32, int32) int32
	NewObject       func(uint32) *Object
	DirectionVector func(int16) types.Pointf
	TraceRay        func(types.Pointf, types.Pointf, MapTraceFlags) bool
	CreateAt        func(*Object, *Object, types.Pointf)
	DelayedDelete   func(*Object)
	Fps             func() uint32
	Frame           func() uint32
	UpdateCallback  unsafe.Pointer
	AddUpdatable    func(*Object)
	PredictLinear   func(*Object)
	CastSound       func(int32) sound.ID
	Audio           func(sound.ID, *Object, int, uint32)
}

// CleansingFlameNative52D5C0 restores the original 732-byte cast entry using
// native owned-list links and a full-width update callback. Each cast owns its
// local positions/vectors; DWORD cache and timer arithmetic retain wrapping.
func CleansingFlameNative52D5C0(id int32, second, recipient, caster *Object, level int32, rt CleansingFlameNativeRuntime52D5C0) int32 {
	if rt.LoadCache(0) == 0 {
		for slot, name := range [10]string{
			"SmallFlameCleanse", "SmallFlameCleanse", "MediumFlameCleanse", "FlameCleanse", "LargeFlameCleanse",
			"SmallBlueFlameCleanse", "SmallBlueFlameCleanse", "MediumBlueFlameCleanse", "BlueFlameCleanse", "LargeBlueFlameCleanse",
		} {
			rt.StoreCache(int32(slot), rt.LookupType(name))
		}
	}
	if second != nil {
		for owned := second.Field129; owned != nil; owned = owned.Field128 {
			kind := uint32(owned.TypeInd)
			for slot := int32(0); slot < 10; slot++ {
				if kind == rt.LoadCache(slot) {
					rt.PriorityMsg(recipient, "plyrspel.c:TooManySpells", 0)
					return 0
				}
			}
		}
	}
	base := int32(5)
	if id == 10 {
		base = 0
	}
	if !rt.HasGameFlag(2048) {
		level = 4
	}
	for remaining := 48; remaining != 0; remaining-- {
		power := level - rt.Random(0, 1)
		if power < 1 {
			continue
		}
		flame := rt.NewObject(rt.LoadCache(base + power - 1))
		if flame == nil {
			continue
		}
		direction := int16(rt.Random(0, 255))
		radius := float64(flame.Shape.Circle.R)
		flame.Direction1 = Dir16(direction)
		fromX := caster.PosVec.X
		distance := monsterMoveToRunAddChop53_544434(radius, float64(caster.Shape.Circle.R))
		distance = monsterMoveToRunAddChop53_544434(distance, 4)
		vector := rt.DirectionVector(direction)
		// Keep the original separate FMUL, FADD and binary32 ToZero spill;
		// ARM64 must not contract this sequence into a fused multiply-add.
		x := monsterMoveForceMulChop53_50D4FE(distance, float64(vector.X))
		fromY := caster.PosVec.Y
		x = monsterMoveToRunAddChop53_544434(x, float64(caster.PosVec.X))
		y := monsterMoveForceMulChop53_50D4FE(distance, float64(vector.Y))
		y = monsterMoveToRunAddChop53_544434(y, float64(caster.PosVec.Y))
		position := types.Ptf(float32(monsterMoveToRunSpill544440(x)), float32(monsterMoveToRunSpill544440(y)))
		if !rt.TraceRay(types.Ptf(fromX, fromY), position, MapTraceFlags(65)) {
			rt.DelayedDelete(flame)
			continue
		}
		rt.CreateAt(flame, caster, position)
		flame.Direction2 = flame.Direction1
		flame.VelVec.X = float32(monsterMoveToRunSpill544440(float64(vector.X) * 4))
		flame.VelVec.Y = float32(monsterMoveToRunSpill544440(float64(vector.Y) * 4))
		fps := rt.Fps()
		duration := rt.Random(int32(3*fps), int32(6*fps))
		flame.Field34 = rt.Frame() + uint32(duration)
		flame.Pos39 = caster.PosVec
		flame.Update = rt.UpdateCallback
		rt.AddUpdatable(flame)
		class := flame.ObjClass | object.ClassClientPredict
		flame.Float28 = 0
		flame.ObjClass = class
		rt.PredictLinear(flame)
	}
	castSound := rt.CastSound(id)
	rt.Audio(castSound, caster, 0, 0)
	return 1
}
