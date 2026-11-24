package legacy

/*
#include "GAME4_2.h"

extern uint32_t dword_5d4594_2487576;
extern uint32_t dword_5d4594_2487580;
extern uint32_t dword_5d4594_2487584;
*/
import "C"

import (
	"math"
	"runtime"
	"unsafe"
)

type mapgenRankedRoom struct {
	mapgenRoomGeometry
	flags                  uint32
	_                      [2]uint32
	sortedNext, sortedPrev uint32
	_                      [148]byte
	visit                  uint8
	_                      [135]byte
	distance, normalized   float32
	bucket                 uint32
}

var (
	_ [52]byte  = [unsafe.Offsetof(mapgenRankedRoom{}.flags)]byte{}
	_ [64]byte  = [unsafe.Offsetof(mapgenRankedRoom{}.sortedNext)]byte{}
	_ [68]byte  = [unsafe.Offsetof(mapgenRankedRoom{}.sortedPrev)]byte{}
	_ [220]byte = [unsafe.Offsetof(mapgenRankedRoom{}.visit)]byte{}
	_ [356]byte = [unsafe.Offsetof(mapgenRankedRoom{}.distance)]byte{}
	_ [360]byte = [unsafe.Offsetof(mapgenRankedRoom{}.normalized)]byte{}
	_ [364]byte = [unsafe.Offsetof(mapgenRankedRoom{}.bucket)]byte{}
)

var mapgenRoomRanking struct {
	farthest, head *mapgenRankedRoom
	maxDistance    float32
}

func mapgenRankedRoomToken(room *mapgenRankedRoom) uint32 {
	return mapgenRegisterListedRoom((*mapgenRoomListNode)(unsafe.Pointer(room)))
}

func mapgenRankedRoomByToken(token uint32) *mapgenRankedRoom {
	return (*mapgenRankedRoom)(unsafe.Pointer(mapgenResolveListedRoom(token)))
}

func mapgenSetMaximumRoomDistance(distance float32) {
	mapgenRoomRanking.maxDistance = distance
	C.dword_5d4594_2487580 = C.uint32_t(math.Float32bits(distance))
}

func mapgenSetFarthestRoom(room *mapgenRankedRoom) {
	mapgenRoomRanking.farthest = room
	C.dword_5d4594_2487576 = C.uint32_t(mapgenRankedRoomToken(room))
}

func mapgenSetSortedRoomHead(room *mapgenRankedRoom) {
	mapgenRoomRanking.head = room
	C.dword_5d4594_2487584 = C.uint32_t(mapgenRankedRoomToken(room))
}

//export nox_mapgenResetRoomRankingNative
func nox_mapgenResetRoomRankingNative() {
	mapgenSetFarthestRoom(nil)
	mapgenSetSortedRoomHead(nil)
	mapgenSetMaximumRoomDistance(0)
}

func mapgenNextSortedRoom(room *mapgenRankedRoom) *mapgenRankedRoom {
	if room == nil {
		return nil
	}
	return mapgenRankedRoomByToken(room.sortedNext)
}

func mapgenPreviousSortedRoom(room *mapgenRankedRoom) *mapgenRankedRoom {
	if room == nil {
		return nil
	}
	return mapgenRankedRoomByToken(room.sortedPrev)
}

//export nox_mapgenRoomSortedPrevNative_525C90
func nox_mapgenRoomSortedPrevNative_525C90(room *C.nox_mapgen_readonly_room_t) *C.uint8_t {
	return (*C.uint8_t)(unsafe.Pointer(mapgenPreviousSortedRoom((*mapgenRankedRoom)(unsafe.Pointer(room)))))
}

//export nox_mapgenRoomSortedNextNative_525C90
func nox_mapgenRoomSortedNextNative_525C90(room *C.nox_mapgen_readonly_room_t) *C.uint8_t {
	return (*C.uint8_t)(unsafe.Pointer(mapgenNextSortedRoom((*mapgenRankedRoom)(unsafe.Pointer(room)))))
}

//export nox_mapgenFarthestRoomNative_5259E0
func nox_mapgenFarthestRoomNative_5259E0() *C.uint8_t {
	return (*C.uint8_t)(unsafe.Pointer(mapgenRoomRanking.farthest))
}

//export nox_mapgenMaxRoomDistanceNative_5259D0
func nox_mapgenMaxRoomDistanceNative_5259D0() C.float {
	return C.float(mapgenRoomRanking.maxDistance)
}

//export nox_mapgenSortedRoomHeadNative_525C90
func nox_mapgenSortedRoomHeadNative_525C90() *C.uint8_t {
	return (*C.uint8_t)(unsafe.Pointer(mapgenRoomRanking.head))
}

func (room *mapgenRankedRoom) eachNeighbor(visit func(*mapgenRankedRoom)) {
	links := (*mapgenRoomLinkRecord)(unsafe.Pointer(room))
	for direction := range links.links {
		row := &links.links[direction]
		for _, token := range row[:min(int(links.counts[direction]), len(row))] {
			if neighbor := mapgenRankedRoomByToken(token); neighbor != nil {
				visit(neighbor)
			}
		}
	}
}

// On ARM64 the C build fuses the width product into the sum after rounding
// the height product. The pinned Go compiler emits the same scalar operation.
//
//go:noinline
func mapgenRoomDiagonal(width, height float32) float64 {
	heightSquared := float32(height * height)
	if runtime.GOARCH == "arm64" {
		return math.Sqrt(float64(width*width + heightSquared))
	}
	return math.Sqrt(float64(float32(width*width) + heightSquared))
}

