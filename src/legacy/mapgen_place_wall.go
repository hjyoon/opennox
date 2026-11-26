package legacy

/*
#include "GAME4_2.h"

*/
import "C"

import (
	"unsafe"

	"github.com/opennox/libs/wall"

	"github.com/opennox/opennox/v1/common/memmap"
)

//export sub_526E60
func sub_526E60(position *C.float) C.int {
	if memmap.Uint32(0x973F18, 35948) == 255 {
		return 1
	}
	var fixed mapgenWorldPoint
	if !mapgenFixWorldCoords((*mapgenWorldPoint)(unsafe.Pointer(position)), &fixed) {
		return 0
	}
	x, y, valid := mapgenWallGridCell(fixed.x, fixed.y)
	if !valid {
		return 0
	}
	target := asWallP(nox_server_getWallAtGrid_410580(x, y))
	exists := target != nil
	if !exists {
		target = asWallP(nox_xxx_wallCreateAt_410250(x, y))
		if target == nil {
			return 0
		}
	}

	target.Tile1 = memmap.Uint8(0x973F18, 35948)
	direction := memmap.Uint8(0x973F18, 35952)
	if exists && archiveAndMapgenStorage.merge_walls == 1 {
		direction = WallMergeDirections(direction, target.Dir0)
	}
	target.Dir0 = direction

	variation := memmap.Uint8(0x973F18, 35956)
	if archiveAndMapgenStorage.repeat_walls != 0 {
		variation = 0
		if period := nox_xxx_map_410E00(int32(target.Tile1)); period != 0 {
			variation = target.X5 % period
		}
	}
	if variation >= nox_xxx_mapWallMaxVariation_410DD0(int32(target.Tile1), int32(direction), 0) {
		variation = 0
	}
	target.Field2 = variation
	target.Flags4 = wall.Flag8
	return 1
}
