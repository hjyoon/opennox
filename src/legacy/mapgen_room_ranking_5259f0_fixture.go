package legacy

/*
#include <stdint.h>
#include <stdlib.h>

#include "GAME4_2.h"
#include "mapgen_legacy_ptr.h"

extern uint32_t dword_5d4594_2487576;
extern uint32_t dword_5d4594_2487580;
extern uint32_t dword_5d4594_2487584;

typedef struct nox_test_mapgen_room_ranking_5259f0_result {
	uintptr_t theme_address;
	uintptr_t room_addresses[5];
	uintptr_t sorted_addresses[5];
	uintptr_t sorted_next_addresses[5];
	uintptr_t sorted_prev_addresses[5];
	uintptr_t farthest_address;
	uintptr_t native_farthest_address;
	uintptr_t legacy_farthest_address;
	uintptr_t sorted_head_address;
	uintptr_t legacy_sorted_head_address;
	uintptr_t farthest_after_cleanup;
	uintptr_t sorted_head_after_cleanup;
	uint32_t room_tokens[5];
	uint32_t sorted_next_tokens[5];
	uint32_t sorted_prev_tokens[5];
	uint32_t buckets[5];
	uint32_t flags[5];
	float distances[5];
	float normalized[5];
	float max_distance;
	float max_distance_after_cleanup;
	uint8_t visits[5];
	int grid_initialized;
	int rooms_added;
	int links_added;
} nox_test_mapgen_room_ranking_5259f0_result;

static nox_test_mapgen_room_ranking_5259f0_result nox_test_mapgen_room_ranking_5259f0(void) {
	nox_test_mapgen_room_ranking_5259f0_result out = {0};
	uint8_t* theme = (uint8_t*)calloc(1, 0x45C);
	uint8_t* rooms[5] = {0};
	int room_count = 0;
	if (!theme) {
		return out;
	}
	out.theme_address = (uintptr_t)theme;
	*(uint32_t*)(theme + 68) = 32;
	out.grid_initialized = sub_520EA0(theme);
	if (!out.grid_initialized) {
		free(theme);
		return out;
	}

	for (int i = 0; i < 5; ++i) {
		rooms[i] = (uint8_t*)nox_xxx_mapGenMakeRoomStruct_521940(2, 2);
		if (!rooms[i]) {
			goto cleanup;
		}
		out.room_addresses[i] = (uintptr_t)rooms[i];
		out.room_tokens[i] = nox_mapgenLegacyPtrRegister(rooms[i]);
		float2 position = {(float)(i * 3) * 32.526913f, 0.0f};
		nox_xxx_mapGenSetRoomPos_521880((uint32_t*)rooms[i], &position);
		if (!nox_xxx_mapGenAddNewRoom_521730((uint32_t*)rooms[i])) {
			goto cleanup;
		}
		++room_count;
		++out.rooms_added;
	}

	for (int i = 0; i < 4; ++i) {
		if (!nox_mapgenRoomLinkBothNative_521A70(rooms[i], rooms[i + 1], 2)) {
			goto cleanup;
		}
		++out.links_added;
	}

	sub_5259F0((int)out.room_tokens[0], 0, 0.0f);
	out.farthest_address = (uintptr_t)sub_525AF0((int)out.room_tokens[0]);
	out.native_farthest_address = (uintptr_t)nox_mapgenFarthestRoomNative_5259E0();
	out.legacy_farthest_address = (uintptr_t)nox_mapgenLegacyPtrResolve(dword_5d4594_2487576);
	out.sorted_head_address = (uintptr_t)nox_mapgenSortedRoomHeadNative_525C90();
	out.legacy_sorted_head_address = (uintptr_t)nox_mapgenLegacyPtrResolve(dword_5d4594_2487584);
	out.max_distance = nox_mapgenMaxRoomDistanceNative_5259D0();

	uint8_t* sorted = nox_mapgenSortedRoomHeadNative_525C90();
	for (int i = 0; i < 5 && sorted; ++i) {
		out.sorted_addresses[i] = (uintptr_t)sorted;
		out.sorted_next_tokens[i] = *(uint32_t*)(sorted + 64);
		out.sorted_prev_tokens[i] = *(uint32_t*)(sorted + 68);
		out.sorted_next_addresses[i] = (uintptr_t)nox_mapgenRoomSortedNextNative_525C90(sorted);
		out.sorted_prev_addresses[i] = (uintptr_t)nox_mapgenRoomSortedPrevNative_525C90(sorted);
		out.distances[i] = *(float*)(sorted + 356);
		out.normalized[i] = *(float*)(sorted + 360);
		out.buckets[i] = *(uint32_t*)(sorted + 364);
		out.flags[i] = *(uint32_t*)(sorted + 52);
		out.visits[i] = sorted[220];
		sorted = nox_mapgenRoomSortedNextNative_525C90(sorted);
	}

cleanup:
	if (room_count) {
		nox_xxx_mapGenFreeTopRoom_521A40();
		for (int i = room_count; i < 5; ++i) {
			if (rooms[i]) {
				sub_521A10(rooms[i]);
			}
		}
	} else {
		for (int i = 0; i < 5; ++i) {
			if (rooms[i]) {
				sub_521A10(rooms[i]);
			}
		}
	}
	out.farthest_after_cleanup = (uintptr_t)nox_mapgenFarthestRoomNative_5259E0();
	out.sorted_head_after_cleanup = (uintptr_t)nox_mapgenSortedRoomHeadNative_525C90();
	out.max_distance_after_cleanup = nox_mapgenMaxRoomDistanceNative_5259D0();
	sub_520F80();
	free(theme);
	return out;
}
*/
import "C"

