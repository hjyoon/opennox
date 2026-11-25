package legacy

/*
#include "GAME1_2.h"
*/
import "C"

import (
	"unsafe"

	"github.com/opennox/libs/wall"

	"github.com/opennox/opennox/v1/server"
)

const mapWallWorldSize = wall.GridStep * server.WallGridSize

func mapWallRegionContains(x, y int32, region *[8]int32) bool {
	if x < 58 || y < 58 || x > mapWallWorldSize || y > mapWallWorldSize {
		return false
	}
	x, y = x/wall.GridStep, y/wall.GridStep
	topX, topY := region[0]/wall.GridStep, region[1]/wall.GridStep
	leftX, leftY := region[2]/wall.GridStep, region[3]/wall.GridStep
	rightX, rightY := region[4]/wall.GridStep, region[5]/wall.GridStep
	bottomY := region[7] / wall.GridStep
	if x < leftX || x > rightX || y < topY || y > bottomY {
		return false
	}

	left := topX + topY - y
	if y > leftY {
		left = leftX + y - leftY
	}
	right := topX + y - topY
	if y > rightY {
		right = rightX + rightY - y
	}
	return x >= left && x <= right
}

func mapWallRegionBounds(region *[8]int32) [4]int32 {
	return [4]int32{
		max(0, min(region[2], region[4])),
		max(0, min(region[1], region[7])),
		min(mapWallWorldSize-1, max(region[2], region[4])),
		min(mapWallWorldSize-1, max(region[1], region[7])),
	}
}

//export nox_xxx_wallMath_427F30
func nox_xxx_wallMath_427F30(pos *C.int2, region *C.int) C.int {
	return C.int(bool2int(mapWallRegionContains(int32(pos.field_0), int32(pos.field_4), (*[8]int32)(unsafe.Pointer(region)))))
}

//export sub_428170
func sub_428170(region unsafe.Pointer, bounds *C.int4) C.int {
	*(*[4]int32)(unsafe.Pointer(bounds)) = mapWallRegionBounds((*[8]int32)(region))
	return 0
}
