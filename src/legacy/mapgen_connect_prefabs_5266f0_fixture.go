package legacy

/*
#include <stdint.h>
#include <stdlib.h>

#include "GAME4_2.h"
#include "mapgen_legacy_ptr.h"
#include "memmap.h"

typedef struct nox_test_mapgen_connect_prefabs_5266f0_result {
	uintptr_t theme_address;
	uintptr_t prefab_address;
	uintptr_t old_room_address;
	uintptr_t new_room_address;
	uintptr_t candidate_addresses[3];
	uintptr_t callback_addresses[3];
	uintptr_t first_next_address;
	uintptr_t forward_neighbor_address;
	uintptr_t reverse_neighbor_address;
	int result;
	int calls;
	int linked;
	int has_type_one_neighbor;
} nox_test_mapgen_connect_prefabs_5266f0_result;

static uintptr_t nox_test_mapgen_candidates_5266f0[3];
static int nox_test_mapgen_calls_5266f0;

static int nox_test_mapgen_connect_prefab_5266f0(uint8_t* prefab, int direction, uint8_t* candidate) {
	(void)prefab;
	if (direction != 0) {
		return 0;
	}
	int call = nox_test_mapgen_calls_5266f0++;
	if (call < 3) {
		nox_test_mapgen_candidates_5266f0[call] = (uintptr_t)candidate;
	}
	return call == 1;
}

static nox_test_mapgen_connect_prefabs_5266f0_result nox_test_mapgen_connect_prefabs_5266f0(void) {
	nox_test_mapgen_connect_prefabs_5266f0_result out = {0};
	uint8_t* theme = (uint8_t*)calloc(1, 0x45C);
	uint8_t* prefab = (uint8_t*)calloc(1, 160);
	uint8_t* old_room = NULL;
	uint8_t* candidates[3] = {0};
	uint32_t saved_opposite_directions[4] = {0};
	int grid_initialized = 0;
	int directions_initialized = 0;
	if (!theme || !prefab) {
		goto cleanup;
	}
	out.theme_address = (uintptr_t)theme;
	out.prefab_address = (uintptr_t)prefab;
	uint32_t prefab_token = nox_mapgenLegacyPtrRegister(prefab);
	if (!prefab_token) {
		goto cleanup;
	}
	static const uint32_t opposite_directions[4] = {1, 0, 3, 2};
	for (int i = 0; i < 4; ++i) {
		saved_opposite_directions[i] = *getMemU32Ptr(0x587000, 254952 + 4 * i);
		*getMemU32Ptr(0x587000, 254952 + 4 * i) = opposite_directions[i];
	}
	directions_initialized = 1;
	*(uint32_t*)(theme + 68) = 32;
	grid_initialized = sub_520EA0(theme);
	if (!grid_initialized) {
		goto cleanup;
	}

	old_room = (uint8_t*)nox_xxx_mapGenMakeRoomStruct_521940(2, 2);
	for (int i = 0; i < 3; ++i) {
		candidates[i] = (uint8_t*)nox_xxx_mapGenMakeRoomStruct_521940(2, 2);
		if (!candidates[i]) {
			goto cleanup_rooms;
		}
		out.candidate_addresses[i] = (uintptr_t)candidates[i];
	}
	if (!old_room) {
		goto cleanup_rooms;
	}
	out.old_room_address = (uintptr_t)old_room;

	float2 position = {0};
	nox_xxx_mapGenSetRoomPos_521880((uint32_t*)old_room, &position);
	if (!nox_xxx_mapGenAddNewRoom_521730((uint32_t*)old_room)) {
		goto cleanup_rooms;
	}
	old_room = NULL;
	const int cells[3] = {-4, -8, -12};
	for (int i = 0; i < 3; ++i) {
		position.field_0 = 0.0f;
		position.field_4 = (float)cells[i] * 32.526913f;
		nox_xxx_mapGenSetRoomPos_521880((uint32_t*)candidates[i], &position);
		if (!nox_xxx_mapGenAddNewRoom_521730((uint32_t*)candidates[i])) {
			goto cleanup_rooms;
		}
		candidates[i] = NULL;
	}

	uint8_t* first = (uint8_t*)out.candidate_addresses[0];
	uint8_t* second = (uint8_t*)out.candidate_addresses[1];
	out.linked = nox_mapgenRoomLinkBothNative_521A70(first, second, 2);
	out.has_type_one_neighbor = nox_mapgenRoomHasTypeOneNeighborNative_5218B0(first, 2);
	out.forward_neighbor_address = (uintptr_t)nox_mapgenLegacyPtrResolve(*(uint32_t*)(first + 152));
	out.reverse_neighbor_address = (uintptr_t)nox_mapgenLegacyPtrResolve(*(uint32_t*)(second + 184));

	*(uint32_t*)(theme + 80) = prefab_token;
	*(uint32_t*)(prefab + 76) = 1;
	*(int32_t*)(prefab + 80) = 0;
	*(int32_t*)(prefab + 84) = 0;
	*(uint32_t*)(prefab + 92) = 1;
	*(float*)(prefab + 60) = 2.0f * 32.526913f;
	*(float*)(prefab + 64) = 2.0f * 32.526913f;
	*(uint32_t*)(prefab + 148) = nox_mapgenLegacyPtrRegister((void*)out.old_room_address);

	for (int i = 0; i < 3; ++i) {
		nox_test_mapgen_candidates_5266f0[i] = 0;
	}
	nox_test_mapgen_calls_5266f0 = 0;
	out.result = nox_mapgenConnectPrefabsWithCallback_5266F0(
		theme, nox_test_mapgen_connect_prefab_5266f0);
	out.calls = nox_test_mapgen_calls_5266f0;
	for (int i = 0; i < 3; ++i) {
		out.callback_addresses[i] = nox_test_mapgen_candidates_5266f0[i];
	}
	out.new_room_address = (uintptr_t)nox_mapgenLegacyPtrResolve(*(uint32_t*)(prefab + 148));
	if (out.calls > 0) {
		uint8_t* callback_first = (uint8_t*)out.callback_addresses[0];
		out.first_next_address = (uintptr_t)nox_mapgenLegacyPtrResolve(*(uint32_t*)(callback_first + 72));
	}

cleanup_rooms:
	nox_xxx_mapGenFreeTopRoom_521A40();
	for (int i = 0; i < 3; ++i) {
		if (candidates[i]) {
			sub_521A10(candidates[i]);
		}
	}
	if (old_room) {
		sub_521A10(old_room);
	}
cleanup:
	if (directions_initialized) {
		for (int i = 0; i < 4; ++i) {
			*getMemU32Ptr(0x587000, 254952 + 4 * i) = saved_opposite_directions[i];
		}
	}
	if (grid_initialized) {
		sub_520F80();
	}
	if (prefab) {
		nox_mapgenLegacyPtrForget(prefab);
		free(prefab);
	}
	free(theme);
	return out;
}

typedef struct nox_test_mapgen_connect_prefabs_actual_5266f0_result {
	uintptr_t replacement_address;
	uintptr_t candidate_address;
	uintptr_t connector_address;
	uintptr_t replacement_neighbor_address;
	uintptr_t candidate_neighbor_address;
	uintptr_t connector_north_neighbor_address;
	uintptr_t connector_south_neighbor_address;
	int result;
	int room_count;
} nox_test_mapgen_connect_prefabs_actual_5266f0_result;

static nox_test_mapgen_connect_prefabs_actual_5266f0_result nox_test_mapgen_connect_prefabs_actual_5266f0(void) {
	nox_test_mapgen_connect_prefabs_actual_5266f0_result out = {0};
	uint8_t* theme = (uint8_t*)calloc(1, 0x45C);
	uint8_t* prefab = (uint8_t*)calloc(1, 160);
	uint8_t* old_room = NULL;
	uint8_t* candidate = NULL;
	uint32_t saved_opposite_directions[4] = {0};
	int grid_initialized = 0;
	int directions_initialized = 0;
	if (!theme || !prefab) {
		goto cleanup;
	}
	uint32_t prefab_token = nox_mapgenLegacyPtrRegister(prefab);
	if (!prefab_token) {
		goto cleanup;
	}
	static const uint32_t opposite_directions[4] = {1, 0, 3, 2};
	for (int i = 0; i < 4; ++i) {
		saved_opposite_directions[i] = *getMemU32Ptr(0x587000, 254952 + 4 * i);
		*getMemU32Ptr(0x587000, 254952 + 4 * i) = opposite_directions[i];
	}
	directions_initialized = 1;
	*(uint32_t*)(theme + 68) = 32;
	grid_initialized = sub_520EA0(theme);
	if (!grid_initialized) {
		goto cleanup;
	}

	old_room = (uint8_t*)nox_xxx_mapGenMakeRoomStruct_521940(2, 2);
	candidate = (uint8_t*)nox_xxx_mapGenMakeRoomStruct_521940(2, 2);
	if (!old_room || !candidate) {
		goto cleanup_rooms;
	}
	float2 position = {0};
	nox_xxx_mapGenSetRoomPos_521880((uint32_t*)old_room, &position);
	if (!nox_xxx_mapGenAddNewRoom_521730((uint32_t*)old_room)) {
		goto cleanup_rooms;
	}
	uint8_t* old_room_in_list = old_room;
	old_room = NULL;
	position.field_4 = -4.0f * 32.526913f;
	nox_xxx_mapGenSetRoomPos_521880((uint32_t*)candidate, &position);
	if (!nox_xxx_mapGenAddNewRoom_521730((uint32_t*)candidate)) {
		goto cleanup_rooms;
	}
	out.candidate_address = (uintptr_t)candidate;
	candidate = NULL;

	*(uint32_t*)(theme + 80) = prefab_token;
	*(uint32_t*)(prefab + 76) = 1;
	*(int32_t*)(prefab + 80) = 1;
	*(int32_t*)(prefab + 84) = 0;
	*(uint32_t*)(prefab + 88) = 1;
	*(uint32_t*)(prefab + 92) = 1;
	*(float*)(prefab + 60) = 2.0f * 32.526913f;
	*(float*)(prefab + 64) = 2.0f * 32.526913f;
	*(uint32_t*)(prefab + 148) = nox_mapgenLegacyPtrRegister(old_room_in_list);

	out.result = nox_xxx_mapGen_InPrefab2_5266F0(theme);
	uint8_t* replacement = (uint8_t*)nox_mapgenLegacyPtrResolve(*(uint32_t*)(prefab + 148));
	out.replacement_address = (uintptr_t)replacement;
	for (uint8_t* room = (uint8_t*)nox_xxx_mapGenGetTopRoom_521710(); room;
		 room = (uint8_t*)nox_mapgenRoomNextNative_521720(room)) {
		++out.room_count;
		if (*(uint32_t*)room == 2) {
			out.connector_address = (uintptr_t)room;
		}
	}
	uint8_t* connector = (uint8_t*)out.connector_address;
	uint8_t* candidate_in_list = (uint8_t*)out.candidate_address;
	if (replacement) {
		out.replacement_neighbor_address =
			(uintptr_t)nox_mapgenLegacyPtrResolve(*(uint32_t*)(replacement + 88));
	}
	if (candidate_in_list) {
		out.candidate_neighbor_address =
			(uintptr_t)nox_mapgenLegacyPtrResolve(*(uint32_t*)(candidate_in_list + 120));
	}
	if (connector) {
		out.connector_north_neighbor_address =
			(uintptr_t)nox_mapgenLegacyPtrResolve(*(uint32_t*)(connector + 88));
		out.connector_south_neighbor_address =
			(uintptr_t)nox_mapgenLegacyPtrResolve(*(uint32_t*)(connector + 120));
	}

cleanup_rooms:
	nox_xxx_mapGenFreeTopRoom_521A40();
	if (candidate) {
		sub_521A10(candidate);
	}
	if (old_room) {
		sub_521A10(old_room);
	}
cleanup:
	if (directions_initialized) {
		for (int i = 0; i < 4; ++i) {
			*getMemU32Ptr(0x587000, 254952 + 4 * i) = saved_opposite_directions[i];
		}
	}
	if (grid_initialized) {
		sub_520F80();
	}
	if (prefab) {
		nox_mapgenLegacyPtrForget(prefab);
		free(prefab);
	}
	free(theme);
	return out;
}
*/
import "C"

