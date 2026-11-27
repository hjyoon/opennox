package legacy

/*
#include "GAME3_2.h"
#include "GAME4_2.h"
#include "cgo_const.h"
*/
import "C"

import (
	"math"
	"unsafe"

	"github.com/opennox/opennox/v1/common/memmap"
)

const mapgenThemeOffset = 1549796

func mapgenThemeHalfGrid() int32 {
	radius := memmap.Float32(0x5D4594, 1549860)
	return x87TruncQwordLow(float64(radius) * 0.030743772)
}

func mapgenGenerateBackdrop(theme *C.uint8_t) {
	half := mapgenThemeHalfGrid()
	size := C.int(2*half + 1)
	room := unsafe.Pointer(nox_xxx_mapGenMakeRoomStruct_521940(size, size))
	start := C.float(float64(-half) * 32.526913)
	pos := C.float2{field_0: start, field_4: start}
	nox_xxx_mapGenSetRoomPos_521880((*C.uint32_t)(room), &pos)
	for node := mapgenRooms.head; node != nil; node = mapgenNextListedRoom(node, mapgenResolveListedRoom) {
		other := (*mapgenRoomGeometry)(unsafe.Pointer(node))
		position := (*C.float2)(unsafe.Pointer(&other.position[0]))
		width := C.float(math.Float32frombits(uint32(other.dimensions[0])))
		height := C.float(math.Float32frombits(uint32(other.dimensions[1])))
		nox_mapgenAddOccupiedRectNative_521BC0((*C.uint8_t)(room), position, width, height)
	}
	nox_mapgenSelectBackdropNative_524070(theme, (*C.uint8_t)(room))
	C.nox_mapgenGenerateRoomNative_524E00(theme, (*C.uint8_t)(room))
	C.nox_mapgenPopulateDecorNative_522340(theme, (*C.uint8_t)(room))
	sub_521A10(room)
}

func mapgenBuildRooms(theme *C.uint8_t) C.int {
	nox_xxx_mapGenMkSmallRoom_4D4F40(theme)
	if nox_xxx_mapGen_InPrefab1_525D20(theme) == 0 {
		return 2
	}
	sub_4D52F0()
	if nox_xxx_mapGen_InPrefab2_5266F0(theme) == 0 {
		return 2
	}
	if nox_xxx_mapGenPlacePrefabs_526830(theme) == 0 {
		return 0
	}
	root := (*C.uint8_t)(unsafe.Pointer(mapgenResolveListedRoom(uint32(archiveAndMapgenStorage.room_head))))
	nox_mapgenComputeRoomDistancesNative_5259F0(root, nil, 0)
	nox_mapgenClassifyRoomsNative_525AF0(root)
	if memmap.Uint32(0x5D4594, 1549980) != 0 {
		mapgenGenerateBackdrop(theme)
	}
	if nox_mapgenMakeRoomsNative_524310(theme) == 0 {
		return 2
	}
	for _, roomType := range [...]bool{true, false} {
		for node := mapgenRooms.head; node != nil; node = mapgenNextListedRoom(node, mapgenResolveListedRoom) {
			room := unsafe.Pointer(node)
			if mapgenIsHall(uint32((*mapgenRoomGeometry)(room).kind)) == roomType {
				nox_xxx_mapGenSetFlags_5235F0(C.char(-100))
				C.nox_mapgenGenerateRoomNative_524E00(theme, (*C.uint8_t)(room))
			}
		}
	}
	nox_mapgenConnectRoomLinesNative_522D30(theme)
	nox_xxx_mapgen_Doors_4D4790()
	nox_mapgenConnectRoomPointsNative_522F40(theme)
	nox_mapgenFinishPopulateNative_5228B0(theme)
	return 1
}

//export nox_xxx_mapGenStep_4D44E0
func nox_xxx_mapGenStep_4D44E0() C.int {
	theme := (*C.uint8_t)(memmap.PtrOff(0x5D4594, mapgenThemeOffset))
	archiveAndMapgenStorage.room_head = 0
	sub_57C490_2(internCStr("theme"))
	sub_526C40(0)
	sub_51D100(0)
	name := (*C.nox_cgo_const_char)(nox_xxx_getRandMapName_4D4310())
	defer sub_520D50((*C.uint32_t)(unsafe.Pointer(theme)))
	if nox_mapgenReadThemeNative_51E260((*C.int)(unsafe.Pointer(theme)), name) == 0 {
		return 0
	}
	*memmap.PtrUint32(0x5D4594, 1549864) = uint32(mapgenThemeHalfGrid())
	nox_xxx_mapGenSetRngSeed_526AB0(C.uint(memmap.Uint32(0x5D4594, 1549872)))
	sub_526950()
	defer sub_526A90()
	if nox_mapgenPrepareWorkspaceNative(theme) == 0 {
		return 0
	}
	result := mapgenBuildRooms(theme)
	nox_xxx_mapGenFreeTopRoom_521A40()
	nox_xxx_mapgenFreeBuffer_521400()
	sub_520F80()
	return result
}
