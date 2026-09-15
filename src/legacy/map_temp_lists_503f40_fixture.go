package legacy

/*
#include <stdint.h>
#include <stdlib.h>
#include <string.h>

#include "GAME4.h"
#include "memmap.h"

extern void* dword_5d4594_1599532;
extern nox_tile_coord_entry_t* dword_5d4594_1599556;
extern void* dword_5d4594_1599548;

typedef struct nox_test_map_temp_lists_503f40_result {
	uintptr_t addresses[10];
	uint32_t pointer_size;
	uint32_t wall_size;
	uint32_t wall_node_size;
	uint32_t waypoint_size;
	uint32_t waypoint_node_size;
	uint32_t tile_entry_size;
	uint32_t tile_layer_size;
	uint32_t wall_links_ok;
	uint32_t wall_lookup_ok;
	uint32_t wall_values_ok;
	uint32_t wall_cleanup_ok;
	uint32_t wall_payload_preserved_ok;
	uint32_t waypoint_links_ok;
	uint32_t waypoint_values_ok;
	uint32_t waypoint_cleanup_ok;
	uint32_t waypoint_payload_preserved_ok;
	uint32_t tile_values_ok;
	uint32_t tile_cleanup_ok;
} nox_test_map_temp_lists_503f40_result;

static nox_test_map_temp_lists_503f40_result nox_test_map_temp_lists_503f40(void) {
	void* saved_walls = dword_5d4594_1599532;
	nox_tile_coord_entry_t* saved_tiles = dword_5d4594_1599556;
	void* saved_waypoints = dword_5d4594_1599548;
	uint32_t saved_tile_count = *getMemU32Ptr(0x5D4594, 1599560);
	dword_5d4594_1599532 = NULL;
	dword_5d4594_1599556 = NULL;
	dword_5d4594_1599548 = NULL;

	nox_test_map_temp_lists_503f40_result result = {
		.pointer_size = sizeof(void*),
		.wall_size = sizeof(nox_map_wall_504290),
		.wall_node_size = sizeof(nox_map_wall_list_node_504290),
		.waypoint_size = sizeof(nox_waypoint_t),
		.waypoint_node_size = sizeof(nox_map_waypoint_list_node_5044B0),
		.tile_entry_size = sizeof(nox_tile_coord_entry_t),
		.tile_layer_size = sizeof(nox_tile_layer_t),
	};

	nox_map_wall_list_node_504290* first_wall = sub_504290(17, 29);
	nox_map_wall_list_node_504290* second_wall = sub_504290(31, 43);
	if (first_wall && second_wall && first_wall->wall && second_wall->wall) {
		result.addresses[0] = (uintptr_t)first_wall;
		result.addresses[1] = (uintptr_t)first_wall->wall;
		result.addresses[2] = (uintptr_t)second_wall;
		result.addresses[3] = (uintptr_t)second_wall->wall;
		result.wall_links_ok =
			second_wall->next == first_wall && second_wall->previous == NULL &&
			first_wall->next == NULL && first_wall->previous == second_wall;
		result.wall_lookup_ok =
			nox_xxx_cliWallGet_5042F0(17, 29) == first_wall &&
			nox_xxx_cliWallGet_5042F0(31, 43) == second_wall &&
			nox_xxx_cliWallGet_5042F0(1, 2) == NULL;
		result.wall_values_ok =
			first_wall->wall->x == 17 && first_wall->wall->y == 29 &&
			second_wall->wall->x == 31 && second_wall->wall->y == 43;
		second_wall->wall->flags |= 0x04;
		second_wall->wall->data = calloc(1, 1);
	}
	nox_mapgen_free_wall_list_503F40(1);
	result.wall_cleanup_ok = dword_5d4594_1599532 == NULL;
	uint8_t* retained_wall_data = calloc(1, 1);
	nox_map_wall_list_node_504290* retained_wall = sub_504290(1, 2);
	if (retained_wall_data && retained_wall && retained_wall->wall) {
		*retained_wall_data = 0x5A;
		retained_wall->wall->flags |= 0x04;
		retained_wall->wall->data = retained_wall_data;
	}
	nox_mapgen_free_wall_list_503F40(0);
	if (retained_wall_data) {
		result.wall_payload_preserved_ok = *retained_wall_data == 0x5A;
		free(retained_wall_data);
	}

	nox_map_waypoint_list_node_5044B0* first_waypoint = sub_5044B0(101, 1.5f, -2.25f);
	nox_map_waypoint_list_node_5044B0* second_waypoint = sub_5044B0(202, 3.25f, 4.5f);
	if (first_waypoint && second_waypoint && first_waypoint->waypoint && second_waypoint->waypoint) {
		result.addresses[4] = (uintptr_t)first_waypoint;
		result.addresses[5] = (uintptr_t)first_waypoint->waypoint;
		result.addresses[6] = (uintptr_t)second_waypoint;
		result.addresses[7] = (uintptr_t)second_waypoint->waypoint;
		result.waypoint_links_ok =
			second_waypoint->next == first_waypoint && second_waypoint->previous == NULL &&
			first_waypoint->next == NULL && first_waypoint->previous == second_waypoint &&
			second_waypoint->waypoint->next == first_waypoint->waypoint &&
			second_waypoint->waypoint->prev == NULL &&
			first_waypoint->waypoint->next == NULL &&
			first_waypoint->waypoint->prev == second_waypoint->waypoint;
		result.waypoint_values_ok =
			first_waypoint->waypoint->ind == 101 && first_waypoint->waypoint->pos.field_0 == 1.5f &&
			first_waypoint->waypoint->pos.field_4 == -2.25f &&
			second_waypoint->waypoint->ind == 202 && second_waypoint->waypoint->pos.field_0 == 3.25f &&
			second_waypoint->waypoint->pos.field_4 == 4.5f &&
			(first_waypoint->waypoint->flags & 0x1000000u) != 0 &&
			(second_waypoint->waypoint->flags & 0x1000000u) != 0;
	}
	nox_mapgen_free_waypoint_list_503F40(1);
	result.waypoint_cleanup_ok = dword_5d4594_1599548 == NULL;
	nox_map_waypoint_list_node_5044B0* retained_waypoint_node = sub_5044B0(303, 6.5f, 7.75f);
	nox_waypoint_t* retained_waypoint = retained_waypoint_node ? retained_waypoint_node->waypoint : NULL;
	nox_mapgen_free_waypoint_list_503F40(0);
	if (retained_waypoint) {
		result.waypoint_payload_preserved_ok =
			retained_waypoint->ind == 303 && retained_waypoint->pos.field_0 == 6.5f &&
			retained_waypoint->pos.field_4 == 7.75f;
		free(retained_waypoint);
	}

	uint32_t kind_bits = 1;
	float kind;
	memcpy(&kind, &kind_bits, sizeof(kind));
	nox_tile_coord_entry_t* tile = nox_xxx_tileAllocTileInCoordList_5040A0(3, 4, kind);
	if (tile && tile->layer) {
		result.addresses[8] = (uintptr_t)tile;
		result.addresses[9] = (uintptr_t)tile->layer;
		result.tile_values_ok =
			tile->kind == 1 && tile->x == 161.0f && tile->y == 184.0f &&
			tile->prev == NULL && tile->next == NULL && tile->layer->subtiles == NULL;
	}
	nox_mapgen_free_tile_list_503F40();
	result.tile_cleanup_ok = dword_5d4594_1599556 == NULL;

	dword_5d4594_1599532 = saved_walls;
	dword_5d4594_1599556 = saved_tiles;
	dword_5d4594_1599548 = saved_waypoints;
	*getMemU32Ptr(0x5D4594, 1599560) = saved_tile_count;
	return result;
}
*/
import "C"