type mapgenConnectPrefabsResult5266F0 struct {
	themeAddress           uintptr
	prefabAddress          uintptr
	oldRoomAddress         uintptr
	newRoomAddress         uintptr
	candidateAddresses     [3]uintptr
	callbackAddresses      [3]uintptr
	firstNextAddress       uintptr
	forwardNeighborAddress uintptr
	reverseNeighborAddress uintptr
	result                 bool
	calls                  int
	linked                 bool
	hasTypeOneNeighbor     bool
}

type mapgenConnectPrefabsActualResult5266F0 struct {
	replacementAddress            uintptr
	candidateAddress              uintptr
	connectorAddress              uintptr
	replacementNeighborAddress    uintptr
	candidateNeighborAddress      uintptr
	connectorNorthNeighborAddress uintptr
	connectorSouthNeighborAddress uintptr
	result                        bool
	roomCount                     int
}

func mapgenConnectPrefabsFixture5266F0() mapgenConnectPrefabsResult5266F0 {
	got := C.nox_test_mapgen_connect_prefabs_5266f0()
	out := mapgenConnectPrefabsResult5266F0{
		themeAddress:           uintptr(got.theme_address),
		prefabAddress:          uintptr(got.prefab_address),
		oldRoomAddress:         uintptr(got.old_room_address),
		newRoomAddress:         uintptr(got.new_room_address),
		firstNextAddress:       uintptr(got.first_next_address),
		forwardNeighborAddress: uintptr(got.forward_neighbor_address),
		reverseNeighborAddress: uintptr(got.reverse_neighbor_address),
		result:                 got.result != 0,
		calls:                  int(got.calls),
		linked:                 got.linked != 0,
		hasTypeOneNeighbor:     got.has_type_one_neighbor != 0,
	}
	for i := range out.candidateAddresses {
		out.candidateAddresses[i] = uintptr(got.candidate_addresses[i])
		out.callbackAddresses[i] = uintptr(got.callback_addresses[i])
	}
	return out
}

func mapgenConnectPrefabsActualFixture5266F0() mapgenConnectPrefabsActualResult5266F0 {
	got := C.nox_test_mapgen_connect_prefabs_actual_5266f0()
	return mapgenConnectPrefabsActualResult5266F0{
		replacementAddress:            uintptr(got.replacement_address),
		candidateAddress:              uintptr(got.candidate_address),
		connectorAddress:              uintptr(got.connector_address),
		replacementNeighborAddress:    uintptr(got.replacement_neighbor_address),
		candidateNeighborAddress:      uintptr(got.candidate_neighbor_address),
		connectorNorthNeighborAddress: uintptr(got.connector_north_neighbor_address),
		connectorSouthNeighborAddress: uintptr(got.connector_south_neighbor_address),
		result:                        got.result != 0,
		roomCount:                     int(got.room_count),
	}
}
