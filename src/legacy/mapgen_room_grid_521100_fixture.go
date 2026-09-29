package legacy

/*
#include <stdint.h>
#include <stdlib.h>

#include "GAME4_2.h"
#include "mapgen_legacy_ptr.h"

typedef struct nox_test_mapgen_room_grid_521100_result {
	uintptr_t theme_address;
	uintptr_t first_address;
	uintptr_t second_address;
	uintptr_t collision_address;
	uintptr_t first_head_address;
	uintptr_t second_head_address;
	uintptr_t second_next_address;
	uintptr_t removed_head_address;
	uintptr_t vacated_address;
	int grid_initialized;
	int first_added;
	int second_added;
	int second_removed;
} nox_test_mapgen_room_grid_521100_result;

static nox_test_mapgen_room_grid_521100_result nox_test_mapgen_room_grid_521100(void) {
	nox_test_mapgen_room_grid_521100_result out = {0};
	uint8_t* theme = (uint8_t*)calloc(1, 0x45C);
	uint8_t* first = NULL;
	uint8_t* second = NULL;
	int first_linked = 0;
	int second_linked = 0;
	if (!theme) {
		return out;
	}
	out.theme_address = (uintptr_t)theme;
	*(uint32_t*)(theme + 68) = 4;
	out.grid_initialized = sub_520EA0(theme);
	if (!out.grid_initialized) {
		free(theme);
		return out;
	}

	first = (uint8_t*)nox_xxx_mapGenMakeRoomStruct_521940(2, 2);
	second = (uint8_t*)nox_xxx_mapGenMakeRoomStruct_521940(2, 2);
	if (!first || !second) {
		goto cleanup;
	}
	out.first_address = (uintptr_t)first;
	out.second_address = (uintptr_t)second;
	float2 position = {0};
	nox_xxx_mapGenSetRoomPos_521880((uint32_t*)first, &position);
	nox_xxx_mapGenSetRoomPos_521880((uint32_t*)second, &position);
	out.first_added = nox_xxx_mapGenAddNewRoom_521730((uint32_t*)first);
	first_linked = out.first_added != 0;
	out.first_head_address = (uintptr_t)nox_xxx_mapGenGetTopRoom_521710();
	out.collision_address = (uintptr_t)nox_mapgenRoomAtNative_521200(second);

	position.field_0 = 2.0f * 32.526913f;
	position.field_4 = 2.0f * 32.526913f;
	nox_xxx_mapGenSetRoomPos_521880((uint32_t*)second, &position);
	out.second_added = nox_xxx_mapGenAddNewRoom_521730((uint32_t*)second);
	second_linked = out.second_added != 0;
	out.second_head_address = (uintptr_t)nox_xxx_mapGenGetTopRoom_521710();
	out.second_next_address = (uintptr_t)nox_mapgenRoomNextNative_521720(second);

	if (second_linked) {
		uint32_t token = nox_mapgenLegacyPtrRegister(second);
		out.second_removed = sub_521760((int)token);
		second_linked = 0;
	}
	out.removed_head_address = (uintptr_t)nox_xxx_mapGenGetTopRoom_521710();
	out.vacated_address = (uintptr_t)nox_mapgenRoomAtNative_521200(second);

cleanup:
	if (second_linked) {
		sub_521760((int)nox_mapgenLegacyPtrRegister(second));
		second_linked = 0;
	}
	if (first_linked) {
		nox_xxx_mapGenFreeTopRoom_521A40();
		first = NULL;
	} else if (first) {
		sub_521A10(first);
		first = NULL;
	}
	if (second) {
		sub_521A10(second);
		second = NULL;
	}
	sub_520F80();
	free(theme);
	return out;
}
*/
import "C"

type mapgenRoomGridResult521100 struct {
	themeAddress      uintptr
	firstAddress      uintptr
	secondAddress     uintptr
	collisionAddress  uintptr
	firstHeadAddress  uintptr
	secondHeadAddress uintptr
	secondNextAddress uintptr
	removedHead       uintptr
	vacatedAddress    uintptr
	gridInitialized   bool
	firstAdded        bool
	secondAdded       bool
	secondRemoved     bool
}

func mapgenRoomGridFixture521100() mapgenRoomGridResult521100 {
	got := C.nox_test_mapgen_room_grid_521100()
	return mapgenRoomGridResult521100{
		themeAddress:      uintptr(got.theme_address),
		firstAddress:      uintptr(got.first_address),
		secondAddress:     uintptr(got.second_address),
		collisionAddress:  uintptr(got.collision_address),
		firstHeadAddress:  uintptr(got.first_head_address),
		secondHeadAddress: uintptr(got.second_head_address),
		secondNextAddress: uintptr(got.second_next_address),
		removedHead:       uintptr(got.removed_head_address),
		vacatedAddress:    uintptr(got.vacated_address),
		gridInitialized:   got.grid_initialized != 0,
		firstAdded:        got.first_added != 0,
		secondAdded:       got.second_added != 0,
		secondRemoved:     got.second_removed != 0,
	}
}
