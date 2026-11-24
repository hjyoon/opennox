package legacy

/*
#include "GAME4_2.h"
*/
import "C"

import "unsafe"

const (
	mapgenFloorRectangle = iota
	mapgenFloorBeveled
	mapgenFloorIrregular
)

type mapgenFloorFillArea struct {
	origin        mapgenWorldPoint
	width, height int32
}

type mapgenFloorFillOps struct {
	rectangle func(mapgenWorldPoint, int32, int32)
	column    func(mapgenWorldPoint, int32)
	random    func(int32, int32) int32
}

func mapgenFloorFillCellOffset(origin float32, cells int32) float32 {
	return float32(float64(cells)*mapgenCellSize + float64(origin))
}

func (area mapgenFloorFillArea) center() mapgenWorldPoint {
	const halfCell = mapgenCellSize / 2
	return mapgenWorldPoint{
		x: float32(float64(area.width)*halfCell + float64(area.origin.x)),
		y: float32(float64(area.height)*halfCell + float64(area.origin.y)),
	}
}

func mapgenDrawBeveledFloorRows(point mapgenWorldPoint, span, width, rows int32, left float32, stepY int32,
	rectangle func(mapgenWorldPoint, int32, int32)) {
	for i := int32(0); i < rows; i++ {
		rectangle(point, span, 1)
		span = min(span+2, width)
		point.x = mapgenFloorFillCellOffset(point.x, -1)
		point.y = mapgenFloorFillCellOffset(point.y, stepY)
		if span == width {
			point.x = left
		}
	}
}

func (area mapgenFloorFillArea) drawBeveled(rectangle func(mapgenWorldPoint, int32, int32)) {
	inset := area.width / 3
	offset := float64(inset) * mapgenCellSize
	span := area.width - 2*inset
	halfHeight := area.height / 2
	top := mapgenWorldPoint{x: float32(offset + float64(area.origin.x)), y: area.origin.y}
	mapgenDrawBeveledFloorRows(top, span, area.width, halfHeight+1, area.origin.x, 1, rectangle)

	// The bottom pass reuses a rounded copy of the first pass's offset.
	bottom := mapgenWorldPoint{
		x: float32(float64(float32(offset)) + float64(area.origin.x)),
		y: mapgenFloorFillCellOffset(area.origin.y, area.height-1),
	}
	mapgenDrawBeveledFloorRows(bottom, span, area.width, halfHeight, area.origin.x, -1, rectangle)
}

func mapgenIrregularFloorSpan(length int32, random func(int32, int32) int32) (start, span int32) {
	span = random(1, length-3)
	start = random(1, length-span-2)
	return start, span
}

func (area mapgenFloorFillArea) drawIrregular(ops mapgenFloorFillOps) {
	inside := mapgenWorldPoint{
		x: mapgenFloorFillCellOffset(area.origin.x, 1),
		y: mapgenFloorFillCellOffset(area.origin.y, 1),
	}
	ops.rectangle(inside, area.width-2, area.height-2)

	if area.width == 3 {
		top := mapgenWorldPoint{x: inside.x, y: area.origin.y}
		ops.rectangle(top, 1, 1)
		top.y = mapgenFloorFillCellOffset(top.y, 2)
		ops.rectangle(top, 1, 1)
	} else {
		top := area.origin.y
		bottom := mapgenFloorFillCellOffset(top, area.height-1)
		for _, y := range [...]float32{top, bottom} {
			start, span := mapgenIrregularFloorSpan(area.width, ops.random)
			point := mapgenWorldPoint{x: mapgenFloorFillCellOffset(area.origin.x, start), y: y}
			ops.rectangle(point, span, 1)
		}
	}

	if area.height == 3 {
		left := mapgenWorldPoint{x: area.origin.x, y: inside.y}
		ops.rectangle(left, 1, 1)
		left.x = mapgenFloorFillCellOffset(left.x, 2)
		ops.rectangle(left, 1, 1)
	} else {
		left := area.origin.x
		right := mapgenFloorFillCellOffset(left, area.width-1)
		for _, x := range [...]float32{left, right} {
			start, span := mapgenIrregularFloorSpan(area.height, ops.random)
			point := mapgenWorldPoint{x: x, y: mapgenFloorFillCellOffset(area.origin.y, start)}
			ops.column(point, span)
		}
	}
}

func mapgenPlaceFloorFill(layer C.int, fill *C.nox_mapgen_decor_fill_legacy, area mapgenFloorFillArea, kind uint32) C.int {
	if kind == mapgenFloorIrregular && (area.width < 3 || area.height < 3) {
		return 0
	}
	nox_xxx_tileGetDefByName_51D4D0(&fill.tile_name[0])
	rectangle := func(point mapgenWorldPoint, width, height int32) {
		sub_5245A0(layer, (*C.float)(unsafe.Pointer(&point)), C.int(width), C.int(height))
	}
	switch kind {
	case mapgenFloorBeveled:
		area.drawBeveled(rectangle)
	case mapgenFloorIrregular:
		area.drawIrregular(mapgenFloorFillOps{
			rectangle: rectangle,
			column: func(point mapgenWorldPoint, height int32) {
				sub_524610(layer, (*C.float)(unsafe.Pointer(&point)), C.int(height))
			},
			random: func(min, max int32) int32 {
				return int32(nox_xxx_mapGenRandFunc_526AC0(C.int(min), C.int(max)))
			},
		})
	default:
		rectangle(area.origin, area.width, area.height)
	}
	sub_544020(&fill.object_name[0])
	center := area.center()
	return sub_543680((*C.float)(unsafe.Pointer(&center)))
}

func mapgenApplyFloorFill(layer, token C.int, origin *C.float, dimensions *C.int, kind uint32) C.int {
	fill := mapgenResolveDecorFill(C.uint32_t(token))
	if fill == nil {
		return 0
	}
	size := (*[2]C.int)(unsafe.Pointer(dimensions))
	area := mapgenFloorFillArea{
		origin: *(*mapgenWorldPoint)(unsafe.Pointer(origin)),
		width:  int32(size[0]),
		height: int32(size[1]),
	}
	return mapgenPlaceFloorFill(layer, fill, area, kind)
}

//export sub_524950
func sub_524950(layer, token C.int, origin *C.float, dimensions *C.int) C.int {
	return mapgenApplyFloorFill(layer, token, origin, dimensions, mapgenFloorRectangle)
}

//export sub_5249C0
func sub_5249C0(layer, token C.int, origin *C.float, dimensions *C.int) C.int {
	return mapgenApplyFloorFill(layer, token, origin, dimensions, mapgenFloorBeveled)
}

//export sub_524B50
func sub_524B50(layer, token C.int, origin *C.float, dimensions *C.int) {
	mapgenApplyFloorFill(layer, token, origin, dimensions, mapgenFloorIrregular)
}
