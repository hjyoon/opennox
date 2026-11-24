package legacy

/*
#include "GAME4_2.h"
*/
import "C"

import (
	"math"
	"unsafe"

	"github.com/opennox/opennox/v1/common/memmap"
)

//export sub_526260
func sub_526260(raw *C.uint8_t, destination *C.float) C.longlong {
	room := (*mapgenRoomGeometry)(unsafe.Pointer(raw))
	next := memmap.PtrUint32(0x5D4594, 2487608)
	index := memmap.Uint32(0x5D4594, 2487588+4*uintptr(*next))
	slots := mapgenPrefabStartCells()
	start := [2]float32{math.Float32frombits(*slots[2*index]), math.Float32frombits(*slots[2*index+1])}
	*next++
	output := (*[2]C.float)(unsafe.Pointer(destination))
	var cell int64
	for axis, position := range start {
		extent := float64(mapgenRoomFloat(room.dimensions[axis]))
		centered := float32(float64(position) - 0.5*extent)
		cell = x87TruncQword(float64(centered) * 0.030743772)
		output[axis] = C.float(float64(int32(cell)) * mapgenHallCellSize)
	}
	return C.longlong(cell)
}

func mapgenPrefabRoomCells(size float32, maxMarkerCount uint32) C.int {
	grid := x87TruncQwordLow(float64(size)*0.030743772 + 0.5)
	return C.int(4*maxMarkerCount + uint32(grid))
}

//export nox_xxx_mapGenPrefabMkRoom_526100
func nox_xxx_mapGenPrefabMkRoom_526100(theme, raw *C.uint8_t) C.int {
	if sub_5262F0(theme, raw) == 0 {
		return 0
	}
	prefab := (*mapgenThemePrefabRecord)(unsafe.Pointer(raw))
	width := mapgenPrefabRoomCells(prefab.width, prefab.maxMarkerCount)
	height := mapgenPrefabRoomCells(prefab.height, prefab.maxMarkerCount)
	room := unsafe.Pointer(nox_xxx_mapGenMakeRoomStruct_521940(width, height))
	if room == nil {
		return 0
	}
	prefab.room = mapgenRegisterListedRoom((*mapgenRoomListNode)(room))
	native := (*C.uint8_t)(room)
	var position C.float2
	sub_526260(native, &position.field_0)
	nox_xxx_mapGenSetRoomPos_521880((*C.uint32_t)(room), &position)

	blocker := nox_mapgenRoomAtNative_521200(native)
	canPlace := blocker == nil || sub_5212B0(native, (*C.uint32_t)(blocker)) != 0
	if nox_mapgenRoomWithinThemeNative_5217A0(theme, native) == 0 || !canPlace {
		sub_521A10(room)
		prefab.room = 0
		return 0
	}

	nox_xxx_mapGenAddNewRoom_521730((*C.uint32_t)(room))
	record := (*mapgenRankedRoom)(room)
	record.flags |= 2
	relative := 2 * prefab.maxMarkerCount
	for i := range prefab.markers {
		marker := &prefab.markers[i]
		if marker.present != 0 {
			marker.x += int32(relative + uint32(record.grid[0]))
			marker.y += int32(relative + uint32(record.grid[1]))
		}
	}
	return 1
}