type mapTempListsResult503F40 struct {
	addresses                  [10]uintptr
	pointerSize                uint32
	wallSize                   uint32
	wallNodeSize               uint32
	waypointSize               uint32
	waypointNodeSize           uint32
	tileEntrySize              uint32
	tileLayerSize              uint32
	wallLinksOK                bool
	wallLookupOK               bool
	wallValuesOK               bool
	wallCleanupOK              bool
	wallPayloadPreservedOK     bool
	waypointLinksOK            bool
	waypointValuesOK           bool
	waypointCleanupOK          bool
	waypointPayloadPreservedOK bool
	tileValuesOK               bool
	tileCleanupOK              bool
}

func mapTempListsFixture503F40() mapTempListsResult503F40 {
	v := C.nox_test_map_temp_lists_503f40()
	var addresses [10]uintptr
	for i := range addresses {
		addresses[i] = uintptr(v.addresses[i])
	}
	return mapTempListsResult503F40{
		addresses:                  addresses,
		pointerSize:                uint32(v.pointer_size),
		wallSize:                   uint32(v.wall_size),
		wallNodeSize:               uint32(v.wall_node_size),
		waypointSize:               uint32(v.waypoint_size),
		waypointNodeSize:           uint32(v.waypoint_node_size),
		tileEntrySize:              uint32(v.tile_entry_size),
		tileLayerSize:              uint32(v.tile_layer_size),
		wallLinksOK:                v.wall_links_ok != 0,
		wallLookupOK:               v.wall_lookup_ok != 0,
		wallValuesOK:               v.wall_values_ok != 0,
		wallCleanupOK:              v.wall_cleanup_ok != 0,
		wallPayloadPreservedOK:     v.wall_payload_preserved_ok != 0,
		waypointLinksOK:            v.waypoint_links_ok != 0,
		waypointValuesOK:           v.waypoint_values_ok != 0,
		waypointCleanupOK:          v.waypoint_cleanup_ok != 0,
		waypointPayloadPreservedOK: v.waypoint_payload_preserved_ok != 0,
		tileValuesOK:               v.tile_values_ok != 0,
		tileCleanupOK:              v.tile_cleanup_ok != 0,
	}
}
