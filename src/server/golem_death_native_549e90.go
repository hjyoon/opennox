package server

import (
	"github.com/opennox/libs/object"
	"github.com/opennox/libs/types"
	"math"
)

// GolemDeathNative549E90 implements the two raw public C callbacks. Ordinary
// dispatcher admission, normalized indices and its adapters remain separate.
// The source-relative origin survives callback allocation/stack growth without
// placing the original nil-base position offset in a Go pointer slot.
func (unit *Object) GolemDeathNative549E90(mech bool, state *GolemDeathState549E90, rt GolemDeathRuntime549E90) int32 {
	sound := uint32(487)
	if mech {
		sound = 494
	}
	rt.Audio(sound, unit, 0, 0)
	origin := skeletonOrigin54A310(unit)
	rt.PointFX(0x8a, origin)
	spill := func(value float64) float32 {
		if rt.RoundToNearest {
			return float32(value)
		}
		return experienceSpill32TowardZero(value)
	}
	var pos types.Pointf
	if mech {
		// Both the sentinel read and the subsequent name load remain live; a
		// constructor/effect can change a later table slot before the next read.
		for slot := int32(0); rt.MechName(slot) != nil; slot++ {
			part := rt.New(rt.MechName(slot))
			if part == nil {
				break
			}
			rt.Point(30, origin, &pos)
			rt.Create(part, nil, pos, 0)
			rt.Raise(part, spill(rt.RandomFloat(10, 70)))
			vertical := rt.RandomFloat(-2, 0)
			flags := part.ObjFlags
			part.Field27 = spill(vertical)
			part.Field29 = math.Float32bits(2)
			part.ObjFlags = flags | object.FlagBouncy
			seconds := rt.RandomInt(10, 20, GolemDecaySeconds549E90)
			rt.Decay(part, rt.FPS()*uint32(seconds))
		}
		return 1
	}
	count := int32(6)
	if rt.Flags() != 0 {
		count = rt.RandomInt(20, 30, GolemPartCount549E90)
	}
	if count <= 0 {
		return 1
	}
	index := int32(*state.Index)
	for i := int32(0); i < count; i++ {
		part := rt.New(rt.RockName(index))
		if part == nil {
			break
		}
		rt.Point(30, origin, &pos)
		rt.Create(part, nil, pos, 0)
		rt.Raise(part, spill(rt.RandomFloat(10, 70)))
		part.Field27 = spill(rt.RandomFloat(-2, 0))
		mass := rt.Mass(int32(*state.Index))
		flags := part.ObjFlags
		part.ObjFlags = flags | object.FlagBouncy
		part.Field29 = math.Float32bits(float32(mass))
		rt.Force(origin, part, spill(rt.RandomFloat(5, 20)))
		lo, hi := int32(5), int32(10)
		if rt.Flags() != 0 {
			lo, hi = 10, 20
		}
		seconds := rt.RandomInt(lo, hi, GolemDecaySeconds549E90)
		rt.Decay(part, rt.FPS()*uint32(seconds))
		// Original INC wraps the DWORD before signed IDIV. Go's MinInt/-1
		// remainder is zero, so retain the original CPU divide-overflow fault.
		next, divisor := int32(*state.Index+1), int32(*state.Divisor)
		if next == math.MinInt32 && divisor == -1 {
			panic("Golem rock modulo divide overflow")
		}
		index = next % divisor
		*state.Index = uint32(index)
	}
	return 1
}
