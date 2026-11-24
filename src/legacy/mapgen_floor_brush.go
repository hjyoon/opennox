package legacy

/*
#include <stdlib.h>
#include "GAME1_1.h"
#include "GAME4_2.h"
#include "GAME4_3.h"
#include "mapgen_legacy_ptr.h"

extern obj_5D4594_2650668_t** ptr_5D4594_2650668;
extern void* dword_5d4594_588084;
extern uint32_t dword_5d4594_3835348, dword_5d4594_3835352;
*/
import "C"

import (
	"unsafe"

	"github.com/opennox/opennox/v1/common/memmap"
)

func mapgenFloorCell(x, y int32) *C.obj_5D4594_2650668_t {
	row := worldTileRow411160(unsafe.Pointer(C.ptr_5D4594_2650668), x)
	return (*C.obj_5D4594_2650668_t)(unsafe.Add(row, uintptr(y)*C.sizeof_obj_5D4594_2650668_t))
}

func mapgenBrushVariation(brush *mapgenFloorBrush, x, y, layer int32) C.int {
	if brush.pattern != 0 && brush.kind == 0 {
		brush.anchorX, brush.anchorY, brush.kind = x, y, uint8(layer)
		return 0
	}
	parity := int32(bool2int(layer != 1))
	if brush.pattern != 0 {
		x -= brush.anchorX
		y -= brush.anchorY
		switch {
		case int32(brush.kind) == layer:
			parity = 0
		case brush.kind == 1:
			parity = 1
		default:
			parity = -1
		}
	}
	def := &floorTileStorage.definitions[brush.tileID]
	width, height := int32(def.field_52), int32(def.field_53)
	column, row := (y+x)%width, (parity+y-x)%height
	if column < 0 {
		column += width
	}
	if row < 0 {
		row += height
	}
	if memmap.Uint32(mapgenTileStateBase, mapgenTileModeOffset) == 1 {
		return C.int(C.dword_5d4594_3835348)
	}
	return C.int(column + row*width)
}

//export nox_floorBrushApplyNative_51DA70
func nox_floorBrushApplyNative_51DA70(x, y C.int, native *C.nox_floor_brush_t, layer, merge C.int) C.int {
	brush := (*mapgenFloorBrush)(unsafe.Pointer(native))
	tile, variation := C.int(255), C.int(0)
	if brush != nil {
		tile = C.int(brush.tileID)
		variation = mapgenBrushVariation(brush, int32(x), int32(y), int32(layer))
	}
	if layer != 1 && layer != 2 {
		return 0
	}
	cell := mapgenFloorCell(int32(x), int32(y))
	baseTile, baseVariation, head := &cell.field_1, &cell.field_2, &cell.field_5
	if layer == 2 {
		baseTile, baseVariation, head = &cell.field_6, &cell.field_7, &cell.field_10
	}
	if C.dword_5d4594_3835352 == 1 {
		edge, edgeVariation := C.int(255), C.int(0)
		if brush != nil {
			edge, edgeVariation = C.int(brush.edgeID), C.int(brush.edgeVariation)
		}
		return nox_xxx_tile_543C50(C.uint32_t(*baseTile), head, tile, variation, edge, edgeVariation, merge)
	}
	if *baseTile == tile && *baseVariation == variation {
		return 0
	}
	*baseTile, *baseVariation = tile, variation
	if tile == 255 {
		nox_xxx_tileFreeTile_422200(head)
	}
	bit := C.uint8_t(layer)
	if brush != nil {
		cell.field_0 |= bit
	} else {
		cell.field_0 &^= bit
	}
	return 1
}

//export sub_51DA70
func sub_51DA70(x, y, brushToken, layer, merge C.int) C.int {
	brush := (*C.nox_floor_brush_t)(C.nox_mapgenLegacyPtrResolve(C.uintptr_t(uint32(brushToken))))
	return nox_floorBrushApplyNative_51DA70(x, y, brush, layer, merge)
}

//export nox_xxx_tileListAddNewSubtile_422160
func nox_xxx_tileListAddNewSubtile_422160(tile, variation, edge, edgeVariation C.int) *C.nox_tile_list_node_t {
	node := (*C.nox_tile_list_node_t)(C.dword_5d4594_588084)
	if node == nil {
		const batch = 10
		node = (*C.nox_tile_list_node_t)(C.calloc(batch, C.sizeof_nox_tile_list_node_t))
		if node == nil {
			return nil
		}
		nodes := unsafe.Slice(node, batch)
		for i := range nodes[:batch-1] {
			nodes[i].next = &nodes[i+1]
		}
	}
	C.dword_5d4594_588084 = unsafe.Pointer(node.next)
	node.field_0, node.field_1, node.field_2, node.field_3 = tile, variation, edge, edgeVariation
	node.next = nil
	return node
}

//export nox_xxx_tileFreeTileOne_4221E0
func nox_xxx_tileFreeTileOne_4221E0(ptr unsafe.Pointer) *C.nox_tile_list_node_t {
	node := (*C.nox_tile_list_node_t)(ptr)
	node.next = (*C.nox_tile_list_node_t)(C.dword_5d4594_588084)
	C.dword_5d4594_588084 = ptr
	return node
}