func mapgenRoomEdgeDistance(parent *mapgenRankedRoom, accumulated float32) float64 {
	if parent == nil {
		return 0
	}
	switch parent.kind {
	case 1:
		return mapgenRoomDiagonal(math.Float32frombits(uint32(parent.dimensions[0])), math.Float32frombits(uint32(parent.dimensions[1])))
	case 2, 3:
		return float64(math.Float32frombits(uint32(parent.dimensions[1])))
	case 4, 5:
		return float64(math.Float32frombits(uint32(parent.dimensions[0])))
	default:
		return float64(accumulated)
	}
}

func mapgenComputeRoomDistances(room, parent *mapgenRankedRoom, distance float32) {
	if room == nil {
		return
	}
	flags := uint8(0x9C)
	C.nox_xxx_mapGenSetFlags_5235F0(C.char(flags))
	candidate := float32(mapgenRoomEdgeDistance(parent, distance) + float64(distance))
	if room.visit == 1 && !(candidate < room.distance) {
		return
	}
	room.distance, room.visit = candidate, 1
	room.eachNeighbor(func(neighbor *mapgenRankedRoom) {
		mapgenComputeRoomDistances(neighbor, room, candidate)
	})
}

//export nox_mapgenComputeRoomDistancesNative_5259F0
func nox_mapgenComputeRoomDistancesNative_5259F0(room, parent *C.uint8_t, distance C.float) {
	mapgenComputeRoomDistances((*mapgenRankedRoom)(unsafe.Pointer(room)), (*mapgenRankedRoom)(unsafe.Pointer(parent)), float32(distance))
}

//export sub_5259F0
func sub_5259F0(room, parent C.int, distance C.float) {
	mapgenComputeRoomDistances(mapgenRankedRoomByToken(uint32(room)), mapgenRankedRoomByToken(uint32(parent)), float32(distance))
}

func mapgenFindFarthestRoom(room *mapgenRankedRoom) {
	if room == nil || room.visit == 2 {
		return
	}
	if room.kind == 1 && mapgenRoomRanking.maxDistance < room.distance {
		mapgenSetFarthestRoom(room)
		mapgenSetMaximumRoomDistance(room.distance)
	}
	room.visit = 2
	room.eachNeighbor(mapgenFindFarthestRoom)
}

//export nox_mapgenFindFarthestRoomNative_525BF0
func nox_mapgenFindFarthestRoomNative_525BF0(room *C.uint8_t) {
	mapgenFindFarthestRoom((*mapgenRankedRoom)(unsafe.Pointer(room)))
}

//export sub_525BF0
func sub_525BF0(token C.int) {
	mapgenFindFarthestRoom(mapgenRankedRoomByToken(uint32(token)))
}

func mapgenSortRooms() *mapgenRankedRoom {
	mapgenSetSortedRoomHead(nil)
	for node := mapgenRooms.head; node != nil; node = mapgenNextListedRoom(node, mapgenResolveListedRoom) {
		room := (*mapgenRankedRoom)(unsafe.Pointer(node))
		var previous *mapgenRankedRoom
		current := mapgenRoomRanking.head
		for current != nil && !(room.distance < current.distance) {
			previous = current
			current = mapgenNextSortedRoom(current)
		}
		token := mapgenRankedRoomToken(room)
		if previous == nil {
			mapgenSetSortedRoomHead(room)
		} else {
			previous.sortedNext = token
		}
		room.sortedPrev = mapgenRankedRoomToken(previous)
		room.sortedNext = mapgenRankedRoomToken(current)
		if current != nil {
			current.sortedPrev = token
		}
	}
	return mapgenRoomRanking.head
}

//export nox_mapgenSortRoomsNative_525C90
func nox_mapgenSortRoomsNative_525C90() *C.uint8_t {
	return (*C.uint8_t)(unsafe.Pointer(mapgenSortRooms()))
}

//export sub_525C90
func sub_525C90() *C.float {
	mapgenSortRooms()
	return nil
}

func mapgenClassifyRooms(root *mapgenRankedRoom) *mapgenRankedRoom {
	if root == nil {
		nox_mapgenResetRoomRankingNative()
		return nil
	}
	mapgenSetMaximumRoomDistance(0)
	mapgenSetFarthestRoom(root)
	mapgenFindFarthestRoom(root)
	farthest := mapgenRoomRanking.farthest
	farthest.flags |= 1

	count := int32(0)
	for node := mapgenRooms.head; node != nil; node = mapgenNextListedRoom(node, mapgenResolveListedRoom) {
		count++
		room := (*mapgenRankedRoom)(unsafe.Pointer(node))
		room.normalized = room.distance / mapgenRoomRanking.maxDistance
	}
	mapgenSortRooms()

	rank := int32(0)
	for room := root; room != nil; room = mapgenNextSortedRoom(room) {
		percentile := int32(100)
		if count > 1 {
			percentile = rank / (count - 1)
		}
		switch {
		case room == root:
			room.bucket = 1
		case percentile >= 90:
			room.bucket = 16
		case percentile >= 60:
			room.bucket = 8
		case percentile >= 30:
			room.bucket = 4
		default:
			room.bucket = 2
		}
		rank += 100
	}
	farthest.bucket = 32
	return farthest
}

//export nox_mapgenClassifyRoomsNative_525AF0
func nox_mapgenClassifyRoomsNative_525AF0(root *C.uint8_t) *C.uint8_t {
	return (*C.uint8_t)(unsafe.Pointer(mapgenClassifyRooms((*mapgenRankedRoom)(unsafe.Pointer(root)))))
}

//export sub_525AF0
func sub_525AF0(token C.int) *C.float {
	return (*C.float)(unsafe.Pointer(mapgenClassifyRooms(mapgenRankedRoomByToken(uint32(token)))))
}
