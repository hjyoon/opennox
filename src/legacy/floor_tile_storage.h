#ifndef OPENNOX_FLOOR_TILE_STORAGE_H
#define OPENNOX_FLOOR_TILE_STORAGE_H

#include "defs.h"
#include <stddef.h>

typedef struct {
	void* brush_nodes;
	uint32_t selected_edge;
	uint32_t edge_tile;
	uint32_t edge_variation;
	uint32_t brush_mode;
	uint32_t flood_count;
} nox_floor_edit_state_t;
_Static_assert(offsetof(nox_floor_edit_state_t, selected_edge) == sizeof(void*), "floor editing word offset");
nox_floor_edit_state_t* nox_floor_edit_storage(void);

typedef struct nox_floor_tile_storage {
	uint32_t count;
	nox_tileDef_t definitions[176];
} nox_floor_tile_storage_t;

nox_floor_tile_storage_t* nox_floor_tile_storage(void);

#endif
