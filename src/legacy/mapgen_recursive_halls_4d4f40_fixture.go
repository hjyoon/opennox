package legacy

/*
#include <stdint.h>
#include <string.h>

#include "GAME3_2.h"
#include "GAME4_2.h"
#include "mapgen_legacy_ptr.h"
#include "memmap.h"

extern uint32_t dword_5d4594_1549844;
extern uint32_t dword_5d4594_1550912;
extern uint32_t dword_5d4594_1550916;

typedef struct nox_test_mapgen_recursive_halls_4d4f40_result {
	uintptr_t theme_address;
	uintptr_t special_root_address;
	uintptr_t special_root_resolved_address;
	uintptr_t tracking_head_address;
	uintptr_t recursive_root_address;
	uintptr_t recursive_child_address;
	uintptr_t recursive_link_address;
	uint32_t special_root_token;
	uint32_t tracking_head_token;
	uint32_t recursive_link_token;
	int special_grid_initialized;
	int special_room_count;
	int tracked_room_count;
	int tracking_valid;
	int traversal_preserved_rooms;
	int recursive_grid_initialized;
	int recursive_result;
	int recursive_room_count;
	int recursive_root_link_count;
	int recursive_child_type;
	int cleanup_head_is_null;
} nox_test_mapgen_recursive_halls_4d4f40_result;

static int nox_test_mapgen_room_count_4d4f40(void) {
	int count = 0;
	for (uint8_t* room = (uint8_t*)nox_xxx_mapGenGetTopRoom_521710();
		 room && count < 256;
		 room = (uint8_t*)nox_mapgenRoomNextNative_521720(room)) {
		++count;
	}
	return count;
}

static nox_test_mapgen_recursive_halls_4d4f40_result nox_test_mapgen_recursive_halls_4d4f40(void) {
	nox_test_mapgen_recursive_halls_4d4f40_result out = {0};
	uint8_t* theme = getMemAt(0x5D4594, 1549796);
	uint8_t saved_theme[0x45C];
	uint32_t saved_tracking_head = dword_5d4594_1550912;
	uint32_t saved_root = dword_5d4594_1550916;
	uint32_t saved_join_chance = dword_5d4594_1549844;
	int* max_depth = getMemIntPtr(0x5D4594, 1549868);
	int* max_hall_depth = getMemIntPtr(0x5D4594, 1549816);
	int* branch_chance = getMemIntPtr(0x5D4594, 1549820);
	int* room_chance = getMemIntPtr(0x5D4594, 1549824);
	int saved_max_depth = *max_depth;
	int saved_max_hall_depth = *max_hall_depth;
	int saved_branch_chance = *branch_chance;
	int saved_room_chance = *room_chance;
	uint8_t* recursive_root = NULL;
	int recursive_root_added = 0;

	memcpy(saved_theme, theme, sizeof(saved_theme));
	out.theme_address = (uintptr_t)theme;
	nox_xxx_mapGenFreeTopRoom_521A40();
	sub_520F80();
	dword_5d4594_1550912 = 0;
	dword_5d4594_1550916 = 0;

	memset(theme, 0, sizeof(saved_theme));
	*(uint32_t*)theme = 1;
	*(float*)(theme + 64) = 4096.0f;
	*(uint32_t*)(theme + 68) = 128;
	*max_depth = 1;
	nox_xxx_mapGenSetRngSeed_526AB0(0x4D4F40u);
	out.special_grid_initialized = sub_520EA0(theme);
	if (out.special_grid_initialized) {
		uint8_t* special_root = (uint8_t*)nox_xxx_mapGenMkSmallRoom_4D4F40(theme);
		out.special_root_address = (uintptr_t)special_root;
		out.special_root_token = dword_5d4594_1550916;
		out.special_root_resolved_address =
			(uintptr_t)nox_mapgenLegacyPtrResolve(out.special_root_token);
		out.tracking_head_token = dword_5d4594_1550912;
		out.tracking_head_address =
			(uintptr_t)nox_mapgenLegacyPtrResolve(out.tracking_head_token);
		out.special_room_count = nox_test_mapgen_room_count_4d4f40();

		uint32_t seen[64] = {0};
		uint32_t token = out.tracking_head_token;
		out.tracking_valid = token != 0;
		while (token && out.tracked_room_count < 64) {
			for (int i = 0; i < out.tracked_room_count; ++i) {
				if (seen[i] == token) {
					out.tracking_valid = 0;
					token = 0;
					break;
				}
			}
			if (!token) {
				break;
			}
			uint8_t* tracked = (uint8_t*)nox_mapgenLegacyPtrResolve(token);
			if (!tracked) {
				out.tracking_valid = 0;
				break;
			}
			seen[out.tracked_room_count++] = token;
			token = *(uint32_t*)(tracked + 84);
		}
		if (token) {
			out.tracking_valid = 0;
		}

		int before_traversal = out.special_room_count;
		sub_4D52F0();
		out.traversal_preserved_rooms =
			nox_test_mapgen_room_count_4d4f40() == before_traversal;
	}

	nox_xxx_mapGenFreeTopRoom_521A40();
	sub_520F80();
	dword_5d4594_1550912 = 0;
	dword_5d4594_1550916 = 0;

	memset(theme, 0, sizeof(saved_theme));
	*(uint32_t*)(theme + 4) = 5;
	*(uint32_t*)(theme + 32) = 5;
	*(float*)(theme + 64) = 4096.0f;
	*(uint32_t*)(theme + 68) = 128;
	*max_depth = 1;
	*max_hall_depth = 0;
	*branch_chance = 100;
	*room_chance = 100;
	dword_5d4594_1549844 = 100;
	nox_xxx_mapGenSetRngSeed_526AB0(0x4D5630u);
	out.recursive_grid_initialized = sub_520EA0(theme);
	if (out.recursive_grid_initialized) {
		recursive_root = (uint8_t*)nox_mapgenMakeHallStructNative_523E30(2, 2, 4);
		if (recursive_root) {
			float2 root_position = {0.0f, 0.0f};
			nox_xxx_mapGenSetRoomPos_521880((uint32_t*)recursive_root, &root_position);
			recursive_root_added =
				nox_xxx_mapGenAddNewRoom_521730((uint32_t*)recursive_root) != 0;
			if (recursive_root_added) {
				out.recursive_root_address = (uintptr_t)recursive_root;
				out.recursive_result = sub_4D5630(recursive_root, 0, 0, 2, NULL);
				out.recursive_room_count = nox_test_mapgen_room_count_4d4f40();
				out.recursive_root_link_count = recursive_root[216];
				out.recursive_link_token = *(uint32_t*)(recursive_root + 88);
				uint8_t* child =
					(uint8_t*)nox_mapgenLegacyPtrResolve(out.recursive_link_token);
				out.recursive_child_address = (uintptr_t)child;
				out.recursive_link_address =
					(uintptr_t)nox_mapgenLegacyPtrResolve(out.recursive_link_token);
				out.recursive_child_type = child ? *(uint32_t*)child : 0;
			}
		}
	}

	if (recursive_root_added) {
		nox_xxx_mapGenFreeTopRoom_521A40();
		recursive_root = NULL;
	} else if (recursive_root) {
		sub_521A10(recursive_root);
		recursive_root = NULL;
	}
	sub_520F80();
	out.cleanup_head_is_null = nox_xxx_mapGenGetTopRoom_521710() == NULL;

	memcpy(theme, saved_theme, sizeof(saved_theme));
	*max_depth = saved_max_depth;
	*max_hall_depth = saved_max_hall_depth;
	*branch_chance = saved_branch_chance;
	*room_chance = saved_room_chance;
	dword_5d4594_1549844 = saved_join_chance;
	dword_5d4594_1550912 = saved_tracking_head;
	dword_5d4594_1550916 = saved_root;
	return out;
}
*/
import "C"

