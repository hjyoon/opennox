package legacy

/*
#include "GAME4_1.h"
extern uint32_t dword_5d4594_3835364;
*/
import "C"

import (
	"unsafe"

	"github.com/opennox/opennox/v1/common/memmap"
)

type mapgenTileEmissionAccess struct {
	fix          func(*mapgenWorldPoint, *mapgenWorldPoint)
	paint        func(*mapgenWorldPoint) bool
	rebuildWalls func() bool
	rebuild      func(*mapgenWorldPoint) bool
}

func mapgenEmitTile(position *mapgenWorldPoint, tileID uint32, access mapgenTileEmissionAccess) bool {
	if position == nil {
		return false
	}
	if tileID == 255 {
		return true
	}
	var fixed mapgenWorldPoint
	access.fix(position, &fixed)
	if !access.paint(&fixed) {
		return false
	}
	return !access.rebuildWalls() || access.rebuild(&fixed)
}

//export sub_51D5E0
func sub_51D5E0(position *C.float2) C.int {
	if position == nil {
		return 0
	}
	tileID := memmap.Uint32(0x973F18, 35912)
	point := (*mapgenWorldPoint)(unsafe.Pointer(position))
	return C.int(bool2int(mapgenEmitTile(point, tileID, mapgenTileEmissionAccess{
		fix: func(input, output *mapgenWorldPoint) {
			mapgenFixWorldCoords(input, output)
		},
		paint: func(point *mapgenWorldPoint) bool {
			return sub_51D8F0((*C.float2)(unsafe.Pointer(point))) != 0
		},
		rebuildWalls: func() bool { return C.dword_5d4594_3835364 == 1 },
		rebuild: func(point *mapgenWorldPoint) bool {
			return sub_527380((*C.float2)(unsafe.Pointer(point))) != 0
		},
	})))
}
