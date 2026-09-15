#include "map_temp_lists_503f40.h"

#include <stdlib.h>

#include "GAME1.h"
#include "GAME1_1.h"
#include "GAME1_2.h"

extern uint32_t dword_5d4594_3835368;
extern void* dword_5d4594_1599532;
extern nox_tile_coord_entry_t* dword_5d4594_1599556;
extern void* dword_5d4594_1599548;

nox_waypoint_t* sub_579E70(void);
void sub_579E90(nox_waypoint_t* waypoint);

static nox_map_wall_list_node_504290* wall_list_head_504290(void) {
	return (nox_map_wall_list_node_504290*)dword_5d4594_1599532;
}

static void wall_list_set_head_504290(nox_map_wall_list_node_504290* node) {
	dword_5d4594_1599532 = node;
}

static nox_map_waypoint_list_node_5044B0* waypoint_list_head_5044B0(void) {
	return (nox_map_waypoint_list_node_5044B0*)dword_5d4594_1599548;
}

static void waypoint_list_set_head_5044B0(nox_map_waypoint_list_node_5044B0* node) {
	dword_5d4594_1599548 = node;
}

nox_map_wall_list_node_504290* sub_504290(int8_t x, int8_t y) {
	nox_map_wall_list_node_504290* node = calloc(1, sizeof(*node));
	if (node == NULL) {
		return NULL;
	}
	node->wall = calloc(1, sizeof(*node->wall));
	if (node->wall == NULL) {
		free(node);
		return NULL;
	}
	node->wall->x = (uint8_t)x;
	node->wall->y = (uint8_t)y;
	node->next = wall_list_head_504290();
	if (node->next != NULL) {
		node->next->previous = node;
	}
	wall_list_set_head_504290(node);
	return node;
}

nox_map_wall_list_node_504290* nox_xxx_cliWallGet_5042F0(int32_t x, int32_t y) {
	for (nox_map_wall_list_node_504290* node = wall_list_head_504290(); node != NULL; node = node->next) {
		if ((int32_t)node->wall->x == x && (int32_t)node->wall->y == y) {
			return node;
		}
	}
	return NULL;
}

int32_t sub_504330(int32_t delta_x, int32_t delta_y) {
	for (nox_map_wall_list_node_504290* node = wall_list_head_504290(); node != NULL; node = node->next) {
		nox_map_wall_504290* source = node->wall;
		const int32_t x = (delta_x + 23 * (int32_t)source->x) / 23;
		const int32_t y = (delta_y + 23 * (int32_t)source->y) / 23;
		nox_map_wall_504290* wall = (nox_map_wall_504290*)nox_server_getWallAtGrid_410580(x, y);
		if (wall != NULL) {
			wall->dir = dword_5d4594_3835368 ? nox_xxx_wall_42A6C0(wall->dir, source->dir) : source->dir;
		} else {
			wall = (nox_map_wall_504290*)nox_xxx_wallCreateAt_410250(x, y);
			if (wall == NULL) {
				return 0;
			}
			wall->dir = source->dir;
		}
		wall->tile = source->tile;
		wall->variation = source->variation;
		wall->health = source->health;
		if ((source->flags & UINT8_C(0x80)) != 0) {
			wall->flags |= UINT8_C(0x80);
		}
		if (wall->variation >= nox_xxx_mapWallMaxVariation_410DD0(wall->tile, wall->dir, 0)) {
			wall->variation = 0;
		}
		if ((wall->flags & UINT8_C(0x04)) != 0) {
			sub_4107A0(wall->data);
			wall->flags &= (uint8_t)~UINT8_C(0x04);
			wall->data = NULL;
		}
		if ((source->flags & UINT8_C(0x04)) != 0) {
			// The Go wall bridge also repairs the secret's native back-pointer;
			// leaving it aimed at the temporary wall would become a UAF when the
			// temporary list is released.
			nox_server_wallAttachSecret(wall, source->data, source->field_10);
			nox_xxx_wallSecretBlock_410760((nox_secret_wall_t*)source->data);
			source->data = NULL;
		}
		if ((source->flags & UINT8_C(0x08)) != 0 && (wall->flags & UINT8_C(0x08)) == 0) {
			wall->flags |= UINT8_C(0x08);
			nox_xxx_wallBreackableListAdd_410840(wall);
		}
		if ((source->flags & UINT8_C(0x40)) != 0) {
			wall->flags |= UINT8_C(0x40);
		}
	}
	return 1;
}

