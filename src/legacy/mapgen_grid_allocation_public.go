package legacy

/*
#include "GAME4_2.h"

*/
import "C"

import "unsafe"

type mapgenGridStorage struct {
	table unsafe.Pointer
	mapgenGridView
}

var mapgenGrid mapgenGridStorage

func (grid mapgenGridStorage) release() {
	if grid.table == nil {
		return
	}
	for _, row := range grid.rows {
		nox_mapgen_release_memory(unsafe.Pointer(row))
	}
	forgetMapgenPointer(grid.table)
	nox_mapgen_release_memory(grid.table)
}

//export sub_520EA0
func sub_520EA0(rawTheme *C.uint8_t) C.int {
	theme := (*mapgenThemeBounds)(unsafe.Pointer(rawTheme))
	half := uint32(theme.gridHalf)
	side := 2*half + 1

	sub_520F80()
	legacyWords.dword_5d4594_2487536 = C.uint32_t(half)
	legacyWords.dword_5d4594_2487540 = C.uint32_t(side)
	table := nox_mapgen_alloc_memory(C.size_t(side), C.size_t(unsafe.Sizeof((*mapgenGridCell)(nil))))
	if table == nil {
		return 0
	}
	token := registerMapgenPointer(table)
	if token == 0 {
		nox_mapgen_release_memory(table)
		return 0
	}
	legacyWords.dword_5d4594_2487532 = C.uint32_t(token)

	grid := mapgenGridStorage{table: table}
	grid.offset = int32(half)
	if size := int32(side); size > 0 {
		grid.rows = unsafe.Slice((**mapgenGridCell)(table), size)
		for x := range grid.rows {
			row := nox_mapgen_alloc_memory(C.size_t(size), C.size_t(unsafe.Sizeof(mapgenGridCell{})))
			if row == nil {
				grid.rows = grid.rows[:x]
				grid.release()
				legacyWords.dword_5d4594_2487532 = 0
				return 0
			}
			grid.rows[x] = (*mapgenGridCell)(row)
		}
		for x, row := range grid.rows {
			cells := unsafe.Slice(row, size)
			for y := range cells {
				cells[y].x = int32(x) - grid.offset
				cells[y].y = int32(y) - grid.offset
				cells[y].room = 0
			}
		}
	}
	mapgenGrid = grid
	return 1
}

//export sub_520F80
func sub_520F80() {
	mapgenGrid.release()
	mapgenGrid = mapgenGridStorage{}
	legacyWords.dword_5d4594_2487532 = 0
}

//export nox_mapgenRoomAtCellNative_521290
func nox_mapgenRoomAtCellNative_521290(pos *C.int2) unsafe.Pointer {
	cell := mapgenCurrentGrid().at(int32(pos.field_0), int32(pos.field_4))
	if cell == nil || cell.room == 0 {
		return nil
	}
	return resolveMapgenPointer(uintptr(cell.room))
}
