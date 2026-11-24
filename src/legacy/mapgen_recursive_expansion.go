package legacy

/*
#include "GAME3_2.h"
#include "GAME4_2.h"
#include "cgo_const.h"
*/
import "C"

import (
	"unsafe"

	"github.com/opennox/opennox/v1/common/memmap"
)

func mapgenExpansionTheme() *C.uint8_t {
	return (*C.uint8_t)(memmap.PtrOff(0x5D4594, mapgenThemeOffset))
}

func mapgenExpansionGeometry(room *C.uint8_t) (position, dimensions [2]C.float) {
	geometry := (*mapgenRoomGeometry)(unsafe.Pointer(room))
	position = *(*[2]C.float)(unsafe.Pointer(&geometry.position))
	dimensions = *(*[2]C.float)(unsafe.Pointer(&geometry.dimensions))
	return position, dimensions
}

func mapgenPositionBeyond(kind C.uint32_t, parent, child *C.uint8_t) C.float2 {
	start, size := mapgenExpansionGeometry(parent)
	_, childSize := mapgenExpansionGeometry(child)
	point := C.float2{field_0: start[0], field_4: start[1]}
	switch kind {
	case 2:
		point.field_4 -= childSize[1]
	case 3:
		point.field_4 += size[1]
	case 4:
		point.field_0 += size[0]
	case 5:
		point.field_0 -= childSize[0]
	}
	return point
}

func mapgenAlignEnd(origin, size, childSize C.float) C.float {
	if mapgenFloatEvalMethod > 0 {
		return C.float(float64(origin) + float64(size) - float64(childSize))
	}
	return (origin + size) - childSize
}

func mapgenPositionAlongSide(kind C.uint32_t, parent, hall *C.uint8_t, left bool) C.float2 {
	start, size := mapgenExpansionGeometry(parent)
	_, newSize := mapgenExpansionGeometry(hall)
	x, y := start[0], start[1]
	switch kind {
	case 2:
		x += size[0]
		if left {
			x = start[0] - newSize[0]
		}
	case 3:
		x += size[0]
		if !left {
			x = start[0] - newSize[0]
		}
		y = mapgenAlignEnd(y, size[1], newSize[1])
	case 4:
		x = mapgenAlignEnd(x, size[0], newSize[0])
		y += size[1]
		if left {
			y = start[1] - newSize[1]
		}
	case 5:
		y -= newSize[1]
		if left {
			y = start[1] + size[1]
		}
	}
	return C.float2{field_0: x, field_4: y}
}

func mapgenFinishHall(hall, parent, avoid *C.uint8_t, depth, hallDepth, width C.int, fromRoom bool) bool {
	raw := unsafe.Pointer(hall)
	existing := (*mapgenRankedRoom)(nox_mapgenRoomAtNative_521200(hall))
	if existing != nil {
		validKind := existing.kind == 1
		if fromRoom {
			validKind = !mapgenIsHall(uint32(existing.kind))
		}
		if !validKind || existing.flags&2 != 0 || unsafe.Pointer(existing) == unsafe.Pointer(avoid) ||
			nox_xxx_mapGenRandFunc_526AC0(1, 100) > C.int(memmap.Int32(0x5D4594, mapgenThemeOffset+4*mapgenThemeMergeRateIndex)) ||
			nox_mapgenAdjustHallNative_523A10(hall, (*C.nox_mapgen_readonly_float_t)(unsafe.Pointer(existing))) == 0 {
			sub_521A10(raw)
			return false
		}
		nox_xxx_mapGenAddNewRoom_521730((*C.uint32_t)(raw))
		direction := sub_523920(C.int((*mapgenRoomGeometry)(raw).kind))
		nox_mapgenRoomLinkBothNative_521A70(hall, (*C.uint8_t)(unsafe.Pointer(existing)), direction)
		return true
	}

	nox_xxx_mapGenAddNewRoom_521730((*C.uint32_t)(raw))
	if fromRoom {
		avoid = parent
	}
	if sub_4D5350(hall, depth, hallDepth, width, avoid) != 0 {
		return true
	}
	nox_mapgenRemoveRoomNative_521760(hall)
	sub_521A10(raw)
	return false
}

//export nox_xxx_mapGenFillRoom_4D53B0
func nox_xxx_mapGenFillRoom_4D53B0(room *C.uint8_t, depth, hallDepth, hallWidth C.int, avoid *C.uint8_t) C.int {
	var available [4]bool
	anyDirection := false
	for dir := range available {
		available[dir] = nox_mapgenRoomHasTypeOneNeighborNative_5218B0(room, C.int(dir)) == 0
		anyDirection = anyDirection || available[dir]
	}
	if !anyDirection {
		return 1
	}
	theme := mapgenExpansionTheme()
	direction := int(nox_xxx_mapGenRandFunc_526AC0(0, 3))
	for attempts := 0; attempts < 8; {
		direction = (direction + 1) % len(available)
		if !available[direction] {
			continue
		}
		attempts++
		mean := memmap.Int32(0x5D4594, 1549808)
		spread := memmap.Int32(0x5D4594, 1549812)
		width := nox_xxx_mapGenRandFunc_526AC0(C.int(mean-spread), C.int(mean+spread))
		kind := C.uint32_t(direction + 2)
		hall := (*C.uint8_t)(unsafe.Pointer(nox_mapgenMakeHallNative_523EC0(theme, C.int(kind), width)))
		if hall == nil {
			return 0
		}
		positioned := mapgenPositionBeyond(kind, room, hall)
		parent := (*C.nox_mapgen_readonly_room_t)(unsafe.Pointer(room))
		child := (*C.nox_mapgen_readonly_room_t)(unsafe.Pointer(hall))
		if kind == 2 || kind == 3 {
			positioned.field_0 = nox_mapgenRoomRandPosXNative_521B00(parent, child)
		} else {
			positioned.field_4 = nox_mapgenRoomRandPosYNative_521B30(parent, child)
		}
		nox_xxx_mapGenSetRoomPos_521880((*C.uint32_t)(unsafe.Pointer(hall)), &positioned)
		if nox_mapgenRoomWithinThemeNative_5217A0(theme, hall) == 0 {
			sub_521A10(unsafe.Pointer(hall))
			continue
		}
		edge := sub_523920(C.int(kind))
		nox_mapgenRoomLinkNative_521900(hall, room, sub_523960(edge))
		if mapgenFinishHall(hall, room, avoid, depth, 1, width, true) {
			nox_mapgenRoomLinkNative_521900(room, hall, edge)
		}
	}
	return 1
}

