#ifndef NOX_MAP_TEMP_LISTS_503F40_H
#define NOX_MAP_TEMP_LISTS_503F40_H

#include <stddef.h>
#include <stdint.h>

#include "defs.h"

// Temporary walls use the same scalar prefix and native-width pointer tail as
// server.Wall. GAME.EXE allocated 0x24 bytes because all five pointers were
// four bytes wide; keeping that allocation size on a 64-bit host overwrites
// the heap as soon as one of the pointer fields is used.
typedef struct nox_map_wall_504290 {
	uint8_t dir;
	uint8_t tile;
	uint8_t variation;
	uint8_t field_3;
	uint8_t flags;
	uint8_t x;
	uint8_t y;
	uint8_t health;
	uint16_t field_8;
	uint16_t field_10;
	uint32_t field_12;
	struct nox_map_wall_504290* next_by_pos;
	struct nox_map_wall_504290* next;
	struct nox_map_wall_504290* next_by_y;
	void* data;
	void* client_data;
} nox_map_wall_504290;

typedef struct nox_map_wall_list_node_504290 {
	nox_map_wall_504290* wall;
	struct nox_map_wall_list_node_504290* next;
	struct nox_map_wall_list_node_504290* previous;
} nox_map_wall_list_node_504290;

typedef struct nox_map_waypoint_list_node_5044B0 {
	nox_waypoint_t* waypoint;
	struct nox_map_waypoint_list_node_5044B0* next;
	struct nox_map_waypoint_list_node_5044B0* previous;
} nox_map_waypoint_list_node_5044B0;

_Static_assert(offsetof(nox_map_wall_504290, flags) == 4, "wrong temporary-wall flags offset");
_Static_assert(offsetof(nox_map_wall_504290, x) == 5, "wrong temporary-wall X offset");
_Static_assert(offsetof(nox_map_wall_504290, y) == 6, "wrong temporary-wall Y offset");
_Static_assert(offsetof(nox_map_wall_504290, data) == (sizeof(void*) == 4 ? 28 : 40),
	"wrong temporary-wall data offset");
_Static_assert(sizeof(nox_map_wall_504290) == (sizeof(void*) == 4 ? 36 : 56),
	"wrong native temporary-wall size");
_Static_assert(sizeof(nox_map_wall_list_node_504290) == 3 * sizeof(void*),
	"wrong native temporary-wall node size");
_Static_assert(sizeof(nox_map_waypoint_list_node_5044B0) == 3 * sizeof(void*),
	"wrong native temporary-waypoint node size");

nox_map_wall_list_node_504290* sub_504290(int8_t x, int8_t y);
nox_map_wall_list_node_504290* nox_xxx_cliWallGet_5042F0(int32_t x, int32_t y);
int32_t sub_504330(int32_t delta_x, int32_t delta_y);
nox_map_waypoint_list_node_5044B0* sub_5044B0(int32_t index, float x, float y);
int32_t sub_504560(int32_t delta_x, int32_t delta_y);

nox_map_wall_504290* nox_map_wall_list_value_504290(nox_map_wall_list_node_504290* node);
nox_waypoint_t* nox_map_waypoint_list_value_5044B0(nox_map_waypoint_list_node_5044B0* node);

void nox_mapgen_free_waypoint_list_503F40(int32_t free_payloads);
void nox_mapgen_free_tile_list_503F40(void);
void nox_mapgen_free_wall_list_503F40(int32_t free_payloads);

// Applies the placement transform used by the WayPoints section reader.
void nox_mapgen_adjust_waypoint_506260(void* bounds, float2* position);

#endif // NOX_MAP_TEMP_LISTS_503F40_H
