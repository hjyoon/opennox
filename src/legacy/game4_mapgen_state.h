#ifndef OPENNOX_GAME4_MAPGEN_STATE_H
#define OPENNOX_GAME4_MAPGEN_STATE_H

#include <stddef.h>
#include <stdio.h>
#include "map_object_list_5048a0.h"
#include "map_temp_lists_503f40.h"

typedef struct nox_game4_mapgen_state {
	uint32_t loaded_index;
	uint32_t tiles_placed;
	FILE* archive;
	void* archive_path;
	void* staging_path;
	nox_map_object_list_node_5048A0* object_head;
	nox_map_wall_list_node_504290* wall_head;
	nox_map_waypoint_list_node_5044B0* waypoint_head;
	nox_tile_coord_entry_t* tile_head;
} nox_game4_mapgen_state;

_Static_assert(offsetof(nox_game4_mapgen_state, archive) == 8,
	"mapgen index and placement status occupy two DWORDs");
_Static_assert(sizeof(nox_game4_mapgen_state) == 8 + 7 * sizeof(void*),
	"mapgen archive, path and list pointers use native width");

nox_game4_mapgen_state* nox_game4_mapgen_storage(void);

#endif
