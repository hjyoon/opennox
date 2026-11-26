package legacy

/*
#include "GAME4_1.h"
#include "floor_brush_native.h"
*/
import "C"

import (
	"unsafe"

	"github.com/opennox/opennox/v1/common/memmap"
)

type mapgenFloorBrush struct {
	tileID        uint32
	variation     uint32
	pattern       uint32
	anchorX       int32
	anchorY       int32
	kind          uint8
	_             [3]byte
	edgeID        uint32
	edgeVariation uint32
}

var (
	_ [32]byte = [unsafe.Sizeof(mapgenFloorBrush{})]byte{}
	_ [20]byte = [unsafe.Offsetof(mapgenFloorBrush{}.kind)]byte{}
	_ [24]byte = [unsafe.Offsetof(mapgenFloorBrush{}.edgeID)]byte{}
)

func mapgenReadTileCoordinate(position *mapgenWorldPoint, tileID, variation uint32,
	place func(int32, int32, int32, int32, *mapgenFloorBrush) int32) int32 {
	if position == nil {
		return 0
	}
	const (
		halfTile = 11.5
		cellSize = 46
		scale    = mapgenWallCoordinateScale / 2
	)
	x := addFloat64TowardZero(float64(position.x), halfTile)
	y := addFloat64TowardZero(float64(position.y), halfTile)
	cellX := x87TruncQwordLow(mulFloat64TowardZero(x, scale))
	cellY := x87TruncQwordLow(mulFloat64TowardZero(y, scale))
	dx := x87TruncQwordLow(x) % cellSize
	// The original spills only Y before converting the remainder.
	dy := x87TruncQwordLow(float64(storeFloat32TowardZero(y))) % cellSize
	if tileID == 255 {
		return place(cellX, cellY, dx, dy, nil)
	}
	brush := mapgenFloorBrush{tileID: tileID, variation: variation, anchorX: -1, anchorY: -1, edgeID: 255}
	return place(cellX, cellY, dx, dy, &brush)
}

func mapgenApplyTileQuarter(x, y, dx, dy int32, brush *mapgenFloorBrush,
	apply func(int32, int32, *mapgenFloorBrush, int32) int32) int32 {
	if x-1 <= 0 || x >= 127 || y-1 <= 0 || y >= 127 {
		return 0
	}
	above := 46-dx <= dy
	if dx <= dy {
		if above {
			return apply(x, y, brush, 2)
		}
		return apply(x-1, y, brush, 1)
	}
	if above {
		return apply(x, y, brush, 1)
	}
	return apply(x, y-1, brush, 2)
}

func mapgenWriteFloorBrush(x, y int32, brush *mapgenFloorBrush, layer int32) int32 {
	return int32(nox_floorBrushApplyNative_51DA70(C.int(x), C.int(y),
		(*C.nox_floor_brush_t)(unsafe.Pointer(brush)), C.int(layer), 0))
}

func mapgenPlaceTileQuarter(x, y, dx, dy int32, brush *mapgenFloorBrush) int32 {
	// Reject invalid grid inputs before the native row lookup.
	if x < 2 || y < 2 {
		return 0
	}
	return mapgenApplyTileQuarter(x, y, dx, dy, brush, mapgenWriteFloorBrush)
}

//export sub_51D8F0
func sub_51D8F0(position *C.float2) C.int {
	tileID := memmap.Uint32(0x973F18, 35912)
	variation := uint32(legacyWords.dword_5d4594_3835348)
	return C.int(mapgenReadTileCoordinate((*mapgenWorldPoint)(unsafe.Pointer(position)), tileID, variation, mapgenPlaceTileQuarter))
}

//export sub_51D9C0
func sub_51D9C0(x, y, dx, dy C.int, brush *C.nox_floor_brush_t) C.int {
	return C.int(mapgenPlaceTileQuarter(int32(x), int32(y), int32(dx), int32(dy), (*mapgenFloorBrush)(unsafe.Pointer(brush))))
}
