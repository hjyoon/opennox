package legacy

/*
#include "GAME2_2.h"
*/
import "C"

import (
	"encoding/binary"
	"unsafe"

	"github.com/opennox/opennox/v1/common/memmap"
)

const floorGridStep = 46

var (
	floorTileDraw = Get_nox_xxx_tileDraw_4815E0()
	floorEdgeDraw = Get_nox_xxx_drawTexEdgesProbably_481900()
)

func Set_func_587000_154940(draw unsafe.Pointer) {
	floorTileDraw = draw
}

func Set_func_587000_154944(draw unsafe.Pointer) {
	floorEdgeDraw = draw
}

//export sub_476D70
func sub_476D70(dst *C.uint32_t, src *C.int, size C.uint) C.short {
	var last uint16
	offset := 0
	for range size >> 2 {
		from := unsafe.Slice((*byte)(unsafe.Add(unsafe.Pointer(src), offset)), 4)
		to := unsafe.Slice((*byte)(unsafe.Add(unsafe.Pointer(dst), offset)), 4)
		word := binary.LittleEndian.Uint32(from)
		binary.LittleEndian.PutUint32(to, word)
		last = uint16(word)
		offset += 4
	}
	if size&3 != 0 {
		from := unsafe.Slice((*byte)(unsafe.Add(unsafe.Pointer(src), offset)), 2)
		to := unsafe.Slice((*byte)(unsafe.Add(unsafe.Pointer(dst), offset)), 2)
		last = binary.LittleEndian.Uint16(from)
		binary.LittleEndian.PutUint16(to, last)
	}
	return C.short(last)
}

//export nox_xxx_tileCallDrawEdges_481BF0
func nox_xxx_tileCallDrawEdges_481BF0(pos *C.int2, edges *C.nox_tile_list_node_t) {
	for edge := edges; edge != nil; edge = edge.next {
		switch fn := floorEdgeDraw; fn {
		case nil, Get_nox_xxx_drawNoEdges_4818F0():
			return
		case Get_nox_xxx_drawTexEdgesProbably_481900():
			nox_xxx_drawTexEdgesProbably_481900(pos, edge)
		default:
			callFloorEdgeDraw(fn, pos, edge)
		}
	}
}

func drawFloorCellPart(pos *C.int2, tileIndex C.uint16_t, variation C.int, edges *C.nox_tile_list_node_t) {
	if C.uint32_t(tileIndex) >= floorTileStorage.count {
		return
	}
	tile := &floorTileStorage.definitions[tileIndex]
	index := int(variation) + int(tile.field_46)
	if tile.data_32 == nil || index < 0 {
		return
	}
	image := floorDrawImageAt(unsafe.Pointer(tile.data_32), index)
	switch fn := floorTileDraw; fn {
	case nil:
		return
	case Get_nox_xxx_tileDraw_4815E0():
		nox_xxx_tileDraw_4815E0(pos, image, tileIndex)
	case Get_sub_481770():
		sub_481770(pos, image, tileIndex)
	default:
		callFloorTileDraw(fn, pos, image, tileIndex)
	}
	*memmap.PtrUint32(0x85B3FC, 228+4*uintptr(tileIndex)) = 1
	nox_xxx_tileCallDrawEdges_481BF0(pos, edges)
}

func drawFloorCell(cell *C.obj_5D4594_2650668_t, x, y int32) {
	if cell.field_0&2 != 0 {
		pos := C.int2{field_0: C.int(x), field_4: C.int(y + 23)}
		drawFloorCellPart(&pos, C.uint16_t(cell.field_6), cell.field_7, cell.field_10)
	}
	if cell.field_0&1 != 0 {
		pos := C.int2{field_0: C.int(x + 23), field_4: C.int(y)}
		drawFloorCellPart(&pos, C.uint16_t(cell.field_1), cell.field_2, cell.field_5)
	}
}

func floorVisibleCells(vp *nox_draw_viewport_t, width, height, capacity int32) (left, top, right, bottom int32) {
	last := capacity - 1
	x, y := int32(vp.field_4-vp.x1), int32(vp.field_5-vp.y1)
	left = max(0, (x-11)/floorGridStep)
	top = max(0, (y-11)/floorGridStep-1)
	right, bottom = left+width-1, top+height
	if right > last {
		left, right = last-width, last
	}
	if bottom > last {
		top, bottom = last-height, last
	}
	return
}

func floorTileNeedsTextures(tile C.int) bool {
	return tile >= 0 && C.uint32_t(tile) < floorTileStorage.count && floorTileStorage.definitions[tile].field_58&1 != 0
}

//export nox_xxx_tileCheckRedrawMB_482570
func nox_xxx_tileCheckRedrawMB_482570(vp *nox_draw_viewport_t) C.int {
	if vp == nil || worldOwnerStorage.tile_coordinates == nil {
		return 0
	}
	left, top, right, bottom := floorVisibleCells(vp, int32(clientDisplayStorage.tile_columns), int32(clientDisplayStorage.tile_rows), worldTileCapacity411160())
	for y := max(0, top); y < bottom; y++ {
		for x := max(0, left); x < right; x++ {
			cell := mapgenFloorCell(x, y)
			if (cell.field_0&1 != 0 && floorTileNeedsTextures(cell.field_1)) ||
				(cell.field_0&2 != 0 && floorTileNeedsTextures(cell.field_6)) {
				return 1
			}
		}
	}
	return 0
}

