package legacy

/*
#include "GAME4_2.h"
#include "mapgen_legacy_ptr.h"

extern uint8_t** nox_mapgen_occupancy_grid;
extern uint32_t dword_5d4594_2487536;
extern uint32_t dword_5d4594_2487540;
*/
import "C"

import (
	"math"
	"unsafe"
)

type mapgenGridCell struct {
	_    [4]uint32
	room uint32
}

// Remaining C callers use the original 20-byte cell and its PE32 room token.
var (
	_ [20]byte = [unsafe.Sizeof(mapgenGridCell{})]byte{}
	_ [16]byte = [unsafe.Offsetof(mapgenGridCell{}.room)]byte{}
)

type mapgenGridView struct {
	rows   []*mapgenGridCell
	offset int32
}

func (grid mapgenGridView) at(x, y int32) *mapgenGridCell {
	x += grid.offset
	y += grid.offset
	side := int32(len(grid.rows))
	if x < 0 || y < 0 || x >= side || y >= side {
		return nil
	}
	row := grid.rows[x]
	if row == nil {
		return nil
	}
	return &unsafe.Slice(row, side)[y]
}

func mapgenCurrentGrid() mapgenGridView {
	size := int32(C.dword_5d4594_2487540)
	table := C.nox_mapgen_occupancy_grid
	if size <= 0 || table == nil {
		return mapgenGridView{}
	}
	rows := unsafe.Slice((**mapgenGridCell)(unsafe.Pointer(table)), size)
	return mapgenGridView{rows: rows, offset: int32(C.dword_5d4594_2487536)}
}

func mapgenWalkRoomCells(room *mapgenRoomGeometry, visit func(int32, int32) bool) {
	x0 := int32(mapgenGridCoordinate(math.Float32frombits(uint32(room.position[0]))))
	y0 := int32(mapgenGridCoordinate(math.Float32frombits(uint32(room.position[1]))))
	for y := int32(0); y < int32(room.cellDimensions[1]); y++ {
		for x := int32(0); x < int32(room.cellDimensions[0]); x++ {
			if visit(x0+x, y0+y) {
				return
			}
		}
	}
}

func mapgenSetRoomCells(room *mapgenRoomGeometry, token uint32, at func(int32, int32) *mapgenGridCell) int32 {
	mapgenWalkRoomCells(room, func(x, y int32) bool {
		if cell := at(x, y); cell != nil {
			cell.room = token
		}
		return false
	})
	return int32(room.cellDimensions[1])
}

func mapgenFindRoomOccupant(room *mapgenRoomGeometry, at func(int32, int32) *mapgenGridCell) uint32 {
	var token uint32
	mapgenWalkRoomCells(room, func(x, y int32) bool {
		if cell := at(x, y); cell != nil {
			token = cell.room
		}
		return token != 0
	})
	return token
}

func mapgenRoomWithinTheme(room *mapgenRoomGeometry, radius float32) bool {
	width, height := int32(room.cellDimensions[0]), int32(room.cellDimensions[1])
	startX := float64(math.Float32frombits(uint32(room.position[0])))
	y := float64(math.Float32frombits(uint32(room.position[1])))
	limit := float64(radius)
	for row := int32(0); row < height; row++ {
		x := startX
		for column := int32(0); column < width; column++ {
			if !(limit >= math.Abs(x) && limit >= math.Abs(y)) {
				return false
			}
			x = addFloat64TowardZero(x, mapgenCellSize)
		}
		y = addFloat64TowardZero(y, mapgenCellSize)
	}
	return true
}

type mapgenThemeBounds struct {
	_      [64]byte
	radius float32
}

func mapgenRoomFits(room *mapgenRoomGeometry, radius float32, at func(int32, int32) *mapgenGridCell) bool {
	return mapgenRoomWithinTheme(room, radius) && mapgenFindRoomOccupant(room, at) == 0
}

//export sub_520E60
func sub_520E60(pos *C.int2) unsafe.Pointer {
	grid := mapgenCurrentGrid()
	return unsafe.Pointer(grid.at(int32(pos.field_0), int32(pos.field_4)))
}

//export nox_mapgenRoomOccupyNative_521100
func nox_mapgenRoomOccupyNative_521100(raw *C.uint8_t) C.int {
	token := uint32(C.nox_mapgenLegacyPtrRegister(unsafe.Pointer(raw)))
	if token == 0 {
		return 0
	}
	room := (*mapgenRoomGeometry)(unsafe.Pointer(raw))
	return C.int(mapgenSetRoomCells(room, token, mapgenCurrentGrid().at))
}

//export nox_mapgenRoomVacateNative_521180
func nox_mapgenRoomVacateNative_521180(raw *C.uint8_t) C.int {
	room := (*mapgenRoomGeometry)(unsafe.Pointer(raw))
	return C.int(mapgenSetRoomCells(room, 0, mapgenCurrentGrid().at))
}

//export nox_mapgenRoomAtNative_521200
func nox_mapgenRoomAtNative_521200(raw *C.uint8_t) unsafe.Pointer {
	room := (*mapgenRoomGeometry)(unsafe.Pointer(raw))
	token := mapgenFindRoomOccupant(room, mapgenCurrentGrid().at)
	if token == 0 {
		return nil
	}
	return C.nox_mapgenLegacyPtrResolve(C.uintptr_t(token))
}

//export nox_mapgenRoomWithinThemeNative_5217A0
func nox_mapgenRoomWithinThemeNative_5217A0(rawTheme, rawRoom *C.uint8_t) C.int {
	theme := (*mapgenThemeBounds)(unsafe.Pointer(rawTheme))
	room := (*mapgenRoomGeometry)(unsafe.Pointer(rawRoom))
	return C.int(bool2int(mapgenRoomWithinTheme(room, theme.radius)))
}

//export nox_mapgenRoomFitsNative_521820
func nox_mapgenRoomFitsNative_521820(rawTheme, rawRoom *C.uint8_t) C.int {
	theme := (*mapgenThemeBounds)(unsafe.Pointer(rawTheme))
	room := (*mapgenRoomGeometry)(unsafe.Pointer(rawRoom))
	return C.int(bool2int(mapgenRoomFits(room, theme.radius, mapgenCurrentGrid().at)))
}
