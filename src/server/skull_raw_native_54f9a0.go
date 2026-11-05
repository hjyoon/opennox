package server

import (
	"math"
	"unsafe"

	"github.com/opennox/libs/types"
)

//go:noinline
func skullRawFlags54F9A0(o *Object) uint32 { return uint32(o.ObjFlags) }

//go:noinline
func skullRawUpdatePointer54F9A0(o *Object) unsafe.Pointer { return o.UpdateData }

//go:noinline
func skullRawRadius54FA80(o *Object) float32 { return o.Shape.Circle.R }

//go:noinline
func skullRawDirection54FA80(o *Object) int16 { return int16(o.Direction1) }

//go:noinline
func skullRawX54FBF0(o *Object) float32 { return o.PosVec.X }

//go:noinline
func skullRawY54FBF0(o *Object) float32 { return o.PosVec.Y }

//go:noinline
func skullRawClassByte54FC50(o *Object) uint8 { return *(*uint8)(unsafe.Pointer(&o.ObjClass)) }

func skullRawAdd54F9A0(a, b float64, nearest bool) float64 {
	if nearest {
		return float64(a + b)
	}
	return experienceAdd53TowardZero(a, b)
}
func skullRawMul54F9A0(a, b float64, nearest bool) float64 {
	if nearest {
		// The original PC53 multiply rounds before the following addition.
		// An explicit conversion prevents a fused native multiply-add.
		return float64(a * b)
	}
	return experienceMul53TowardZero(a, b)
}
func skullRawSpill54F9A0(value float64, nearest bool) float32 {
	if nearest {
		return float32(value)
	}
	return experienceSpill32TowardZero(value)
}
func skullRawFISTP54F9A0(value float32, nearest bool) int32 {
	if !nearest {
		return netFXTruncateSignedDWORD419A70(value)
	}
	v := math.RoundToEven(float64(value))
	if math.IsNaN(v) || v < -2147483648 || v >= 2147483648 {
		return math.MinInt32
	}
	return int32(v)
}

// SkullRawUpdate54F9A0 preserves the unguarded public C contract, including
// native cached data pointers and live numeric PE cache cells across reentry.
// Ordinary Go dispatcher admission remains in SkullUpdate54F9A0.
func (source *Object) SkullRawUpdate54F9A0(state *SkullRawState54F9A0, rt SkullRawRuntime54F9A0) uint32 {
	flags := skullRawFlags54F9A0(source)
	data := (*SkullUpdateData)(skullRawUpdatePointer54F9A0(source))
	if flags&0x01000000 == 0 {
		data.Enabled = 0
		return flags
	}
	if data.Enabled == 0 {
		data.ScanDelay = 0
		data.FireDelay = 0
	}
	scan := data.ScanDelay
	data.Enabled = 1
	if scan == 0 {
		source.SkullRawState54FBB0(state, rt)
		data.ScanDelay = rt.FPS()
	}
	if data.TargetReady == 1 && data.FireDelay == 0 {
		if state.UpdateTrap1 == 0 {
			state.UpdateTrap1 = rt.Resolve("ArrowTrap1")
			state.UpdateTrap2 = rt.Resolve("ArrowTrap2")
		}
		source.SkullRawProjectile54FA80(data.ProjectileType, state, rt)
		first := state.UpdateTrap1
		typ := uint32(source.TypeInd)
		if typ == first {
			rt.TrapFX(&source.PosVec, 1)
		} else if typ == state.UpdateTrap2 {
			rt.TrapFX(&source.PosVec, 2)
		}
		data.FireDelay = 30
	}
	if data.ScanDelay != 0 {
		data.ScanDelay--
	}
	value := data.FireDelay
	if value != 0 {
		value--
		data.FireDelay = value
	}
	return value
}

