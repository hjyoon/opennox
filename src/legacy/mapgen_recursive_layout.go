package legacy

/*
#include "GAME3_2.h"
#include "GAME4_2.h"
#include "cgo_const.h"

extern uint32_t dword_5d4594_1550912;
extern uint32_t dword_5d4594_1550916;
*/
import "C"

import (
	"math"
	"runtime"
	"unsafe"

	"github.com/opennox/opennox/v1/common/memmap"
)

const (
	mapgenRingDirectionOffset         = 197924
	mapgenRingSideLength              = 10
	mapgenRingBranchLength            = 8
	mapgenRingHallStep        C.float = 162.63457
	mapgenRingOrigin          C.float = 878.22662
)

func mapgenTrackExpansionRoom(room *C.float) {
	record := (*mapgenRoomLinkRecord)(unsafe.Pointer(room))
	record.expansionNext = uint32(C.dword_5d4594_1550912)
	C.dword_5d4594_1550912 = C.uint32_t(mapgenRegisterRoomLinkRecord(record))
}

func mapgenPlaceStartingRoom(room *C.float, position *C.float2) {
	raw := (*C.uint32_t)(unsafe.Pointer(room))
	nox_xxx_mapGenSetRoomPos_521880(raw, position)
	nox_xxx_mapGenAddNewRoom_521730(raw)
}

func mapgenBuildHallRing() *C.float {
	var rooms [4*mapgenRingSideLength + mapgenRingBranchLength]*C.float
	var previous, branch *C.float
	var cursor C.float2
	count := 0
	for side := range 4 {
		direction := C.int(memmap.Int32(0x587000, uintptr(mapgenRingDirectionOffset+4*side)))
		for i := range mapgenRingSideLength {
			hall := nox_mapgenMakeHallStructNative_523E30(direction, 4, 5)
			if hall == nil {
				return nil
			}
			rooms[count] = hall
			count++

			var priorKind C.uint32_t
			if previous != nil {
				priorKind = (*mapgenRoomGeometry)(unsafe.Pointer(previous)).kind
			}
			edge := C.int(-1)
			switch direction {
			case 2:
				cursor.field_4 -= mapgenRingHallStep
				edge = 1
			case 3:
				if priorKind == 4 {
					cursor.field_0 += C.float(mapgenCellSize)
					cursor.field_4 += 130.10765
				} else {
					cursor.field_4 += mapgenRingHallStep
				}
				edge = 0
			case 4:
				if previous != nil {
					cursor.field_0 += mapgenRingHallStep
					edge = 3
				}
			case 5:
				if priorKind == 3 {
					cursor.field_4 += C.float(mapgenCellSize)
				}
				cursor.field_0 -= mapgenRingHallStep
				edge = 2
				if i == 5 {
					branch = hall
				}
			}
			if edge >= 0 {
				nox_mapgenRoomLinkNative_521900((*C.uint8_t)(unsafe.Pointer(hall)), (*C.uint8_t)(unsafe.Pointer(previous)), edge)
				nox_mapgenRoomLinkNative_521900((*C.uint8_t)(unsafe.Pointer(previous)), (*C.uint8_t)(unsafe.Pointer(hall)), edge^1)
			}
			position := C.float2{field_0: cursor.field_0 - mapgenRingOrigin, field_4: cursor.field_4 - mapgenRingOrigin}
			mapgenPlaceStartingRoom(hall, &position)
			previous = hall
		}
	}
	if branch == nil {
		return previous
	}
	last, first := (*C.uint8_t)(unsafe.Pointer(previous)), (*C.uint8_t)(unsafe.Pointer(rooms[0]))
	nox_mapgenRoomLinkNative_521900(last, first, 2)
	nox_mapgenRoomLinkNative_521900(first, last, 3)

	previous = branch
	geometry := (*mapgenRoomGeometry)(unsafe.Pointer(branch))
	cursor.field_0 = C.float(math.Float32frombits(uint32(geometry.position[0])))
	cursor.field_4 = C.float(math.Float32frombits(uint32(geometry.position[1])) +
		math.Float32frombits(uint32(geometry.dimensions[1])))
	for range mapgenRingBranchLength {
		hall := nox_mapgenMakeHallStructNative_523E30(3, 4, 5)
		if hall == nil {
			return nil
		}
		rooms[count] = hall
		count++
		nox_mapgenRoomLinkNative_521900((*C.uint8_t)(unsafe.Pointer(hall)), (*C.uint8_t)(unsafe.Pointer(previous)), 0)
		nox_mapgenRoomLinkNative_521900((*C.uint8_t)(unsafe.Pointer(previous)), (*C.uint8_t)(unsafe.Pointer(hall)), 1)
		mapgenPlaceStartingRoom(hall, &cursor)
		previous = hall
		cursor.field_4 += mapgenRingHallStep
	}
	stride := C.int(max(1, count/5))
	for i := nox_xxx_mapGenRandFunc_526AC0(0, stride); i < C.int(count); i += nox_xxx_mapGenRandFunc_526AC0(1, stride) {
		mapgenTrackExpansionRoom(rooms[i])
	}
	C.dword_5d4594_1550916 = C.uint32_t(mapgenRegisterRoomLinkRecord((*mapgenRoomLinkRecord)(unsafe.Pointer(previous))))
	return previous
}

