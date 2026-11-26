package legacy

/*
#include "GAME4_2.h"
#include "GAME4_3.h"
*/
import "C"

import (
	"unsafe"

	"github.com/opennox/opennox/v1/common/memmap"
)

const (
	floorFloodWorklistOffset = 16200
	floorFloodWorklistLimit  = 500
	floorFloodOverflowOffset = 22200
)

type floorFloodEntry struct {
	x, y, layer uint32
}

func floorFloodWorklistEntry(index int32) *floorFloodEntry {
	return memmap.PtrT[floorFloodEntry](mapgenTileStateBase, floorFloodWorklistOffset+12*uintptr(index))
}

func floorFloodOverflow() *uint32 {
	return memmap.PtrUint32(mapgenTileStateBase, floorFloodOverflowOffset)
}

func floorFloodWithinGrid(x, y C.int) bool {
	return x > 0 && x < 127 && y > 0 && y < 127
}

func floorFloodValidLayer(x, y, layer C.int) bool {
	return (layer&1 == 0 || y != 1) && (layer&2 == 0 || x != 1)
}

//export sub_51DD50
func sub_51DD50(x, y, layer, tile C.int) {
	if !floorFloodWithinGrid(x, y) || layer&3 == 0 {
		return
	}
	cell := mapgenFloorCell(int32(x), int32(y))
	if !(layer&2 != 0 && tile == cell.field_6 || layer&1 != 0 && tile == cell.field_1) || !floorFloodValidLayer(x, y, layer) {
		return
	}
	entry := floorFloodEntry{uint32(x), uint32(y), uint32(layer)}
	count := floorEditState.flood_count
	for i := int32(0); i < int32(count); i++ {
		if *floorFloodWorklistEntry(i) == entry {
			return
		}
	}
	if count >= floorFloodWorklistLimit {
		*floorFloodOverflow() = 1
		return
	}
	floorEditState.flood_count++
	*floorFloodWorklistEntry(int32(count)) = entry
}

//export sub_51DE30
func sub_51DE30(x, y, layer *C.uint32_t) C.int {
	count := int32(floorEditState.flood_count)
	if count <= 0 {
		return 0
	}
	floorEditState.flood_count--
	*x = C.uint32_t(floorFloodWorklistEntry(count - 1).x)
	*y = C.uint32_t(floorFloodWorklistEntry(int32(floorEditState.flood_count)).y)
	*layer = C.uint32_t(floorFloodWorklistEntry(int32(floorEditState.flood_count)).layer)
	return 1
}

func selectedFloorEdgeBrush() C.nox_floor_brush_t {
	return C.nox_floor_brush_t{
		tile_id:        C.uint32_t(memmap.Uint32(mapgenTileStateBase, mapgenTileIDOffset)),
		variation:      legacyWords.dword_5d4594_3835348,
		edge_id:        floorEditState.edge_tile,
		edge_variation: floorEditState.edge_variation,
	}
}

//export sub_543680
func sub_543680(position *C.float) C.int {
	if floorEditState.edge_tile == 255 {
		return 1
	}
	var fixed mapgenWorldPoint
	if !mapgenFixWorldCoords((*mapgenWorldPoint)(unsafe.Pointer(position)), &fixed) {
		return 0
	}
	brush := selectedFloorEdgeBrush()
	point := [2]C.int{C.int(x87TruncQwordLow(float64(fixed.x))), C.int(x87TruncQwordLow(float64(fixed.y)))}
	floorEditState.brush_mode = 1
	result := sub_5437E0(&point[0], &brush, 46)
	floorEditState.brush_mode = 0
	return result
}

func floorBrushGridPosition(x, y float64) (col, row, dx, dy C.int) {
	const scale = mapgenWallCoordinateScale / 2
	x = float64(x + 11.5)
	y = float64(y + 11.5)
	// The legacy coordinate calculation stores Y as a float for the remainder.
	return C.int(x87TruncQwordLow(x * scale)), C.int(x87TruncQwordLow(y * scale)),
		C.int(x87TruncQwordLow(x) % 46), C.int(int64(float32(y)) % 46)
}

