package legacy

/*
#include <stdint.h>

#include "GAME4_2.h"
#include "mapgen_legacy_ptr.h"

typedef struct nox_test_mapgen_adjacent_reservation_524fb0_result {
	uintptr_t room_address[4];
	uintptr_t neighbor_address[4];
	uintptr_t rectangle_address[4];
	uint32_t room_token[4];
	uint32_t neighbor_token[4];
	uint32_t rectangle_token[4];
	float min_x[4];
	float min_y[4];
	float max_x[4];
	float max_y[4];
	int added[4];
	int used_legacy_wrapper;
} nox_test_mapgen_adjacent_reservation_524fb0_result;

static nox_test_mapgen_adjacent_reservation_524fb0_result nox_test_mapgen_adjacent_reservation_524fb0(void) {
	nox_test_mapgen_adjacent_reservation_524fb0_result out = {0};
	for (int direction = 0; direction < 4; ++direction) {
		uint8_t* room = (uint8_t*)nox_xxx_mapGenMakeRoomStruct_521940(4, 3);
		uint8_t* neighbor = (uint8_t*)nox_xxx_mapGenMakeRoomStruct_521940(2, 5);
		if (!room || !neighbor) {
			if (room) {
				sub_521A10(room);
			}
			if (neighbor) {
				sub_521A10(neighbor);
			}
			continue;
		}

		out.room_address[direction] = (uintptr_t)room;
		out.neighbor_address[direction] = (uintptr_t)neighbor;
		out.room_token[direction] = nox_mapgenLegacyPtrRegister(room);
		out.neighbor_token[direction] = nox_mapgenLegacyPtrRegister(neighbor);

		float2 room_position = {100.0f, 200.0f};
		float2 neighbor_position = {110.0f, 210.0f};
		nox_xxx_mapGenSetRoomPos_521880((uint32_t*)room, &room_position);
		nox_xxx_mapGenSetRoomPos_521880((uint32_t*)neighbor, &neighbor_position);

		if (direction == 0) {
			sub_524FB0((int)out.room_token[direction],
				(int)out.neighbor_token[direction], direction);
			out.used_legacy_wrapper = 1;
		} else {
			nox_mapgenReserveAdjacentNative_524FB0(room, neighbor, direction);
		}

		out.rectangle_token[direction] = *(uint32_t*)(room + 368);
		uint8_t* rectangle =
			(uint8_t*)nox_mapgenLegacyPtrResolve(out.rectangle_token[direction]);
		out.rectangle_address[direction] = (uintptr_t)rectangle;
		if (rectangle) {
			out.added[direction] = 1;
			out.min_x[direction] = *(float*)(rectangle + 4);
			out.min_y[direction] = *(float*)(rectangle + 8);
			out.max_x[direction] = *(float*)(rectangle + 12);
			out.max_y[direction] = *(float*)(rectangle + 16);
		}

		sub_521A10(neighbor);
		sub_521A10(room);
	}
	return out;
}
*/
import "C"

type mapgenAdjacentReservationCase524FB0 struct {
	roomAddress      uintptr
	neighborAddress  uintptr
	rectangleAddress uintptr
	roomToken        uint32
	neighborToken    uint32
	rectangleToken   uint32
	minX             float32
	minY             float32
	maxX             float32
	maxY             float32
	added            bool
}

type mapgenAdjacentReservationResult524FB0 struct {
	cases             [4]mapgenAdjacentReservationCase524FB0
	usedLegacyWrapper bool
}

func mapgenAdjacentReservationFixture524FB0() mapgenAdjacentReservationResult524FB0 {
	got := C.nox_test_mapgen_adjacent_reservation_524fb0()
	out := mapgenAdjacentReservationResult524FB0{
		usedLegacyWrapper: got.used_legacy_wrapper != 0,
	}
	for direction := range out.cases {
		out.cases[direction] = mapgenAdjacentReservationCase524FB0{
			roomAddress:      uintptr(got.room_address[direction]),
			neighborAddress:  uintptr(got.neighbor_address[direction]),
			rectangleAddress: uintptr(got.rectangle_address[direction]),
			roomToken:        uint32(got.room_token[direction]),
			neighborToken:    uint32(got.neighbor_token[direction]),
			rectangleToken:   uint32(got.rectangle_token[direction]),
			minX:             float32(got.min_x[direction]),
			minY:             float32(got.min_y[direction]),
			maxX:             float32(got.max_x[direction]),
			maxY:             float32(got.max_y[direction]),
			added:            got.added[direction] != 0,
		}
	}
	return out
}