nox_map_waypoint_list_node_5044B0* sub_5044B0(int32_t index, float x, float y) {
	nox_map_waypoint_list_node_5044B0* node = calloc(1, sizeof(*node));
	if (node == NULL) {
		return NULL;
	}
	node->waypoint = sub_579E70();
	if (node->waypoint == NULL) {
		free(node);
		return NULL;
	}
	node->next = waypoint_list_head_5044B0();
	if (node->next != NULL) {
		node->next->previous = node;
	}
	waypoint_list_set_head_5044B0(node);

	node->waypoint->ind = (uint32_t)index;
	node->waypoint->pos.field_0 = x;
	node->waypoint->pos.field_4 = y;
	node->waypoint->prev = NULL;
	if (node->next != NULL) {
		node->waypoint->next = node->next->waypoint;
		node->next->waypoint->prev = node->waypoint;
	} else {
		node->waypoint->next = NULL;
	}
	return node;
}

int32_t sub_504560(int32_t delta_x, int32_t delta_y) {
	const float x_offset = (float)delta_x;
	const float y_offset = (float)delta_y;
	for (nox_map_waypoint_list_node_5044B0* node = waypoint_list_head_5044B0(); node != NULL;
		 node = node->next) {
		node->waypoint->pos.field_0 += x_offset;
		node->waypoint->pos.field_4 += y_offset;
		sub_579E90(node->waypoint);
	}
	return 1;
}

nox_map_wall_504290* nox_map_wall_list_value_504290(nox_map_wall_list_node_504290* node) {
	return node == NULL ? NULL : node->wall;
}

nox_waypoint_t* nox_map_waypoint_list_value_5044B0(nox_map_waypoint_list_node_5044B0* node) {
	return node == NULL ? NULL : node->waypoint;
}

void nox_mapgen_free_waypoint_list_503F40(int32_t free_payloads) {
	nox_map_waypoint_list_node_5044B0* node = waypoint_list_head_5044B0();
	while (node != NULL) {
		nox_map_waypoint_list_node_5044B0* next = node->next;
		if (free_payloads && node->waypoint != NULL) {
			free(node->waypoint);
		}
		free(node);
		node = next;
	}
	waypoint_list_set_head_5044B0(NULL);
}

void nox_mapgen_free_tile_list_503F40(void) {
	nox_tile_coord_entry_t* entry = dword_5d4594_1599556;
	while (entry != NULL) {
		// 00504210 walks the list through prev, so the cleanup does the same.
		nox_tile_coord_entry_t* next = entry->prev;
		if (entry->layer != NULL) {
			nox_xxx_tileFreeTile_422200(&entry->layer->subtiles);
			free(entry->layer);
		}
		free(entry);
		entry = next;
	}
	dword_5d4594_1599556 = NULL;
}

void nox_mapgen_free_wall_list_503F40(int32_t free_payloads) {
	nox_map_wall_list_node_504290* node = wall_list_head_504290();
	while (node != NULL) {
		nox_map_wall_list_node_504290* next = node->next;
		if (node->wall != NULL) {
			if (free_payloads && (node->wall->flags & UINT8_C(0x04)) != 0) {
				free(node->wall->data);
			}
			free(node->wall);
		}
		free(node);
		node = next;
	}
	wall_list_set_head_504290(NULL);
}

void nox_mapgen_adjust_waypoint_506260(void* bounds, float2* position) {
	if (bounds == NULL || position == NULL) {
		return;
	}
	char* map_size = nox_xxx_mapGetWallSize_426A70();
	int4 map_origin;
	sub_428170(bounds, &map_origin);
	position->field_0 = position->field_0 - (float)(23 * *(int32_t*)map_size) + (float)map_origin.field_0 - 11.0f;
	position->field_4 = position->field_4 - (float)(23 * *((int32_t*)map_size + 1)) + (float)map_origin.field_4 - 11.0f;
}
