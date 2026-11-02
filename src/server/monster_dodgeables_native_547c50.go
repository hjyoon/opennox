package server

import (
	"math"
	"unsafe"

	"github.com/opennox/libs/types"
	"github.com/opennox/opennox/v1/common/unit/ai"
)

// x87 invalid arithmetic returns the negative indefinite NaN. Preserve its
// bits across native architectures; operand NaNs retain their own payload.
func monsterDodgeMul547C50(a, b float64) float64 {
	if math.IsInf(a, 0) && b == 0 || math.IsInf(b, 0) && a == 0 {
		return math.Float64frombits(0xfff8000000000000)
	}
	return monsterMoveForceMulChop53_50D4FE(a, b)
}

func monsterDodgeAdd547C50(a, b float64) float64 {
	if math.IsInf(a, 0) && math.IsInf(b, 0) && math.Signbit(a) != math.Signbit(b) {
		return math.Float64frombits(0xfff8000000000000)
	}
	return monsterMoveToRunAddChop53_544434(a, b)
}

// MonsterCheckDodgeablesNative547C50 probes the original mutable local ray.
// Directions 0..255 use the original table. The raw entry keeps original
// dereferences; class and nil admission belong to its callers.
func (unit *Object) MonsterCheckDodgeablesNative547C50(
	randomFloat func(float32, float32) float64,
	randomInt func(int, int) int,
	trace func(*[4]float32, uint32) int,
	obstacles func(*Object, *[4]float32) int,
	tile func(*types.Pointf) int,
	pop func(*Object) int,
	push func(*Object, ai.ActionType) *AIStackItem,
	frame, fps func() uint32,
) int {
	direction := int32(int16(unit.Direction1)) * 2
	sine := monsterConfusedDirectionFloat545140(direction + 1)
	cosine := monsterConfusedDirectionFloat545140(direction)
	ray := [4]float32{unit.PosVec.X, unit.PosVec.Y}
	for attempt := 0; attempt < 5; attempt++ {
		// FMUL547CA2 reads original object+176: the collision radius,
		// independent of current movement speed at object+544.
		sample := randomFloat(2, 3)
		raw := monsterDodgeMul547C50(sample, float64(unit.Shape.Circle.R))
		distance := float32(monsterMoveToRunSpill544440(raw))
		if raw > 15 {
			distance = 15
		}
		if randomInt(0, 100) < 50 {
			distance = -distance
		}
		x := monsterDodgeMul547C50(float64(distance), float64(-sine))
		x = monsterDodgeAdd547C50(x, float64(unit.PosVec.X))
		ray[2] = float32(monsterMoveToRunSpill544440(x))
		y := monsterDodgeMul547C50(float64(distance), float64(cosine))
		y = monsterDodgeAdd547C50(y, float64(unit.PosVec.Y))
		ray[3] = float32(monsterMoveToRunSpill544440(y))
		if trace(&ray, 1) == 0 || obstacles(unit, &ray) == 0 || tile((*types.Pointf)(unsafe.Pointer(&ray[2]))) == 6 {
			continue
		}
		unit.MonsterPopAttackActionsNative5471B0(pop)
		if timer := push(unit, ai.DEPENDENCY_TIME); timer != nil {
			current := frame()
			rate := fps()
			timer.Args[0] = uintptr(current + rate)
		}
		if dodge := push(unit, ai.ACTION_DODGE); dodge != nil {
			dodge.Args[0] = uintptr(math.Float32bits(ray[2]))
			dodge.Args[1] = uintptr(math.Float32bits(ray[3]))
			dodge.Args[2] = 0
		}
		return 1
	}
	return 0
}
