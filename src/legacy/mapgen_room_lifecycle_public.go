package legacy

/*
#include "GAME4_2.h"
#include "mapgen_legacy_ptr.h"

extern uint32_t dword_5d4594_2487560;
*/
import "C"

import "unsafe"

type mapgenRoomListNode struct {
	_          [56]byte
	next, prev uint32
}

type mapgenRoomList struct {
	head  *mapgenRoomListNode
	token uint32
}

var mapgenRooms mapgenRoomList

func mapgenNextListedRoom(room *mapgenRoomListNode, resolve func(uint32) *mapgenRoomListNode) *mapgenRoomListNode {
	if room == nil {
		return nil
	}
	return resolve(room.next)
}

func (list *mapgenRoomList) add(room *mapgenRoomListNode, register func(*mapgenRoomListNode) uint32) bool {
	token := register(room)
	oldHead := register(list.head)
	if token == 0 {
		return false
	}
	room.prev = 0
	room.next = oldHead
	if list.head != nil {
		list.head.prev = token
	}
	list.head, list.token = room, token
	return true
}

func (list *mapgenRoomList) remove(room *mapgenRoomListNode, resolve func(uint32) *mapgenRoomListNode, register func(*mapgenRoomListNode) uint32) {
	if prev := resolve(room.prev); prev != nil {
		prev.next = room.next
	} else {
		list.head = mapgenNextListedRoom(room, resolve)
		list.token = register(list.head)
	}
	if next := mapgenNextListedRoom(room, resolve); next != nil {
		next.prev = room.prev
	}
}

func (list *mapgenRoomList) releaseAll(resolve func(uint32) *mapgenRoomListNode, release func(*mapgenRoomListNode)) {
	for room := list.head; room != nil; {
		next := mapgenNextListedRoom(room, resolve)
		release(room)
		room = next
	}
	*list = mapgenRoomList{}
}

func mapgenReleaseRoomRectangles(head *uint32, resolve func(uint32) *mapGenOccupiedRect5227B0, release func(*mapGenOccupiedRect5227B0)) {
	for token := *head; token != 0; {
		rect := resolve(token)
		if rect == nil {
			break
		}
		token = rect.next
		release(rect)
	}
	*head = 0
}

func mapgenRegisterListedRoom(room *mapgenRoomListNode) uint32 {
	return uint32(C.nox_mapgenLegacyPtrRegister(unsafe.Pointer(room)))
}

func mapgenResolveListedRoom(token uint32) *mapgenRoomListNode {
	return (*mapgenRoomListNode)(C.nox_mapgenLegacyPtrResolve(C.uintptr_t(token)))
}

func mapgenReleaseListedRoom(room *mapgenRoomListNode) {
	raw := unsafe.Pointer(room)
	mapgenReleaseRoomRectangles(mapGenOccupiedRectHead((*C.uint8_t)(raw)), mapGenResolveOccupiedRect, mapGenReleaseOccupiedRect)
	C.nox_mapgenLegacyPtrForget(raw)
	nox_mapgen_release_memory(raw)
}

//export nox_xxx_mapGenGetTopRoom_521710
func nox_xxx_mapGenGetTopRoom_521710() unsafe.Pointer {
	return unsafe.Pointer(mapgenRooms.head)
}

//export nox_mapgenRoomNextNative_521720
func nox_mapgenRoomNextNative_521720(room *C.nox_mapgen_const_room_t) unsafe.Pointer {
	return unsafe.Pointer(mapgenNextListedRoom((*mapgenRoomListNode)(unsafe.Pointer(room)), mapgenResolveListedRoom))
}

//export nox_xxx_mapGenAddNewRoom_521730
func nox_xxx_mapGenAddNewRoom_521730(room *C.uint32_t) C.int {
	raw := unsafe.Pointer(room)
	if !mapgenRooms.add((*mapgenRoomListNode)(raw), mapgenRegisterListedRoom) {
		return 0
	}
	C.dword_5d4594_2487560 = C.uint32_t(mapgenRooms.token)
	return nox_mapgenRoomOccupyNative_521100((*C.uint8_t)(raw))
}

//export nox_mapgenRemoveRoomNative_521760
func nox_mapgenRemoveRoomNative_521760(room *C.uint8_t) C.int {
	if room == nil {
		return 0
	}
	mapgenRooms.remove((*mapgenRoomListNode)(unsafe.Pointer(room)), mapgenResolveListedRoom, mapgenRegisterListedRoom)
	C.dword_5d4594_2487560 = C.uint32_t(mapgenRooms.token)
	return nox_mapgenRoomVacateNative_521180(room)
}

//export sub_521760
func sub_521760(token C.int) C.int {
	room := mapgenResolveListedRoom(uint32(token))
	return nox_mapgenRemoveRoomNative_521760((*C.uint8_t)(unsafe.Pointer(room)))
}

//export sub_521A10
func sub_521A10(room unsafe.Pointer) {
	if room != nil {
		mapgenReleaseListedRoom((*mapgenRoomListNode)(room))
	}
}

//export nox_xxx_mapGenFreeTopRoom_521A40
func nox_xxx_mapGenFreeTopRoom_521A40() *C.uint32_t {
	mapgenRooms.releaseAll(mapgenResolveListedRoom, mapgenReleaseListedRoom)
	C.dword_5d4594_2487560 = 0
	C.nox_mapgenResetRoomRankingNative()
	return nil
}