//export nox_xxx_mapSwitchLevel_4D12E0_tileFree
func nox_xxx_mapSwitchLevel_4D12E0_tileFree() {
	count := int(worldTileCapacity411160())
	rows := unsafe.Slice(C.ptr_5D4594_2650668, count)
	// The original traverses the grid column-first. Preserve the order
	// in which the two layers are returned to the shared subtile pool.
	for column := range count {
		offset := uintptr(column) * C.sizeof_obj_5D4594_2650668_t
		for _, row := range rows {
			tile := (*C.obj_5D4594_2650668_t)(unsafe.Add(unsafe.Pointer(row), offset))
			tile.field_0 = 0
			tile.field_1 = 255
			tile.field_6 = 255
			nox_xxx_tileFreeTile_422200(&tile.field_5)
			nox_xxx_tileFreeTile_422200(&tile.field_10)
		}
	}
}

//export nox_xxx_tileFreeTile_422200
func nox_xxx_tileFreeTile_422200(head **C.nox_tile_list_node_t) {
	for node := *head; node != nil; {
		next := node.next
		nox_xxx_tileFreeTileOne_4221E0(unsafe.Pointer(node))
		node = next
	}
	*head = nil
}

func sameFloorSubtile(node *C.nox_tile_list_node_t, tile, variation, edge C.int) bool {
	return node.field_0 == tile && node.field_1 == variation && node.field_2 == edge
}

//export nox_xxx_tile_543C50
func nox_xxx_tile_543C50(baseTile C.uint32_t, head **C.nox_tile_list_node_t, tile, variation, edge, edgeVariation, merge C.int) C.int {
	link := head
	if tile == 255 || edge == 255 {
		for *link != nil && (*link).next != nil {
			link = &(*link).next
		}
		if *link != nil {
			nox_xxx_tileFreeTileOne_4221E0(unsafe.Pointer(*link))
			*link = nil
		}
		return 1
	}
	if baseTile == 255 {
		return 1
	}
	if merge == 0 {
		for node := *link; node != nil; node = *link {
			if sameFloorSubtile(node, tile, variation, edge) && node.field_3 == edgeVariation {
				return 0
			}
			link = &node.next
		}
		*link = nox_xxx_tileListAddNewSubtile_422160(tile, variation, edge, edgeVariation)
		return 1
	}

	matched := false
	for node := *link; node != nil; node = *link {
		if sameFloorSubtile(node, tile, variation, edge) && sub_543E60(node, edgeVariation) != 0 {
			matched = true
		}
		link = &node.next
	}
	if !matched {
		mapped := nox_xxx_mapGenEdge_543EB0(edge, edgeVariation)
		*link = nox_xxx_tileListAddNewSubtile_422160(tile, variation, edge, mapped)
		return 1
	}
	for removed := true; removed; {
		removed = false
		for node := *head; node != nil; node = node.next {
			if !sameFloorSubtile(node, tile, variation, edge) {
				continue
			}
			link = &node.next
			for other := *link; other != nil; other = *link {
				if sameFloorSubtile(other, tile, variation, edge) && sub_543E60(node, sub_411490(other.field_2, other.field_3)) != 0 {
					removed = true
					*link = other.next
					nox_xxx_tileFreeTileOne_4221E0(unsafe.Pointer(other))
				} else {
					link = &other.next
				}
			}
		}
	}
	return 1
}

//export sub_543E60
func sub_543E60(node *C.nox_tile_list_node_t, edgeVariation C.int) C.int {
	index := edgeVariation + 12*sub_411490(node.field_2, node.field_3)
	variation := memmap.Uint32(0x587000, uintptr(282736+4*index))
	if variation == 255 {
		return 0
	}
	if mapped := nox_xxx_mapGenEdge_543EB0(node.field_2, C.int(variation)); mapped != node.field_3 {
		node.field_3 = mapped
	}
	return 1
}

//export nox_xxx_mapGenEdge_543EB0
func nox_xxx_mapGenEdge_543EB0(edge, variation C.int) C.int {
	size := memmap.Slice(0x85B3FC, uintptr(28696+60*edge))[:2]
	x, y := int32(size[0]), int32(size[1])
	if x == 3 && y == 3 || variation == 0 {
		return variation
	}
	switch variation {
	case 1:
		return C.int(nox_common_randomInt_415FA0(1, x-2))
	case 2:
		return C.int(x - 1)
	case 3:
		return C.int(x + 2*nox_common_randomInt_415FA0(0, y-3))
	case 4:
		return C.int(x + 2*nox_common_randomInt_415FA0(0, y-3) + 1)
	case 5:
		return C.int(x + 2*y - 4)
	case 6:
		return C.int(nox_common_randomInt_415FA0(1, x-2) + x + 2*y - 4)
	case 7:
		return C.int(2*(x+y) - 5)
	default:
		return variation + C.int(2*(x+y)-12)
	}
}
