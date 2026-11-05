package server

import (
	"math"
	"unsafe"

	"github.com/opennox/libs/object"
	"github.com/opennox/libs/types"
)

// SpecialDeathNative54E460 implements the three raw C entries. Ordinary
// callback admission/normalization remains in the ordinary dispatcher.
func (unit *Object) SpecialDeathNative54E460(entry int, state *GolemDeathState549E90, rt SpecialDeathRuntime54E460) {
	switch entry {
	case 0:
		parent := rt.Parent(unit)
		if parent != nil {
			specialDeathMarkerSlot54E460(specialDeathParentUpdate54E460(parent), unit)
		}
		rt.PointFX(0x8a, skeletonOrigin54A310(unit))
		rt.Delete(unit)
	case 1:
		rt.Audio(757, unit, 0, 0)
		origin := skeletonOrigin54A310(unit)
		rt.PointFX(0x8a, origin)
		count := rt.RandomInt(20, 30, GolemPartCount549E90)
		if count <= 0 {
			rt.Delete(unit)
			return
		}
		spill := func(value float64) float32 {
			if rt.RoundToNearest {
				return float32(value)
			}
			return experienceSpill32TowardZero(value)
		}
		index := int32(*state.Index)
		var pos types.Pointf
		for i := int32(0); i < count; i++ {
			part := rt.New(rt.RockName(index))
			if part == nil {
				// The original allocation-failure path returns without deletion.
				return
			}
			rt.Point(30, origin, &pos)
			rt.Create(part, nil, pos, 0)
			rt.Raise(part, spill(rt.RandomFloat(10, 70)))
			specialDeathPartVertical54E460(part, spill(rt.RandomFloat(-2, 0)))
			mass := rt.Mass(int32(*state.Index))
			specialDeathPartFlags54E460(part)
			specialDeathPartMass54E460(part, mass)
			rt.Force(origin, part, spill(rt.RandomFloat(5, 20)))
			seconds := rt.RandomInt(45, 75, GolemDecaySeconds549E90)
			rt.Decay(part, rt.FPS()*uint32(seconds))
			index = specialDeathBoulderModulo54E460(*state.Index+1, *state.Divisor)
			*state.Index = uint32(index)
		}
		rt.Delete(unit)
	case 2:
		data := specialDeathSourceUpdate54E460(unit)
		rt.Timer(rt.Frame())
		rt.Mode(0)
		rt.Script(SpecialDeathCallback54E460{Data: data}, specialDeathSourceOwner54E460(unit), unit, 3)
		rt.Audio(1000, unit, 0, 0)
		rt.GeneratorFX(skeletonOrigin54A310(unit), 200)
		if rt.Quest() != 0 {
			if owner := specialDeathSourceOwner54E460(unit); owner != nil {
				parent := rt.Parent(owner)
				if specialDeathParentClass54E460(parent)&4 != 0 {
					rt.Credit(parent)
				}
			}
		}
		if part := rt.New(rt.DestroyedName); part != nil {
			// Original stack argument evaluation reads Y before X, after New.
			y := specialDeathSourceY54E460(unit)
			x := specialDeathSourceX54E460(unit)
			rt.Create(part, nil, types.Ptf(x, y), 0)
		}
		rt.Delete(unit)
	default:
		panic("unknown raw special death entry")
	}
}

// Fault sites stay independent of compiler inlining and preserve the original
// read/write ordering. No extra class, source or update-data guards are added.
//
//go:noinline
func specialDeathParentUpdate54E460(unit *Object) *PlayerUpdateData {
	return (*PlayerUpdateData)(unit.UpdateData)
}

//go:noinline
func specialDeathMarkerSlot54E460(data *PlayerUpdateData, unit *Object) {
	for i := 0; i < 4; i++ {
		if data.Field29[i] == unit {
			data.Field29[i] = nil
			return
		}
	}
}

//go:noinline
func specialDeathSourceUpdate54E460(unit *Object) unsafe.Pointer { return unit.UpdateData }

//go:noinline
func specialDeathSourceOwner54E460(unit *Object) *Object { return unit.Obj130 }

//go:noinline
func specialDeathParentClass54E460(unit *Object) byte {
	return (*[unsafe.Sizeof(Object{})]byte)(unsafe.Pointer(unit))[unsafe.Offsetof(Object{}.ObjClass)]
}

//go:noinline
func specialDeathSourceY54E460(unit *Object) float32 { return unit.PosVec.Y }

//go:noinline
func specialDeathSourceX54E460(unit *Object) float32 { return unit.PosVec.X }

//go:noinline
func specialDeathPartVertical54E460(unit *Object, value float32) { unit.Field27 = value }

//go:noinline
func specialDeathPartFlags54E460(unit *Object) { unit.ObjFlags |= object.FlagBouncy }

//go:noinline
func specialDeathPartMass54E460(unit *Object, value byte) {
	unit.Field29 = math.Float32bits(float32(value))
}

//go:noinline
func specialDeathBoulderModulo54E460(next, divisor uint32) int32 {
	if int32(next) == math.MinInt32 && int32(divisor) == -1 {
		panic("Boulder modulo divide overflow")
	}
	return int32(next) % int32(divisor)
}
