package legacy

import (
	"unsafe"

	"github.com/opennox/libs/types"
)

var (
	_ = [1]struct{}{}[8-unsafe.Sizeof(types.Pointf{})]
	_ = [1]struct{}{}[0-unsafe.Offsetof(types.Pointf{}.X)]
	_ = [1]struct{}{}[4-unsafe.Offsetof(types.Pointf{}.Y)]
)

const (
	pointDirectionSlopeSmall4E6CE0 float32 = 0x1.a6f4bep-2
	pointDirectionSlopeLarge4E6CE0 float32 = 0x1.35e51p+1
)

// pointDirectionProjection4E6CE0 evaluates one of GAME.EXE's x87 linear
// projections. Both inputs and the coefficient originate as float32 values;
// float64 can represent the product and the cancellation range exactly.
func pointDirectionProjection4E6CE0(dx, dy, slope float32) float64 {
	return float64(dx)*float64(slope) - float64(dy)
}

func pointDirectionResult4E6CE0(mask uint8) int {
	switch mask {
	case 0:
		return 2
	case 2:
		return 6
	case 3:
		return 4
	case 4:
		return 10
	case 11:
		return 5
	case 12:
		return 8
	case 13:
		return 9
	case 15:
		return 1
	default:
		return 0
	}
}

// pointDirection4E6CE0 classifies the vector from a to b into the original
// eight direction sectors. Three projections are rounded through float32
// scratch slots before comparison. The positive large-slope projection stays
// in the x87 register and must be compared before that rounding.
func pointDirection4E6CE0(a, b types.Pointf) int {
	dx := b.X - a.X
	dy := b.Y - a.Y

	projection8 := float32(pointDirectionProjection4E6CE0(dx, dy, pointDirectionSlopeSmall4E6CE0))
	projection4 := pointDirectionProjection4E6CE0(dx, dy, pointDirectionSlopeLarge4E6CE0)
	projection2 := float32(pointDirectionProjection4E6CE0(dx, dy, -pointDirectionSlopeLarge4E6CE0))
	projection1 := float32(pointDirectionProjection4E6CE0(dx, dy, -pointDirectionSlopeSmall4E6CE0))

	var mask uint8
	// An x87 unordered comparison sets C0, taking the same branch as a
	// negative value. Go's >= comparison is false for NaN and therefore
	// reproduces the original bit construction, including both signed zeros.
	if projection8 >= 0 {
		mask |= 8
	}
	if projection4 >= 0 {
		mask |= 4
	}
	if projection2 >= 0 {
		mask |= 2
	}
	if projection1 >= 0 {
		mask |= 1
	}
	return pointDirectionResult4E6CE0(mask)
}
