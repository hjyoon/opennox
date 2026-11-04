package server

import (
	"github.com/opennox/libs/object"
	"github.com/opennox/libs/types"
	"math"
	"unsafe"
)

// The raw C entries can form a nil-base position address and pass it to a
// service that does not dereference it, then return after allocation failure.
// Keep this one address representation separate from checked native object
// arithmetic; the consuming world service retains the original memory fault.
//
//go:noinline
//go:nocheckptr
func skeletonNilOrigin54A310() *types.Pointf {
	return (*types.Pointf)(unsafe.Pointer(unsafe.Offsetof(Object{}.PosVec)))
}

func skeletonOrigin54A310(unit *Object) *types.Pointf {
	if unit == nil {
		return skeletonNilOrigin54A310()
	}
	return (*types.Pointf)(unsafe.Add(unsafe.Pointer(unit), unsafe.Offsetof(Object{}.PosVec)))
}

// SkeletonDeathNative54A310 preserves the raw public C callbacks. Ordinary
// monster admission and its guarded Go adapters remain separate.
func (unit *Object) SkeletonDeathNative54A310(lord bool, state *SkeletonDeathState54A310, rt SkeletonDeathRuntime54A310) int32 {
	roll := rt.RandomInt(0, 100, SkeletonRoll54A310)
	unit.SkeletonDebrisNative54A4C0(state, rt)
	if roll > 20 {
		name := rt.Names.WoodenShield
		mods := [4]unsafe.Pointer{nil, rt.Names.Material, nil, nil}
		if roll <= 50 {
			name = rt.Names.Sword
			mods[0] = rt.Names.Power
		} else if lord {
			name = rt.Names.SteelShield
		}
		unit.SkeletonDropNative54A390(name, mods, 0, rt)
	}
	return 1
}

// SkeletonDropNative54A390 does not read source/name/use pointers before the
// original flags/allocation/class gates. Class and use data stay live after
// modifier services, while the four incoming name pointer values are cached.
func (unit *Object) SkeletonDropNative54A390(name unsafe.Pointer, mods [4]unsafe.Pointer, charge int32, rt SkeletonDeathRuntime54A310) {
	if rt.Flags() == 0 {
		return
	}
	item := rt.New(name)
	if item == nil {
		return
	}
	origin := skeletonOrigin54A310(unit)
	var pos types.Pointf
	rt.Point(50, origin, &pos)
	rt.Create(item, nil, pos, 0)
	if uint32(item.ObjClass)&0x13001000 != 0 && (mods[0] != nil || mods[1] != nil || mods[2] != nil || mods[3] != nil) {
		attrs := ModifierInitData{} // the original fifth stack DWORD was undefined
		for i, p := range mods {
			attrs.Modifiers[i] = (*ModifierEff)(rt.ModifierDesc(rt.ModifierID(p)))
		}
		rt.Apply(item, &attrs)
	}
	if charge != 0 && uint32(item.ObjClass)&0x1000000 != 0 && uint32(item.ObjSubClass)&0x82 != 0 {
		(*[2]byte)(item.UseData.Ptr)[1] = byte(charge)
	}
	class := uint32(item.ObjClass)
	if class&0x2000000 != 0 {
		rt.ArmorSound(item)
	} else if class&0x1001000 != 0 {
		rt.WeaponSound(item)
	}
}

// SkeletonDebrisNative54A4C0 keeps one live origin pointer and one output
// point throughout creation. Decay precedes the part-count RNG draw. The
// initial part index is loaded only for a positive count; successful parts
// advance the live signed-DWORD index using the live signed divisor.
func (unit *Object) SkeletonDebrisNative54A4C0(state *SkeletonDeathState54A310, rt SkeletonDeathRuntime54A310) int32 {
	rt.Audio(356, unit, 0, 0)
	origin := skeletonOrigin54A310(unit)
	rt.PointFX(0x8a, origin)
	skull := rt.New(rt.Names.Skull)
	if skull == nil {
		return 1
	}
	spill := func(v float64) float32 {
		if rt.RoundToNearest {
			return float32(v)
		}
		return experienceSpill32TowardZero(v)
	}
	var pos types.Pointf
	rt.Point(20, origin, &pos)
	rt.Create(skull, nil, pos, 0)
	rt.Raise(skull, 40)
	vertical := rt.RandomFloat(-2, 0)
	flags := skull.ObjFlags
	skull.Field27 = spill(vertical)
	skull.Field29 = math.Float32bits(4)
	skull.ObjFlags = flags | object.FlagBouncy
	rt.Force(origin, skull, spill(rt.RandomFloat(5, 25)))
	lo, hi := int32(2), int32(5)
	if rt.Flags() != 0 {
		lo, hi = 10, 20
	}
	seconds := rt.RandomInt(lo, hi, SkeletonSkullSeconds54A310)
	rt.Decay(skull, rt.FPS()*uint32(seconds))
	countLo, countHi := int32(5), int32(10)
	if lo == 10 {
		countLo, countHi = 10, 20
	}
	count := rt.RandomInt(countLo, countHi, SkeletonPartCount54A310)
	if count <= 0 {
		return 1
	}
	index := int32(*state.Index)
	for i := int32(0); i < count; i++ {
		part := rt.New(rt.BoneName(index))
		if part == nil {
			break
		}
		rt.Point(20, origin, &pos)
		rt.Create(part, nil, pos, 0)
		rt.Raise(part, spill(rt.RandomFloat(10, 35)))
		vertical = rt.RandomFloat(-2, 0)
		flags = part.ObjFlags
		part.Field27 = spill(vertical)
		part.Field29 = math.Float32bits(4)
		part.ObjFlags = flags | object.FlagBouncy
		rt.Force(origin, part, spill(rt.RandomFloat(5, 25)))
		lo, hi = 2, 5
		if rt.Flags() != 0 {
			lo, hi = 10, 20
		}
		seconds = rt.RandomInt(lo, hi, SkeletonBoneSeconds54A310)
		rt.Decay(part, rt.FPS()*uint32(seconds))
		next, divisor := int32(*state.Index+1), int32(*state.Divisor)
		if next == math.MinInt32 && divisor == -1 {
			panic("Skeleton bone modulo divide overflow")
		}
		index = next % divisor
		*state.Index = uint32(index)
	}
	return 1
}
