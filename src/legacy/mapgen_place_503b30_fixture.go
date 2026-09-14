package legacy

/*
#include <stddef.h>
#include <stdint.h>

#include "GAME4.h"

typedef struct nox_test_mapgen_pending_503b30_result {
	uintptr_t first_address;
	uintptr_t second_address;
	uintptr_t third_address;
	uintptr_t first_next;
	uintptr_t second_next;
	int first_script_id;
	int second_script_id;
	int third_script_id;
	uint32_t first_extent;
	uint32_t second_extent;
	uint32_t third_extent;
	uint32_t first_field_12;
	uint32_t second_field_12;
	uint32_t third_field_12;
} nox_test_mapgen_pending_503b30_result;

static nox_test_mapgen_pending_503b30_result nox_test_mapgen_pending_503b30(void) {
	nox_object_t first = {0};
	nox_object_t second = {0};
	nox_object_t third = {0};
	first.object_next = &second;
	second.object_next = &third;
	first.script_id = 101;
	second.script_id = -202;
	third.script_id = 303;
	first.extent = 0x10203040;
	second.extent = 0x50607080;
	third.extent = 0x90a0b0c0;
	first.field_12 = 0x11223344;
	second.field_12 = 0x55667788;
	third.field_12 = 0x99aabbcc;
	nox_mapgenClearPendingScriptIDs_503B30(&first);
	nox_mapgenClearPendingScriptIDs_503B30(NULL);
	return (nox_test_mapgen_pending_503b30_result){
		.first_address = (uintptr_t)&first,
		.second_address = (uintptr_t)&second,
		.third_address = (uintptr_t)&third,
		.first_next = (uintptr_t)first.object_next,
		.second_next = (uintptr_t)second.object_next,
		.first_script_id = first.script_id,
		.second_script_id = second.script_id,
		.third_script_id = third.script_id,
		.first_extent = first.extent,
		.second_extent = second.extent,
		.third_extent = third.extent,
		.first_field_12 = first.field_12,
		.second_field_12 = second.field_12,
		.third_field_12 = third.field_12,
	};
}
*/
import "C"

type mapgenPendingResult503B30 struct {
	addresses [3]uintptr
	next      [2]uintptr
	scriptIDs [3]int32
	extents   [3]uint32
	field12   [3]uint32
}

func mapgenPendingFixture503B30() mapgenPendingResult503B30 {
	v := C.nox_test_mapgen_pending_503b30()
	return mapgenPendingResult503B30{
		addresses: [3]uintptr{uintptr(v.first_address), uintptr(v.second_address), uintptr(v.third_address)},
		next:      [2]uintptr{uintptr(v.first_next), uintptr(v.second_next)},
		scriptIDs: [3]int32{int32(v.first_script_id), int32(v.second_script_id), int32(v.third_script_id)},
		extents:   [3]uint32{uint32(v.first_extent), uint32(v.second_extent), uint32(v.third_extent)},
		field12:   [3]uint32{uint32(v.first_field_12), uint32(v.second_field_12), uint32(v.third_field_12)},
	}
}