//export sub_5437E0
func sub_5437E0(position *C.int, brush *C.nox_floor_brush_t, step C.int) C.int {
	point := unsafe.Slice(position, 2)
	x, y, dx, dy := floorBrushGridPosition(float64(point[0]), float64(point[1]))
	*floorFloodOverflow() = 0
	floorEditState.flood_count = 0
	tile := step
	if floorFloodWithinGrid(x, y) {
		layer := C.int(2)
		if dx <= dy {
			if step-dx > dy {
				x--
				layer = 1
			}
		} else if step-dx <= dy {
			layer = 1
		} else {
			y--
		}
		cell := mapgenFloorCell(int32(x), int32(y))
		tile = cell.field_6
		if layer == 1 {
			tile = cell.field_1
		}
		sub_51DD50(x, y, layer, tile)
	}

	for i := int32(0); i < int32(floorEditState.flood_count); i++ {
		entry := floorFloodWorklistEntry(i)
		x, y := C.int(entry.x), C.int(entry.y)
		if entry.layer&2 != 0 {
			sub_51DD50(x, y, 1, tile)
			sub_51DD50(x, y+1, 1, tile)
			sub_51DD50(x-1, y, 1, tile)
			sub_51DD50(x-1, y+1, 1, tile)
		} else if entry.layer&1 != 0 {
			sub_51DD50(x+1, y, 2, tile)
			sub_51DD50(x+1, y-1, 2, tile)
			sub_51DD50(x, y, 2, tile)
			sub_51DD50(x, y-1, 2, tile)
		}
	}
	var queueX, queueY, layer C.uint32_t
	for sub_51DE30(&queueX, &queueY, &layer) != 0 {
		x, y := C.int(queueX), C.int(queueY)
		if layer&2 != 0 {
			sub_543BC0(x, y, 1, tile, brush, 1)
			sub_543BC0(x, y+1, 1, tile, brush, 4)
			sub_543BC0(x-1, y, 1, tile, brush, 3)
			sub_543BC0(x-1, y+1, 1, tile, brush, 6)
			sub_543BC0(x, y-1, 2, tile, brush, 0)
			sub_543BC0(x+1, y, 2, tile, brush, 2)
			sub_543BC0(x-1, y, 2, tile, brush, 5)
			sub_543BC0(x, y+1, 2, tile, brush, 7)
		} else if layer&1 != 0 {
			sub_543BC0(x+1, y, 2, tile, brush, 4)
			sub_543BC0(x+1, y-1, 2, tile, brush, 1)
			sub_543BC0(x, y, 2, tile, brush, 6)
			sub_543BC0(x, y-1, 2, tile, brush, 3)
			sub_543BC0(x, y-1, 1, tile, brush, 0)
			sub_543BC0(x+1, y, 1, tile, brush, 2)
			sub_543BC0(x-1, y, 1, tile, brush, 5)
			sub_543BC0(x, y+1, 1, tile, brush, 7)
		}
	}
	return C.int(bool2int(*floorFloodOverflow() == 0))
}

//export sub_543BC0
func sub_543BC0(x, y, layer, tile C.int, brush *C.nox_floor_brush_t, edgeVariation C.int) {
	if !floorFloodWithinGrid(x, y) || !floorFloodValidLayer(x, y, layer) {
		return
	}
	cell := mapgenFloorCell(int32(x), int32(y))
	if layer&2 != 0 && tile == cell.field_6 {
		return
	}
	if (layer&1 == 0 || tile != cell.field_1) && brush != nil {
		brush.edge_variation = C.uint32_t(edgeVariation)
		nox_floorBrushApplyNative_51DA70(x, y, brush, layer, 1)
	}
}

//export nox_xxx_tileSubtile_544310
func nox_xxx_tileSubtile_544310(position *C.float2) C.int {
	floorEditState.brush_mode = 1
	x, y, dx, dy := floorBrushGridPosition(float64(position.field_0), float64(position.field_4))
	brush := selectedFloorEdgeBrush()
	brush.anchor_x, brush.anchor_y = -1, -1
	var selected *C.nox_floor_brush_t
	if brush.tile_id != 255 {
		selected = &brush
	}
	result := sub_51D9C0(x, y, dx, dy, selected)
	floorEditState.brush_mode = 0
	return result
}
