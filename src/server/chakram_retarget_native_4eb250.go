package server

import (
	"math"

	"github.com/opennox/libs/types"
)

type chakramRetargetArithmetic4EB250 struct{ nearest bool }

//go:noinline
func (f chakramRetargetArithmetic4EB250) add(a, b float64) float64 {
	if math.IsInf(a, 0) && a == -b {
		return math.Float64frombits(0xfff8000000000000)
	}
	if f.nearest {
		return a + b
	}
	return experienceAdd53TowardZero(a, b)
}

//go:noinline
func (f chakramRetargetArithmetic4EB250) sub(a, b float64) float64 {
	return firewalkSubtract53_52ED40(a, b, !f.nearest)
}

//go:noinline
func (f chakramRetargetArithmetic4EB250) mul(a, b float64) float64 {
	if a == 0 && math.IsInf(b, 0) || b == 0 && math.IsInf(a, 0) {
		return math.Float64frombits(0xfff8000000000000)
	}
	if f.nearest {
		return a * b
	}
	return experienceMul53TowardZero(a, b)
}

//go:noinline
func (f chakramRetargetArithmetic4EB250) div(a, b float64) float64 {
	if a == 0 && b == 0 || math.IsInf(a, 0) && math.IsInf(b, 0) {
		return math.Float64frombits(0xfff8000000000000)
	}
	if f.nearest {
		return a / b
	}
	return playerStatDiv53TowardZero4EEDC0(a, b)
}

func (f chakramRetargetArithmetic4EB250) sqrt(value float64) float64 {
	if f.nearest {
		return math.Sqrt(value)
	}
	return monsterMoveForceSqrtChop53_50D50C(value)
}

func (f chakramRetargetArithmetic4EB250) spill(value float64) float32 {
	if f.nearest {
		return float32(value)
	}
	return experienceSpill32TowardZero(value)
}

// ChakramRetargetNative4EB250 preserves the raw shared-scratch C contract.
// Its entry-cached update pointer remains live across world calls, while
// nested searches deliberately replace the shared source/exclusions/result.
func (source *Object) ChakramRetargetNative4EB250(state *ChakramRetargetState4EB250, runtime ChakramRetargetRuntime4EB250) *Object {
	update := (*ChakramUpdateData)(source.UpdateData)
	state.Selected = nil
	state.LastHit = update.LastHit
	state.Owner = source.ObjOwner
	state.Source = source
	state.BestDistance = math.Float32frombits(chakramRetargetBestBits4EB250)
	f := chakramRetargetArithmetic4EB250{nearest: runtime.RoundToNearest}
	rect := types.Rectf{
		Min: types.Ptf(f.spill(f.sub(float64(source.PosVec.X), 400)), f.spill(f.sub(float64(source.PosVec.Y), 400))),
		Max: types.Ptf(f.spill(f.add(float64(source.PosVec.X), 400)), f.spill(f.add(float64(source.PosVec.Y), 400))),
	}
	runtime.EachRect(rect, func(candidate *Object, point *types.Pointf) {
		candidate.ChakramRetargetCandidateNative4EB340(point, state, runtime)
	}, &source.PosVec)
	result := state.Selected
	if result == nil {
		return nil
	}
	update.ReturnState = chakramReturnStateSeek4EAF00
	dx := f.sub(float64(result.PosVec.X), float64(source.PosVec.X))
	dyExtended := f.sub(float64(result.PosVec.Y), float64(source.PosVec.Y))
	// FST keeps the extended Y in x87 while spilling the driver argument.
	dy := f.spill(dyExtended)
	square := f.add(f.mul(dyExtended, float64(dy)), f.mul(dx, dx))
	denominator := f.spill(f.add(f.sqrt(square), float64(chakramRetargetEpsilon4EB250)))
	source.VelVec.X = f.spill(f.div(f.mul(dx, float64(source.SpeedCur)), float64(denominator)))
	source.VelVec.Y = f.spill(f.div(f.mul(float64(dy), float64(source.SpeedCur)), float64(denominator)))
	return result
}

func (candidate *Object) ChakramRetargetCandidateNative4EB340(point *types.Pointf, state *ChakramRetargetState4EB250, runtime ChakramRetargetRuntime4EB250) {
	if uint32(candidate.ObjClass)&chakramRetargetClassMask4EB250 == 0 || uint32(candidate.ObjFlags)&chakramRetargetFlagsMask4EB250 != 0 {
		return
	}
	if runtime.HasEnchant(candidate, 0) || candidate == state.Owner || candidate == state.LastHit {
		return
	}
	if !runtime.CanSee(candidate, state.Source) {
		return
	}
	f := chakramRetargetArithmetic4EB250{nearest: runtime.RoundToNearest}
	dx := f.sub(float64(point.X), float64(candidate.PosVec.X))
	dy := f.sub(float64(point.Y), float64(candidate.PosVec.Y))
	distance := f.add(f.mul(dy, dy), f.mul(dx, dx))
	// Original FCOM C0/C6 gates accept unordered values on both comparisons.
	if distance > 160000 || distance >= float64(state.BestDistance) {
		return
	}
	state.BestDistance = f.spill(distance)
	state.Selected = candidate
}