//export nox_xxx_tileDrawImpl_4826A0
func nox_xxx_tileDrawImpl_4826A0(vp *nox_draw_viewport_t) C.int {
	if vp == nil || worldOwnerStorage.tile_coordinates == nil {
		return 0
	}
	legacyWords.dword_5d4594_3798836, legacyWords.dword_5d4594_3798840 = 0, 0
	sub_481410()
	clear(memmap.PtrT[[176]uint32](0x85B3FC, 228)[:])
	clear(memmap.PtrT[[64]uint32](0x5D4594, 2523980)[:])

	left, top, right, bottom := floorVisibleCells(vp, int32(clientDisplayStorage.tile_columns), int32(clientDisplayStorage.tile_rows), worldTileCapacity411160())
	if left < 0 || top < 0 || left >= right || top >= bottom {
		return 0
	}
	legacyWords.dword_5d4594_3798828, legacyWords.dword_5d4594_3798832 = C.uint32_t(left), C.uint32_t(top)
	legacyWords.dword_5d4594_3798820 = C.uint32_t(floorGridStep*left - 11)
	legacyWords.dword_5d4594_3798824 = C.uint32_t(floorGridStep*top - 11)
	for y := top; y < bottom; y++ {
		for x := left; x < right; x++ {
			drawFloorCell(mapgenFloorCell(x, y), floorGridStep*x-11, floorGridStep*y-11)
		}
	}
	return 0
}

//export nox_xxx_tileDrawMB_481C20_A
func nox_xxx_tileDrawMB_481C20_A(vp *nox_draw_viewport_t, world C.int) {
	if worldOwnerStorage.tile_coordinates == nil {
		return
	}
	worldX := int32(world)
	left, origin := int32(legacyWords.dword_5d4594_3798828), int32(legacyWords.dword_5d4594_3798820)
	top, originY := int32(legacyWords.dword_5d4594_3798832), int32(legacyWords.dword_5d4594_3798824)
	width, height := int32(clientDisplayStorage.tile_width), int32(clientDisplayStorage.tile_height)
	columns, rows := int32(clientDisplayStorage.tile_columns), int32(clientDisplayStorage.tile_rows)
	dx, dy := int32(legacyWords.dword_5d4594_3798836), int32(legacyWords.dword_5d4594_3798840)
	var column, screenX int32
	if worldX >= origin+23 {
		right := nox_getBackbufWidth() + worldX
		if uint32(right) <= uint32(width+origin-floorGridStep) || columns+left-1 >= worldTileCapacity411160() {
			return
		}
		if right > width+origin {
			nox_xxx_tileDrawImpl_4826A0(vp)
			return
		}
		left++
		origin += floorGridStep
		column, screenX = columns+left-2, width+origin-2*floorGridStep
		dx += floorGridStep
		if dx >= width {
			dx -= width
			dy++
			if dy >= height {
				dy -= height
			}
		}
	} else {
		if left <= 0 {
			return
		}
		if worldX < origin-23 {
			nox_xxx_tileDrawImpl_4826A0(vp)
			return
		}
		left--
		origin -= floorGridStep
		column, screenX = left, origin
		dx -= floorGridStep
		if dx < 0 {
			dx += width
			dy--
			if dy < 0 {
				dy += height
			}
		}
	}
	legacyWords.dword_5d4594_3798828, legacyWords.dword_5d4594_3798820 = C.uint32_t(left), C.uint32_t(origin)
	legacyWords.dword_5d4594_3798836, legacyWords.dword_5d4594_3798840 = C.uint32_t(dx), C.uint32_t(dy)
	sub_481410()
	for y := top; y < top+rows; y++ {
		drawFloorCell(mapgenFloorCell(column, y), screenX, originY+floorGridStep*(y-top))
	}
}

//export nox_xxx_tileDrawMB_481C20_B
func nox_xxx_tileDrawMB_481C20_B(vp *nox_draw_viewport_t, world C.int) {
	if worldOwnerStorage.tile_coordinates == nil {
		return
	}
	worldY := int32(world)
	top, origin := int32(legacyWords.dword_5d4594_3798832), int32(legacyWords.dword_5d4594_3798824)
	left, originX := int32(legacyWords.dword_5d4594_3798828), int32(legacyWords.dword_5d4594_3798820)
	height, dy := int32(clientDisplayStorage.tile_height), int32(legacyWords.dword_5d4594_3798840)
	columns, rows := int32(clientDisplayStorage.tile_columns), int32(clientDisplayStorage.tile_rows)
	var row, screenY int32
	if worldY >= origin+23 {
		bottom := nox_getBackbufHeight() + worldY
		if bottom <= origin+height || top+rows >= worldTileCapacity411160() {
			return
		}
		if bottom > origin+height+floorGridStep {
			nox_xxx_tileDrawImpl_4826A0(vp)
			return
		}
		row, screenY = top+rows, origin+height
		top++
		origin += floorGridStep
		dy += floorGridStep
		if dy >= height {
			dy -= height
		}
	} else {
		if top <= 0 {
			return
		}
		if worldY < origin-23 {
			nox_xxx_tileDrawImpl_4826A0(vp)
			return
		}
		top--
		origin -= floorGridStep
		row, screenY = top, origin
		dy -= floorGridStep
		if dy < 0 {
			dy += height
		}
	}
	legacyWords.dword_5d4594_3798832, legacyWords.dword_5d4594_3798824 = C.uint32_t(top), C.uint32_t(origin)
	legacyWords.dword_5d4594_3798840 = C.uint32_t(dy)
	sub_481410()
	for x := left; x < left+columns-1; x++ {
		drawFloorCell(mapgenFloorCell(x, row), originX+floorGridStep*(x-left), screenY)
	}
}
