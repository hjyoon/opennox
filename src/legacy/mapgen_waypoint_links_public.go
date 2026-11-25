package legacy

/*
#include "GAME4_1.h"
#include "GAME4_2.h"
extern nox_waypoint_t* dword_5d4594_2487244;
*/
import "C"

import (
	"unsafe"

	"github.com/opennox/libs/types"

	"github.com/opennox/opennox/v1/common/memmap"
	"github.com/opennox/opennox/v1/server"
)

const (
	mapgenWaypointStateBase      = 0x973F18
	mapgenWaypointKindOffset     = 35972
	mapgenWaypointAutoLinkOffset = 35976
)

func mapgenLinkWaypoints(from, to *server.Waypoint, kind byte) bool {
	if from == nil || to == nil || from == to || from.PointsCnt >= 31 {
		return false
	}
	for _, link := range from.Points[:from.PointsCnt] {
		if link.Waypoint == to && link.Ind == kind {
			return false
		}
	}
	point := &from.Points[from.PointsCnt]
	// The next slot belongs to C and can still contain an unused pointer.
	globalStateNativeStorePointer(unsafe.Pointer(&point.Waypoint), unsafe.Pointer(to))
	point.Ind = kind
	from.PointsCnt++
	return true
}

func mapgenFindWaypoint(position *C.float2) *server.Waypoint {
	var fixed mapgenWorldPoint
	if !mapgenFixWorldCoords((*mapgenWorldPoint)(unsafe.Pointer(position)), &fixed) {
		return nil
	}
	return GetServer().S().WPs.Sub_579AD0(types.Ptf(fixed.x, fixed.y))
}

//export sub_51D0F0
func sub_51D0F0(kind C.char) C.int {
	*memmap.PtrUint8(mapgenWaypointStateBase, mapgenWaypointKindOffset) = byte(kind)
	return 1
}

//export sub_51D100
func sub_51D100(mode C.int) C.int {
	if mode != 0 && mode != 1 {
		return 0
	}
	*memmap.PtrUint32(mapgenWaypointStateBase, mapgenWaypointAutoLinkOffset) = uint32(mode)
	return 1
}

//export sub_51D120
func sub_51D120(position *C.float2) *C.nox_waypoint_t {
	var fixed mapgenWorldPoint
	if !mapgenFixWorldCoords((*mapgenWorldPoint)(unsafe.Pointer(position)), &fixed) {
		return nil
	}
	created := GetServer().S().NewWaypoint(types.Ptf(fixed.x, fixed.y))
	if created == nil {
		return nil
	}
	last := asWaypointS(C.dword_5d4594_2487244)
	if last != nil && memmap.Uint32(mapgenWaypointStateBase, mapgenWaypointAutoLinkOffset) == 1 {
		mapgenLinkWaypoints(last, created, memmap.Uint8(mapgenWaypointStateBase, mapgenWaypointKindOffset))
		mapgenLinkWaypoints(created, asWaypointS(C.dword_5d4594_2487244), memmap.Uint8(mapgenWaypointStateBase, mapgenWaypointKindOffset))
	}
	globalStateNativeStorePointer(unsafe.Pointer(&C.dword_5d4594_2487244), unsafe.Pointer(created))
	return asWaypointC(created)
}

//export sub_51D1A0
func sub_51D1A0(position *C.float2) *C.nox_waypoint_t {
	return asWaypointC(mapgenFindWaypoint(position))
}

//export sub_51D3F0
func sub_51D3F0(from, to *C.float2) C.int {
	if from == nil || to == nil {
		return 0
	}
	source := mapgenFindWaypoint(from)
	if source == nil {
		return 0
	}
	target := mapgenFindWaypoint(to)
	if target == nil {
		return 0
	}
	kind := memmap.Uint8(mapgenWaypointStateBase, mapgenWaypointKindOffset)
	return C.int(bool2int(mapgenLinkWaypoints(source, target, kind)))
}

//export sub_522C80
func sub_522C80(position *C.float2) *C.nox_waypoint_t {
	if existing := mapgenFindWaypoint(position); existing != nil {
		return asWaypointC(existing)
	}
	return sub_51D120(position)
}

//export nox_mapgenWaypointLinkNative_51D300
func nox_mapgenWaypointLinkNative_51D300(from, to *C.nox_waypoint_t, kind C.uchar) C.int {
	return C.int(bool2int(mapgenLinkWaypoints(asWaypointS(from), asWaypointS(to), byte(kind))))
}

//export sub_51D2C0
func sub_51D2C0(from, to C.int) C.int {
	return sub_51D300(from, to, C.char(memmap.Uint8(mapgenWaypointStateBase, mapgenWaypointKindOffset)))
}

//export sub_51D300
func sub_51D300(from, to C.int, kind C.char) C.int {
	source := (*server.Waypoint)(resolveMapgenPointer(uintptr(uint32(from))))
	target := (*server.Waypoint)(resolveMapgenPointer(uintptr(uint32(to))))
	return C.int(bool2int(mapgenLinkWaypoints(source, target, byte(kind))))
}