func mapgenStartingRoomY(cell int32, height float32) C.float {
	const margin = float32(97.580734)
	width := float32(cell)
	if mapgenFloatEvalMethod > 0 {
		return C.float(float64(width)*mapgenCellSize - float64(height) + float64(margin))
	}
	scaled := float32(width * mapgenCellSize)
	relative := scaled - height
	// The ARM64 C compiler contracts the product and subtraction.
	if runtime.GOARCH == "arm64" {
		relative = float32(math.FMA(float64(width), mapgenCellSize, -float64(height)))
	}
	return C.float(relative + margin)
}

//export nox_xxx_mapGenMkSmallRoom_4D4F40
func nox_xxx_mapGenMkSmallRoom_4D4F40(theme *C.uint8_t) *C.float {
	C.dword_5d4594_1550912 = 0
	if theme == nil {
		return nil
	}
	words := (*[mapgenThemeWordCount]C.int)(unsafe.Pointer(theme))
	if words[0] == 1 {
		return mapgenBuildHallRing()
	}
	room := nox_mapgenPrepareRoomNative_521990(theme)
	if room == nil {
		return nil
	}
	C.dword_5d4594_1550916 = C.uint32_t(mapgenRegisterRoomLinkRecord((*mapgenRoomLinkRecord)(unsafe.Pointer(room))))
	size := (*mapgenRoomGeometry)(unsafe.Pointer(room)).dimensions[1]
	position := C.float2{field_4: mapgenStartingRoomY(int32(words[17]), math.Float32frombits(uint32(size)))}
	mapgenPlaceStartingRoom(room, &position)
	mapgenTrackExpansionRoom(room)
	return room
}

//export sub_4D52F0
func sub_4D52F0() {
	for room := mapgenResolveRoomLinkRecord(uint32(C.dword_5d4594_1550912)); room != nil; room = mapgenResolveRoomLinkRecord(room.expansionNext) {
		raw := (*C.uint8_t)(unsafe.Pointer(room))
		size := (*mapgenRoomGeometry)(unsafe.Pointer(room)).cellDimensions
		switch room.kind {
		case 1:
			sub_4D5350(raw, 0, 0, 0, nil)
		case 2, 3:
			sub_4D5350(raw, 0, 0, C.int(size[0]), nil)
		case 4, 5:
			sub_4D5350(raw, 0, 0, C.int(size[1]), nil)
		}
	}
}

//export sub_4D5350
func sub_4D5350(room *C.uint8_t, depth, hallDepth, hallSize C.int, avoid *C.uint8_t) C.int {
	depth++
	if depth >= C.int(memmap.Int32(0x5D4594, 1549868)) {
		return 0
	}
	nox_xxx_mapGenSetFlags_5235F0(-101)
	if (*mapgenRoomGeometry)(unsafe.Pointer(room)).kind == 1 {
		return C.nox_xxx_mapGenFillRoom_4D53B0(room, depth, hallDepth, hallSize, avoid)
	}
	return C.sub_4D5630(room, depth, hallDepth, hallSize, avoid)
}

//export sub_4D5D20
func sub_4D5D20(raw *C.nox_cgo_const_uint32) C.int {
	room := (*mapgenRoomGeometry)(unsafe.Pointer(raw))
	width, height := uint32(room.cellDimensions[0]), uint32(room.cellDimensions[1])
	if room.kind == 2 || room.kind == 3 {
		if width >= height {
			return 1
		}
	} else if height >= width {
		return 1
	}
	if nox_xxx_mapGenRandFunc_526AC0(1, 100) <= C.int(memmap.Int32(0x5D4594, 1549820)) {
		nox_xxx_mapGenRandFunc_526AC0(1, 100)
		return 32
	}
	if nox_xxx_mapGenRandFunc_526AC0(1, 100) <= C.int(memmap.Int32(0x5D4594, 1549824)) {
		return 1
	}
	if nox_xxx_mapGenRandFunc_526AC0(1, 100) >= 50 {
		return 4
	}
	return 2
}
