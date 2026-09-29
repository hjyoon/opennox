package legacy

/*
#include <stdint.h>

#include "GAME4_2.h"
#include "mapgen_legacy_ptr.h"

typedef struct nox_test_mapgen_occupied_rects_521bc0_result {
	uintptr_t room_address;
	uintptr_t first_address;
	uintptr_t second_address;
	uintptr_t head_address;
	uintptr_t next_address;
	uintptr_t remaining_address;
	uint32_t stored_head_token;
	int legacy_wrapper_added;
	int intersects;
	int point_occupied;
	int transient_removed;
	int all_removed;
} nox_test_mapgen_occupied_rects_521bc0_result;

static nox_test_mapgen_occupied_rects_521bc0_result nox_test_mapgen_occupied_rects_521bc0(void) {
	nox_test_mapgen_occupied_rects_521bc0_result out = {0};
	uint8_t* room = (uint8_t*)nox_xxx_mapGenMakeRoomStruct_521940(4, 4);
	if (!room) {
		return out;
	}
	out.room_address = (uintptr_t)room;
	uint32_t room_token = nox_mapgenLegacyPtrRegister(room);

	float2 first_pos = {10.0f, 20.0f};
	float* first = sub_521BC0((int)room_token, &first_pos, 5.0f, 6.0f);
	out.legacy_wrapper_added = first != NULL;
	out.first_address = (uintptr_t)first;
	if (!first) {
		sub_521A10(room);
		return out;
	}

	float2 second_pos = {30.0f, 40.0f};
	float* second = nox_mapgenAddOccupiedRectNative_521BC0(room, &second_pos, 7.0f, 8.0f);
	out.second_address = (uintptr_t)second;
	if (!second) {
		sub_521A10(room);
		return out;
	}

	out.stored_head_token = *(uint32_t*)(room + 368);
	uint32_t* head = (uint32_t*)nox_mapgenLegacyPtrResolve(out.stored_head_token);
	out.head_address = (uintptr_t)head;
	out.next_address = head ? (uintptr_t)nox_mapgenLegacyPtrResolve(head[6]) : 0;

	float candidate[7] = {0.0f, 12.0f, 21.0f, 14.0f, 23.0f, 0.0f, 0.0f};
	out.intersects = nox_mapgenOccupiedRectsIntersectNative_521F10(room, candidate);
	float point[2] = {11.0f, 22.0f};
	out.point_occupied = nox_mapgenPointOccupiedNative_5227B0(room, point);

	*(uint32_t*)second = 1;
	nox_mapgenClearTransientOccupiedRectsNative_521C10(room);
	out.remaining_address = (uintptr_t)nox_mapgenLegacyPtrResolve(*(uint32_t*)(room + 368));
	out.transient_removed = out.remaining_address == out.first_address;

	*(uint32_t*)first = 1;
	sub_521C10((int)room_token);
	out.all_removed = *(uint32_t*)(room + 368) == 0;
	sub_521A10(room);
	return out;
}
*/
import "C"

type mapgenOccupiedRectsResult521BC0 struct {
	roomAddress        uintptr
	firstAddress       uintptr
	secondAddress      uintptr
	headAddress        uintptr
	nextAddress        uintptr
	remainingAddress   uintptr
	storedHeadToken    uint32
	legacyWrapperAdded bool
	intersects         bool
	pointOccupied      bool
	transientRemoved   bool
	allRemoved         bool
}

func mapgenOccupiedRectsFixture521BC0() mapgenOccupiedRectsResult521BC0 {
	got := C.nox_test_mapgen_occupied_rects_521bc0()
	return mapgenOccupiedRectsResult521BC0{
		roomAddress:        uintptr(got.room_address),
		firstAddress:       uintptr(got.first_address),
		secondAddress:      uintptr(got.second_address),
		headAddress:        uintptr(got.head_address),
		nextAddress:        uintptr(got.next_address),
		remainingAddress:   uintptr(got.remaining_address),
		storedHeadToken:    uint32(got.stored_head_token),
		legacyWrapperAdded: got.legacy_wrapper_added != 0,
		intersects:         got.intersects != 0,
		pointOccupied:      got.point_occupied != 0,
		transientRemoved:   got.transient_removed != 0,
		allRemoved:         got.all_removed != 0,
	}
}
