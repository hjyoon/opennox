#ifndef OPENNOX_FLOOR_TILE_STORAGE_H
#define OPENNOX_FLOOR_TILE_STORAGE_H

#include "defs.h"

typedef struct nox_floor_tile_storage {
	uint32_t count;
	nox_tileDef_t definitions[176];
} nox_floor_tile_storage_t;

nox_floor_tile_storage_t* nox_floor_tile_storage(void);

#endif
