package legacy

import "github.com/opennox/libs/types"

// twoPointsDirectionResults4E6E50 is GAME.EXE's 9-by-16 uint32 table at
// 0x005B8708. Rows are the {-1,0,1} X/Y facing cells and columns are the
// sector values returned by pointDirection4E6CE0.
var twoPointsDirectionResults4E6E50 = [9][16]uint32{
	{0, 9, 6, 0, 5, 1, 4, 0, 10, 8, 2, 0, 0, 0, 0, 0},
	{0, 1, 2, 0, 4, 5, 6, 0, 8, 9, 10, 0, 0, 0, 0, 0},
	{0, 5, 10, 0, 6, 4, 2, 0, 9, 1, 8, 0, 0, 0, 0, 0},
	{0, 8, 4, 0, 1, 9, 5, 0, 2, 10, 6, 0, 0, 0, 0, 0},
	{0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0},
	{0, 4, 8, 0, 2, 6, 10, 0, 1, 5, 9, 0, 0, 0, 0, 0},
	{0, 10, 5, 0, 9, 8, 1, 0, 6, 2, 4, 0, 0, 0, 0, 0},
	{0, 2, 1, 0, 8, 10, 9, 0, 4, 6, 5, 0, 0, 0, 0, 0},
	{0, 6, 9, 0, 10, 2, 8, 0, 5, 4, 1, 0, 0, 0, 0, 0},
}

// indexedDirectionRow4E6E50 reproduces the threshold-six quantization of the
// 256 signed int32 vector pairs at GAME.EXE 0x005B5E58. The original caller
// contract is a direction in [0,255]. Invalid values use the otherwise
// unreachable neutral row, avoiding the original out-of-bounds table read.
func indexedDirectionRow4E6E50(dir int32) int {
	switch {
	case dir < 0 || dir > 255:
		return 4
	case dir <= 17:
		return 5
	case dir <= 46:
		return 8
	case dir <= 83:
		return 7
	case dir <= 110:
		return 6
	case dir <= 147:
		return 3
	case dir <= 172:
		return 0
	case dir <= 209:
		return 1
	case dir <= 236:
		return 2
	default:
		return 5
	}
}

// twoPointsAndDirection4E6E50 reports how the sector from a to b relates to
// the indexed facing direction. Its return value is the original uint32 table
// value narrowed to the C int contract; callers inspect bits 0 and 1.
func twoPointsAndDirection4E6E50(a types.Pointf, dir int32, b types.Pointf) int32 {
	row := indexedDirectionRow4E6E50(dir)
	sector := pointDirection4E6CE0(a, b)
	return int32(twoPointsDirectionResults4E6E50[row][sector])
}