func mapgenGrowRoomAtHall(theme, hall *C.uint8_t, kind C.uint32_t, depth C.int) bool {
	for range 10 {
		room := (*C.uint8_t)(unsafe.Pointer(nox_mapgenPrepareRoomNative_521990(theme)))
		if room == nil {
			return false
		}
		point := mapgenPositionBeyond(kind, hall, room)
		child := (*C.nox_mapgen_readonly_room_t)(unsafe.Pointer(room))
		parent := (*C.nox_mapgen_readonly_room_t)(unsafe.Pointer(hall))
		if kind == 2 || kind == 3 {
			point.field_0 = nox_mapgenRoomAttachPosXNative_521B60(child, parent)
		} else {
			point.field_4 = nox_mapgenRoomAttachPosYNative_521B90(child, parent)
		}
		raw := unsafe.Pointer(room)
		nox_xxx_mapGenSetRoomPos_521880((*C.uint32_t)(raw), &point)
		if nox_mapgenRoomFitsNative_521820(theme, room) == 0 {
			sub_521A10(raw)
			continue
		}
		nox_xxx_mapGenAddNewRoom_521730((*C.uint32_t)(raw))
		nox_mapgenRoomLinkBothNative_521A70(hall, room, sub_523920(C.int(kind)))
		sub_4D5350(room, depth, 0, 0, room)
		return true
	}
	return false
}

func mapgenExtendHall(theme, parent, hall, avoid *C.uint8_t, position C.float2, backEdge, depth, hallDepth, width C.int) bool {
	nox_xxx_mapGenSetRoomPos_521880((*C.uint32_t)(unsafe.Pointer(hall)), &position)
	nox_mapgenRoomLinkNative_521900(hall, parent, backEdge)
	if nox_mapgenRoomWithinThemeNative_5217A0(theme, hall) == 0 {
		sub_521A10(unsafe.Pointer(hall))
		return false
	}
	if !mapgenFinishHall(hall, parent, avoid, depth, hallDepth, width, false) {
		return false
	}
	nox_mapgenRoomLinkNative_521900(parent, hall, sub_523960(backEdge))
	return true
}

//export sub_4D5630
func sub_4D5630(parent *C.uint8_t, depth, hallDepth, width C.int, avoid *C.uint8_t) C.int {
	kind := (*mapgenRoomGeometry)(unsafe.Pointer(parent)).kind
	theme := mapgenExpansionTheme()
	if hallDepth >= C.int(memmap.Int32(0x5D4594, 1549816)) {
		return C.int(bool2int(mapgenGrowRoomAtHall(theme, parent, kind, depth)))
	}
	expansion := sub_4D5D20((*C.nox_cgo_const_uint32)(unsafe.Pointer(parent)))
	if expansion == 1 {
		return C.int(bool2int(mapgenGrowRoomAtHall(theme, parent, kind, depth)))
	}

	anyTurn := false
	for _, turn := range [...]struct {
		left     bool
		active   C.int
		required C.int
		kinds    uintptr
	}{
		{true, 2 | 8 | 32 | 64, 2 | 8, 197812},
		{false, 4 | 16 | 32 | 64, 4 | 16, 197836},
	} {
		if expansion&turn.active == 0 {
			continue
		}
		turnKind := C.int(memmap.Int32(0x587000, turn.kinds+4*uintptr(kind)))
		hall := (*C.uint8_t)(unsafe.Pointer(nox_mapgenMakeHallNative_523EC0(theme, turnKind, width)))
		if hall == nil {
			return 0
		}
		back := sub_523970(C.int(kind))
		if turn.left {
			back = sub_523960(back)
		}
		position := mapgenPositionAlongSide(kind, parent, hall, turn.left)
		connected := mapgenExtendHall(theme, parent, hall, avoid, position, back, depth, hallDepth+1, width)
		if !connected && expansion&turn.required != 0 {
			return 0
		}
		anyTurn = anyTurn || connected
	}
	if !anyTurn {
		return 0
	}
	if expansion&(8|16|32) == 0 {
		return 1
	}
	hall := (*C.uint8_t)(unsafe.Pointer(nox_mapgenMakeHallNative_523EC0(theme, C.int(kind), width)))
	if hall == nil {
		return 0
	}
	position := mapgenPositionBeyond(kind, parent, hall)
	back := sub_523960(sub_523920(C.int(kind)))
	mapgenExtendHall(theme, parent, hall, avoid, position, back, depth, hallDepth+1, width)
	return 1
}
