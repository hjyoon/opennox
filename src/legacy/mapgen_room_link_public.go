package legacy

/*
#include "GAME4_2.h"
#include "mapgen_legacy_ptr.h"
*/
import "C"

import (
	"unsafe"

	"github.com/opennox/opennox/v1/common/memmap"
)

type mapgenRoomLinkRecord struct {
	kind          uint32
	_             [80]byte
	expansionNext uint32
	links         [4][8]uint32
	counts        [4]uint8
}

var _ [84]byte = [unsafe.Offsetof(mapgenRoomLinkRecord{}.expansionNext)]byte{}

func mapgenRoomHasTypeOneNeighbor(room *mapgenRoomLinkRecord, direction int, resolve func(uint32) *mapgenRoomLinkRecord) bool {
	if room == nil || direction < 0 || direction >= len(room.links) {
		return false
	}
	links := &room.links[direction]
	count := min(int(room.counts[direction]), len(links))
	for _, token := range links[:count] {
		if token == 0 {
			continue
		}
		if neighbor := resolve(token); neighbor != nil && neighbor.kind == 1 {
			return true
		}
	}
	return false
}

func mapgenAddRoomLink(room, neighbor *mapgenRoomLinkRecord, direction int, register func(*mapgenRoomLinkRecord) uint32) bool {
	if room == nil || direction < 0 || direction >= len(room.links) {
		return false
	}
	links := &room.links[direction]
	count := room.counts[direction]
	if int(count) >= len(links) {
		return false
	}
	var token uint32
	if neighbor != nil {
		token = register(neighbor)
		if token == 0 {
			return false
		}
	}
	links[count] = token
	room.counts[direction]++
	return true
}

func mapgenLinkRoomsBoth(room, neighbor *mapgenRoomLinkRecord, direction int, register func(*mapgenRoomLinkRecord) uint32, opposite func(int) int) bool {
	if room == nil || neighbor == nil || direction < 0 || direction >= len(room.links) {
		return false
	}
	mapgenAddRoomLink(room, neighbor, direction, register)
	return mapgenAddRoomLink(neighbor, room, opposite(direction), register)
}

func mapgenRegisterRoomLinkRecord(room *mapgenRoomLinkRecord) uint32 {
	return uint32(C.nox_mapgenLegacyPtrRegister(unsafe.Pointer(room)))
}

func mapgenResolveRoomLinkRecord(token uint32) *mapgenRoomLinkRecord {
	return (*mapgenRoomLinkRecord)(C.nox_mapgenLegacyPtrResolve(C.uintptr_t(token)))
}

func mapgenOppositeRoomDirection(direction int) int {
	return int(int32(memmap.Uint32(0x587000, uintptr(254952+4*direction))))
}

//export nox_mapgenRoomHasTypeOneNeighborNative_5218B0
func nox_mapgenRoomHasTypeOneNeighborNative_5218B0(room *C.uint8_t, direction C.int) C.int {
	return C.int(bool2int(mapgenRoomHasTypeOneNeighbor((*mapgenRoomLinkRecord)(unsafe.Pointer(room)), int(direction), mapgenResolveRoomLinkRecord)))
}

//export nox_mapgenRoomLinkNative_521900
func nox_mapgenRoomLinkNative_521900(room, neighbor *C.uint8_t, direction C.int) C.int {
	return C.int(bool2int(mapgenAddRoomLink((*mapgenRoomLinkRecord)(unsafe.Pointer(room)),
		(*mapgenRoomLinkRecord)(unsafe.Pointer(neighbor)), int(direction), mapgenRegisterRoomLinkRecord)))
}

//export nox_mapgenRoomLinkBothNative_521A70
func nox_mapgenRoomLinkBothNative_521A70(room, neighbor *C.uint8_t, direction C.int) C.int {
	return C.int(bool2int(mapgenLinkRoomsBoth((*mapgenRoomLinkRecord)(unsafe.Pointer(room)),
		(*mapgenRoomLinkRecord)(unsafe.Pointer(neighbor)), int(direction), mapgenRegisterRoomLinkRecord, mapgenOppositeRoomDirection)))
}

//export sub_521A70
func sub_521A70(roomToken, neighborToken, direction C.int) C.int {
	return C.int(bool2int(mapgenLinkRoomsBoth(mapgenResolveRoomLinkRecord(uint32(roomToken)),
		mapgenResolveRoomLinkRecord(uint32(neighborToken)), int(direction), mapgenRegisterRoomLinkRecord, mapgenOppositeRoomDirection)))
}

//export sub_523960
func sub_523960(direction C.int) C.int {
	return C.int(mapgenOppositeRoomDirection(int(direction)))
}
