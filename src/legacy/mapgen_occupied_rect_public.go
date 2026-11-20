package legacy

/*
#include "GAME4_2.h"
#include "mapgen_legacy_ptr.h"
*/
import "C"

import (
	"unsafe"

	"github.com/opennox/libs/types"
)

const mapGenOccupiedRectHeadOffset = 368

type mapGenOccupiedRect5227B0 struct {
	transient   uint32
	bounds      [4]float32
	objectIndex int32
	next        uint32
}

func mapGenAddOccupiedRect521BC0(head *uint32, pos *types.Pointf, width, height float32,
	allocate func() (*mapGenOccupiedRect5227B0, uint32)) *mapGenOccupiedRect5227B0 {
	rect, token := allocate()
	if rect == nil {
		return nil
	}
	*rect = mapGenOccupiedRect5227B0{
		bounds: [4]float32{
			pos.X, pos.Y,
			storeFloat32TowardZero(addFloat64TowardZero(float64(width), float64(pos.X))),
			storeFloat32TowardZero(addFloat64TowardZero(float64(height), float64(pos.Y))),
		},
		next: *head,
	}
	*head = token
	return rect
}

func mapGenClearTransientOccupiedRects521C10(head *uint32, resolve func(uint32) *mapGenOccupiedRect5227B0,
	release func(*mapGenOccupiedRect5227B0)) {
	link := head
	for token := *head; token != 0; {
		rect := resolve(token)
		if rect == nil {
			return
		}
		next := rect.next
		if rect.transient == 1 {
			*link = next
			release(rect)
		} else {
			link = &rect.next
		}
		token = next
	}
}

func mapGenOccupiedRectInterior(bounds [4]float32) [4]float64 {
	const border = 0.5
	return [4]float64{
		addFloat64TowardZero(float64(bounds[0]), border),
		addFloat64TowardZero(float64(bounds[1]), border),
		addFloat64TowardZero(float64(bounds[2]), -border),
		addFloat64TowardZero(float64(bounds[3]), -border),
	}
}

func mapGenOccupiedRectInside521EB0(room, rect [4]float32) bool {
	inner := mapGenOccupiedRectInterior(rect)
	// The original upper-bound branches accept unordered comparisons.
	return inner[0] >= float64(room[0]) && !(inner[2] > float64(room[2])) &&
		inner[1] >= float64(room[1]) && !(inner[3] > float64(room[3]))
}

func mapGenOccupiedRectsIntersect521F10(token uint32, candidate *mapGenOccupiedRect5227B0,
	resolve func(uint32) *mapGenOccupiedRect5227B0) *mapGenOccupiedRect5227B0 {
	var edges [4]float32
	for i, value := range mapGenOccupiedRectInterior(candidate.bounds) {
		edges[i] = storeFloat32TowardZero(value)
	}
	for token != 0 {
		node := resolve(token)
		if node == nil {
			return nil
		}
		if node != candidate {
			other := mapGenOccupiedRectInterior(node.bounds)
			// The original keeps the other X edges in x87 registers and
			// spills its Y edges and all candidate edges to binary32.
			if !(float64(edges[0]) >= other[2]) && float64(edges[2]) > other[0] &&
				!(edges[1] >= storeFloat32TowardZero(other[3])) && edges[3] > storeFloat32TowardZero(other[1]) {
				return node
			}
		}
		token = node.next
	}
	return nil
}

func mapGenOccupiedRectHead(room *C.uint8_t) *uint32 {
	return (*uint32)(unsafe.Add(unsafe.Pointer(room), mapGenOccupiedRectHeadOffset))
}

func mapGenResolveOccupiedRect(token uint32) *mapGenOccupiedRect5227B0 {
	return (*mapGenOccupiedRect5227B0)(C.nox_mapgenLegacyPtrResolve(C.uintptr_t(token)))
}

func mapGenAllocateOccupiedRect() (*mapGenOccupiedRect5227B0, uint32) {
	raw := C.calloc(1, C.sizeof_nox_mapgen_occupied_rect_legacy)
	if raw == nil {
		return nil, 0
	}
	token := uint32(C.nox_mapgenLegacyPtrRegister(raw))
	if token == 0 {
		C.free(raw)
		return nil, 0
	}
	return (*mapGenOccupiedRect5227B0)(raw), token
}

//export nox_mapgenAddOccupiedRectNative_521BC0
func nox_mapgenAddOccupiedRectNative_521BC0(room *C.uint8_t, pos *C.float2, width, height C.float) *C.nox_mapgen_occupied_rect_legacy {
	if room == nil || pos == nil {
		return nil
	}
	rect := mapGenAddOccupiedRect521BC0(mapGenOccupiedRectHead(room), (*types.Pointf)(unsafe.Pointer(pos)),
		float32(width), float32(height), mapGenAllocateOccupiedRect)
	return (*C.nox_mapgen_occupied_rect_legacy)(unsafe.Pointer(rect))
}

//export sub_521BC0
func sub_521BC0(roomToken C.int, pos *C.float2, width, height C.float) *C.float {
	room := (*C.uint8_t)(C.nox_mapgenLegacyPtrResolve(C.uintptr_t(uint32(roomToken))))
	return (*C.float)(unsafe.Pointer(nox_mapgenAddOccupiedRectNative_521BC0(room, pos, width, height)))
}

//export nox_mapgenClearTransientOccupiedRectsNative_521C10
func nox_mapgenClearTransientOccupiedRectsNative_521C10(room *C.uint8_t) *C.uint32_t {
	if room != nil {
		mapGenClearTransientOccupiedRects521C10(mapGenOccupiedRectHead(room), mapGenResolveOccupiedRect,
			func(rect *mapGenOccupiedRect5227B0) {
				raw := unsafe.Pointer(rect)
				C.nox_mapgenLegacyPtrForget(raw)
				C.free(raw)
			})
	}
	return nil
}

//export sub_521C10
func sub_521C10(roomToken C.int) *C.uint32_t {
	room := (*C.uint8_t)(C.nox_mapgenLegacyPtrResolve(C.uintptr_t(uint32(roomToken))))
	return nox_mapgenClearTransientOccupiedRectsNative_521C10(room)
}

//export sub_521EB0
func sub_521EB0(room *C.uint8_t, rect *C.nox_mapgen_occupied_rect_legacy) C.int {
	if room == nil || rect == nil {
		return 0
	}
	return C.int(bool2int(mapGenOccupiedRectInside521EB0(mapGenRoomBounds5226D0(room),
		(*mapGenOccupiedRect5227B0)(unsafe.Pointer(rect)).bounds)))
}

//export nox_mapgenOccupiedRectsIntersectNative_521F10
func nox_mapgenOccupiedRectsIntersectNative_521F10(room *C.uint8_t, rect *C.nox_mapgen_occupied_rect_legacy) *C.nox_mapgen_occupied_rect_legacy {
	if room == nil || rect == nil {
		return nil
	}
	match := mapGenOccupiedRectsIntersect521F10(*mapGenOccupiedRectHead(room),
		(*mapGenOccupiedRect5227B0)(unsafe.Pointer(rect)), mapGenResolveOccupiedRect)
	return (*C.nox_mapgen_occupied_rect_legacy)(unsafe.Pointer(match))
}