type mapgenRoomRankingResult5259F0 struct {
	themeAddress            uintptr
	roomAddresses           [5]uintptr
	sortedAddresses         [5]uintptr
	sortedNextAddresses     [5]uintptr
	sortedPrevAddresses     [5]uintptr
	farthestAddress         uintptr
	nativeFarthestAddress   uintptr
	legacyFarthestAddress   uintptr
	sortedHeadAddress       uintptr
	legacySortedHeadAddress uintptr
	farthestAfterCleanup    uintptr
	sortedHeadAfterCleanup  uintptr
	roomTokens              [5]uint32
	sortedNextTokens        [5]uint32
	sortedPrevTokens        [5]uint32
	buckets                 [5]uint32
	flags                   [5]uint32
	distances               [5]float32
	normalized              [5]float32
	maxDistance             float32
	maxDistanceAfterCleanup float32
	visits                  [5]uint8
	gridInitialized         bool
	roomsAdded              int
	linksAdded              int
}

func mapgenRoomRankingFixture5259F0() mapgenRoomRankingResult5259F0 {
	got := C.nox_test_mapgen_room_ranking_5259f0()
	var out mapgenRoomRankingResult5259F0
	out.themeAddress = uintptr(got.theme_address)
	for i := range out.roomAddresses {
		out.roomAddresses[i] = uintptr(got.room_addresses[i])
		out.sortedAddresses[i] = uintptr(got.sorted_addresses[i])
		out.sortedNextAddresses[i] = uintptr(got.sorted_next_addresses[i])
		out.sortedPrevAddresses[i] = uintptr(got.sorted_prev_addresses[i])
		out.roomTokens[i] = uint32(got.room_tokens[i])
		out.sortedNextTokens[i] = uint32(got.sorted_next_tokens[i])
		out.sortedPrevTokens[i] = uint32(got.sorted_prev_tokens[i])
		out.buckets[i] = uint32(got.buckets[i])
		out.flags[i] = uint32(got.flags[i])
		out.distances[i] = float32(got.distances[i])
		out.normalized[i] = float32(got.normalized[i])
		out.visits[i] = uint8(got.visits[i])
	}
	out.farthestAddress = uintptr(got.farthest_address)
	out.nativeFarthestAddress = uintptr(got.native_farthest_address)
	out.legacyFarthestAddress = uintptr(got.legacy_farthest_address)
	out.sortedHeadAddress = uintptr(got.sorted_head_address)
	out.legacySortedHeadAddress = uintptr(got.legacy_sorted_head_address)
	out.farthestAfterCleanup = uintptr(got.farthest_after_cleanup)
	out.sortedHeadAfterCleanup = uintptr(got.sorted_head_after_cleanup)
	out.maxDistance = float32(got.max_distance)
	out.maxDistanceAfterCleanup = float32(got.max_distance_after_cleanup)
	out.gridInitialized = got.grid_initialized != 0
	out.roomsAdded = int(got.rooms_added)
	out.linksAdded = int(got.links_added)
	return out
}