type mapgenRecursiveHallsResult4D4F40 struct {
	themeAddress               uintptr
	specialRootAddress         uintptr
	specialRootResolvedAddress uintptr
	trackingHeadAddress        uintptr
	recursiveRootAddress       uintptr
	recursiveChildAddress      uintptr
	recursiveLinkAddress       uintptr
	specialRootToken           uint32
	trackingHeadToken          uint32
	recursiveLinkToken         uint32
	specialGridInitialized     bool
	specialRoomCount           int
	trackedRoomCount           int
	trackingValid              bool
	traversalPreservedRooms    bool
	recursiveGridInitialized   bool
	recursiveResult            int
	recursiveRoomCount         int
	recursiveRootLinkCount     int
	recursiveChildType         int
	cleanupHeadIsNull          bool
}

func mapgenRecursiveHallsFixture4D4F40() mapgenRecursiveHallsResult4D4F40 {
	got := C.nox_test_mapgen_recursive_halls_4d4f40()
	return mapgenRecursiveHallsResult4D4F40{
		themeAddress:               uintptr(got.theme_address),
		specialRootAddress:         uintptr(got.special_root_address),
		specialRootResolvedAddress: uintptr(got.special_root_resolved_address),
		trackingHeadAddress:        uintptr(got.tracking_head_address),
		recursiveRootAddress:       uintptr(got.recursive_root_address),
		recursiveChildAddress:      uintptr(got.recursive_child_address),
		recursiveLinkAddress:       uintptr(got.recursive_link_address),
		specialRootToken:           uint32(got.special_root_token),
		trackingHeadToken:          uint32(got.tracking_head_token),
		recursiveLinkToken:         uint32(got.recursive_link_token),
		specialGridInitialized:     got.special_grid_initialized != 0,
		specialRoomCount:           int(got.special_room_count),
		trackedRoomCount:           int(got.tracked_room_count),
		trackingValid:              got.tracking_valid != 0,
		traversalPreservedRooms:    got.traversal_preserved_rooms != 0,
		recursiveGridInitialized:   got.recursive_grid_initialized != 0,
		recursiveResult:            int(got.recursive_result),
		recursiveRoomCount:         int(got.recursive_room_count),
		recursiveRootLinkCount:     int(got.recursive_root_link_count),
		recursiveChildType:         int(got.recursive_child_type),
		cleanupHeadIsNull:          got.cleanup_head_is_null != 0,
	}
}