func (source *Object) SkullRawProjectile54FA80(id uint32, state *SkullRawState54F9A0, rt SkullRawRuntime54F9A0) {
	nearest := rt.RoundToNearest
	radius := skullRawRadius54FA80(source)
	direction := skullRawDirection54FA80(source)
	distance := skullRawAdd54F9A0(float64(radius), 4, nearest)
	xcoef := rt.DirectionComponent(direction, 0)
	xpos := skullRawX54FBF0(source)
	x := skullRawSpill54F9A0(skullRawAdd54F9A0(skullRawMul54F9A0(distance, float64(xcoef), nearest), float64(xpos), nearest), nearest)
	ycoef := rt.DirectionComponent(direction, 1)
	ypos := skullRawY54FBF0(source)
	y := skullRawSpill54F9A0(skullRawAdd54F9A0(skullRawMul54F9A0(distance, float64(ycoef), nearest), float64(ypos), nearest), nearest)
	projectile := rt.New(id)
	if projectile == nil {
		return
	}
	rt.Create(projectile, source, types.Ptf(x, y), 0)
	dir := Dir16(uint16(skullRawDirection54FA80(source)))
	projectile.Direction1 = dir
	projectile.Direction2 = dir
	xcoef = rt.DirectionComponent(skullRawDirection54FA80(source), 0)
	projectile.VelVec.X = skullRawSpill54F9A0(skullRawMul54F9A0(float64(xcoef), float64(projectile.SpeedCur), nearest), nearest)
	ycoef = rt.DirectionComponent(skullRawDirection54FA80(source), 1)
	projectile.VelVec.Y = skullRawSpill54F9A0(skullRawMul54F9A0(float64(ycoef), float64(projectile.SpeedCur), nearest), nearest)
	if state.MercArcherArrow == 0 {
		state.MercArcherArrow = rt.Resolve("MercArcherArrow")
		state.ProjectileTrap1 = rt.Resolve("ArrowTrap1")
		state.ProjectileTrap2 = rt.Resolve("ArrowTrap2")
	}
	first := state.ProjectileTrap1
	typ := uint32(source.TypeInd)
	if typ == first || typ == state.ProjectileTrap2 {
		data := (*ProjectileCollideData)(projectile.CollideData)
		value := skullRawSpill54F9A0(rt.Balance(), nearest)
		damage := skullRawFISTP54F9A0(value, nearest)
		data.Damage = damage
		data.Field4 = damage
	}
	if id == state.MercArcherArrow {
		rt.Audio(889, source)
	}
}

func (source *Object) SkullRawState54FBB0(state *SkullRawState54F9A0, rt SkullRawRuntime54F9A0) {
	data := (*SkullUpdateData)(skullRawUpdatePointer54F9A0(source))
	if source.SkullRawScan54FBF0(state, rt) != 0 {
		if data.TargetReady != 1 {
			data.TargetReady = 1
			data.FireDelay = 0
		}
	} else {
		data.TargetReady = 0
	}
}

func (source *Object) SkullRawScan54FBF0(state *SkullRawState54F9A0, rt SkullRawRuntime54F9A0) uint32 {
	nearest := rt.RoundToNearest
	x1 := skullRawSpill54F9A0(skullRawAdd54F9A0(float64(skullRawX54FBF0(source)), -350, nearest), nearest)
	y1 := skullRawSpill54F9A0(skullRawAdd54F9A0(float64(skullRawY54FBF0(source)), -350, nearest), nearest)
	x2 := skullRawSpill54F9A0(skullRawAdd54F9A0(float64(skullRawX54FBF0(source)), 350, nearest), nearest)
	yExtended := skullRawAdd54F9A0(float64(skullRawY54FBF0(source)), 350, nearest)
	state.Reachable = 0
	y2 := skullRawSpill54F9A0(yExtended, nearest)
	rt.Enumerate(types.Rectf{Min: types.Ptf(x1, y1), Max: types.Ptf(x2, y2)}, source, func(candidate, owner *Object) { owner.SkullRawReachable54FC50(candidate, state, rt) })
	return state.Reachable
}

func (source *Object) SkullRawReachable54FC50(candidate *Object, state *SkullRawState54F9A0, rt SkullRawRuntime54F9A0) {
	if skullRawClassByte54FC50(candidate)&6 == 0 {
		return
	}
	if skullRawFlags54F9A0(candidate)&0x8020 != 0 {
		return
	}
	if rt.Enemy(source, candidate) == 0 || rt.MapCheck(candidate, source) == 0 {
		return
	}
	direction := skullRawDirection54FA80(source)
	if rt.Front(source, int32(direction), candidate)&1 != 0 {
		state.Reachable = 1
	}
}
