package server

import (
	"image"

	"github.com/opennox/libs/types"
)

// IndexedDirectionVector509E20 is the exact two-dword output record written by
// GAME.EXE 00509E20. Both components are signed 32-bit quantized directions.
type IndexedDirectionVector509E20 struct {
	X int32
	Y int32
}

func indexedDirectionClassify509E20(value, threshold int32) int32 {
	if value > threshold {
		return 1
	}
	if value >= -threshold {
		return 0
	}
	return -1
}

// indexedDirectionVector509E20 is the threshold-six classification of the
// exact 256 signed int32 vector pairs at GAME.EXE 005B5E58. The source table
// SHA-256 is 69f2eae1505c2e27de6bf104de077085994f00e4eb1d4ab3c68bd13916d7b079.
func indexedDirectionVector509E20(direction int32) IndexedDirectionVector509E20 {
	switch {
	case direction < 0 || direction > 255:
		panic("direction outside sealed GAME.EXE vector table")
	case direction <= 17:
		return IndexedDirectionVector509E20{X: 1, Y: 0}
	case direction <= 46:
		return IndexedDirectionVector509E20{X: 1, Y: 1}
	case direction <= 83:
		return IndexedDirectionVector509E20{X: 0, Y: 1}
	case direction <= 110:
		return IndexedDirectionVector509E20{X: -1, Y: 1}
	case direction <= 147:
		return IndexedDirectionVector509E20{X: -1, Y: 0}
	case direction <= 172:
		return IndexedDirectionVector509E20{X: -1, Y: -1}
	case direction <= 209:
		return IndexedDirectionVector509E20{X: 0, Y: -1}
	case direction <= 236:
		return IndexedDirectionVector509E20{X: 1, Y: -1}
	default:
		return IndexedDirectionVector509E20{X: 1, Y: 0}
	}
}

// IndexedDirection509E20 preserves GAME.EXE 00509E20's fixed-width table
// lookup, threshold comparisons, output-store order, and residual return
// value. Valid callers pass a direction in [0,255]. Reading adjacent PE data
// for a malformed direction is intentionally not recreated.
func IndexedDirection509E20(direction int32, out *IndexedDirectionVector509E20) int32 {
	vector := indexedDirectionVector509E20(direction)
	out.X = vector.X
	out.Y = vector.Y
	if vector.Y > 0 {
		return 6
	}
	return -6
}

func indexedDirection509E20(direction int16) image.Point {
	var out IndexedDirectionVector509E20
	IndexedDirection509E20(int32(direction), &out)
	return image.Pt(int(out.X), int(out.Y))
}

// DirectionToAngle509E00 binds the centered direction table to the exact
// signed two-dword input record used by GAME.EXE 00509E00.
func DirectionToAngle509E00(data *DirectionInitData) uint32 {
	return directionToAngleNative509E00(data)
}

// DirectionIndexToAngle509E90 returns the full dword at a valid direct table
// index. GAME.EXE does not apply the modulo operation shown by the decompiler.
func DirectionIndexToAngle509E90(index int32) uint32 {
	if index < 0 {
		panic("direction index outside sealed GAME.EXE table")
	}
	return directionIndexToAngleNative509E90(uint32(index))
}

// DirectionOctant509EA0 converts an indexed direction to the original
// row-major {-1,0,1} X/Y cell number in [0,8].
func DirectionOctant509EA0(direction int32) int32 {
	var out IndexedDirectionVector509E20
	IndexedDirection509E20(direction, &out)
	return out.X + 3*out.Y + 4
}

// NormalizeVector509F20 preserves the original x87/binary32 arithmetic and
// memory order of GAME.EXE 00509F20.
func NormalizeVector509F20(point *types.Pointf) {
	chestOpenNormalizeVector509F20(point)
}
