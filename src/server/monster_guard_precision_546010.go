package server

import (
	"math"

	"github.com/opennox/libs/types"
)

// Keep original PC53 operations and binary32 stores separate. Nearest is
// only the alternate original-execution profile; production uses ToZero.
func monsterGuardAdd546010(a, b float64, nearest bool) float64 {
	if nearest {
		return objectForceResult53_52DF80(a, b, float64(a+b))
	}
	return objectForceAdd53_52DF80(a, b)
}

func monsterGuardSub546010(a, b float64, nearest bool) float64 {
	if nearest {
		return objectForceResult53_52DF80(a, b, float64(a-b))
	}
	return objectForceSub53_52DF80(a, b)
}

func monsterGuardMul546010(a, b float64, nearest bool) float64 {
	if nearest {
		return objectForceResult53_52DF80(a, b, float64(a*b))
	}
	return objectForceMul53_52DF80(a, b)
}

func monsterGuardDiv546010(a, b float64, nearest bool) float64 {
	if nearest {
		return objectForceResult53_52DF80(a, b, float64(a/b))
	}
	return objectForceDiv53_52DF80(a, b)
}

func monsterGuardSpill546010(value float64, nearest bool) float32 {
	if nearest {
		return float32(value)
	}
	return experienceSpill32TowardZero(value)
}

func monsterGuardSquare546010(x, y float64, nearest bool) float64 {
	return monsterGuardAdd546010(monsterGuardMul546010(y, y, nearest), monsterGuardMul546010(x, x, nearest), nearest)
}

func monsterGuardDirection546010(direction uint32) types.Pointf {
	// The original LEA wraps in DWORD arithmetic, before addressing the
	// table. Preserve the raw DWORD for the later FACE_ANGLE payload.
	index := int32(direction*8) / 4
	return types.Ptf(monsterConfusedDirectionFloat545140(index), monsterConfusedDirectionFloat545140(index+1))
}

func monsterGuardTargetDirection546010(unit, target *Object, nearest bool) types.Pointf {
	xWide := monsterGuardSub546010(float64(target.PosVec.X), objectForceMemory32_52DF80(unit.PosVec.X), nearest)
	x := monsterGuardSpill546010(xWide, nearest)
	y := monsterGuardSub546010(float64(target.PosVec.Y), objectForceMemory32_52DF80(unit.PosVec.Y), nearest)
	square := monsterGuardSquare546010(float64(x), y, nearest)
	root := math.Sqrt(square)
	if !nearest {
		root = monsterMoveForceSqrtChop53_50D50C(square)
	}
	root = objectForceResult53_52DF80(square, 0, root)
	distance := monsterGuardAdd546010(root, float64(math.Float32frombits(0x3a83126f)), nearest)
	return types.Ptf(monsterGuardSpill546010(monsterGuardDiv546010(float64(x), distance, nearest), nearest), monsterGuardSpill546010(monsterGuardDiv546010(y, distance, nearest), nearest))
}

func monsterGuardUnorderedEqual546010(a, b float32) bool {
	// 005463B6/005463C3 test C3 alone: unordered also follows equality.
	return !(a < b || a > b)
}
