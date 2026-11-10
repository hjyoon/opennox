#ifndef NOX_GAME4_LIST_VOTE_REFERENCE_H
#define NOX_GAME4_LIST_VOTE_REFERENCE_H
#include "game4_list_vote_native.h"

// These two declarations are diagnostic service transports only. The frozen
// reference algorithms below retain their complete original function bytes
// after reversing the recorded identifiers and public-to-static qualifier.
extern void game4_reference_free_go(void* p);
extern int game4_reference_send_packet_go(int recipient, void* packet, int size, void* related, int reliable);
static inline void game4_reference_free_bridge(void* p) { game4_reference_free_go(p); }
static inline int game4_reference_send_packet_bridge(int recipient, const void* packet, int size, nox_object_t* related, int reliable) {
 return game4_reference_send_packet_go(recipient,(void*)packet,size,related,reliable);
}
static void* game4_reference_dword_5d4594_1599532;
static void* game4_reference_dword_5d4594_1599548;
static nox_tile_coord_entry_t* game4_reference_dword_5d4594_1599556;
static void* game4_reference_dword_5d4594_1599540;
static void* game4_reference_nox_alloc_vote_1599652;
static uint32_t game4_reference_dword_5d4594_1599656;
static nox_vote_5066D0* game4_reference_vote_head_5066D0;
static nox_map_wall_list_node_504290* game4_reference_wall_list_head_504290(void);
static void game4_reference_wall_list_set_head_504290(nox_map_wall_list_node_504290* node);
static nox_map_waypoint_list_node_5044B0* game4_reference_waypoint_list_head_5044B0(void);
static void game4_reference_waypoint_list_set_head_5044B0(nox_map_waypoint_list_node_5044B0* node);
static nox_map_wall_list_node_504290* game4_reference_sub_504290(int8_t x, int8_t y);
static nox_map_wall_list_node_504290* game4_reference_nox_xxx_cliWallGet_5042F0(int32_t x, int32_t y);
static int32_t game4_reference_sub_504330(int32_t delta_x, int32_t delta_y);
static nox_map_waypoint_list_node_5044B0* game4_reference_sub_5044B0(int32_t index, float x, float y);
static int32_t game4_reference_sub_504560(int32_t delta_x, int32_t delta_y);
static nox_map_wall_504290* game4_reference_nox_map_wall_list_value_504290(nox_map_wall_list_node_504290* node);
static nox_waypoint_t* game4_reference_nox_map_waypoint_list_value_5044B0(nox_map_waypoint_list_node_5044B0* node);
static void game4_reference_nox_mapgen_free_waypoint_list_503F40(int32_t free_payloads);
static void game4_reference_nox_mapgen_free_tile_list_503F40(void);
static void game4_reference_nox_mapgen_free_wall_list_503F40(int32_t free_payloads);
static void game4_reference_nox_mapgen_adjust_waypoint_506260(void* bounds, float2* position);
static nox_map_object_list_node_5048A0* game4_reference_map_object_list_head_5048A0(void);
static void game4_reference_map_object_list_set_head_5048A0(nox_map_object_list_node_5048A0* node);
static nox_map_object_list_node_5048A0* game4_reference_nox_xxx_unitAddToList_5048A0(nox_object_t* object);
static int32_t game4_reference_sub_504910(int32_t delta_x, int32_t delta_y);
static nox_object_t* game4_reference_sub_504980(void);
static nox_object_t* game4_reference_sub_5049C0(nox_object_t* object);
static nox_map_object_list_node_5048A0* game4_reference_sub_5049D0(void);
static nox_map_object_list_node_5048A0* game4_reference_sub_5049E0(nox_map_object_list_node_5048A0* node);
static int32_t game4_reference_sub_504A10(nox_object_t* object);
static nox_object_t* game4_reference_nox_map_object_list_node_object_5048A0(nox_map_object_list_node_5048A0* node);
static nox_map_object_list_node_5048A0* game4_reference_nox_map_object_list_node_next_5048A0(nox_map_object_list_node_5048A0* node);
static void game4_reference_nox_map_object_list_node_free_5048A0(nox_map_object_list_node_5048A0* node);
static nox_playerInfo* game4_reference_vote_player_5066D0(nox_object_t* unit);
static nox_object_team_t* game4_reference_vote_object_team_5066D0(nox_object_t* object);
static nox_playerInfo* game4_reference_vote_find_player_5066D0(const wchar2_t* name);
static nox_vote_5066D0* game4_reference_vote_find_5066D0(uint32_t type, nox_object_t* target);
static char game4_reference_vote_add_voter_5066D0(nox_vote_5066D0* vote, uint32_t mask);
static void game4_reference_vote_remove_voter_5066D0(nox_vote_5066D0* vote, uint32_t mask);
static int game4_reference_nox_xxx_allocVoteArray_5066D0(void);
static void game4_reference_sub_506700(void);
static int game4_reference_sub_506720(void);
static int game4_reference_sub_506740(nox_object_t* player_unit);
static void game4_reference_sub_5067B0(nox_vote_5066D0* vote);
static nox_vote_5066D0* game4_reference_sub_506810(nox_vote_5066D0* vote);
static int game4_reference_nox_xxx_netSendVote_506840(int player_index);
static char game4_reference_sub_506870(int type, nox_object_t* player_unit, wchar2_t* player_name);
static char game4_reference_sub_5068E0(int type, nox_object_t* player_unit, wchar2_t* player_name);
static nox_vote_5066D0* game4_reference_nox_vote_record_alloc_5066D0(void);
static nox_vote_5066D0* game4_reference_sub_506A20(int type, nox_object_t* player_unit);
static nox_vote_5066D0* game4_reference_nox_xxx_voteAddMB_506AD0(nox_vote_5066D0* vote);
static nox_vote_5066D0* game4_reference_sub_506B00(int type, nox_object_t* player_unit);
static nox_vote_5066D0* game4_reference_sub_506B80(int type, nox_object_t* player_unit, wchar2_t* player_name);
static void game4_reference_sub_506C90(int type, nox_object_t* player_unit, wchar2_t* player_name);
static void game4_reference_sub_506D00(nox_object_t* player_unit, wchar2_t* player_name);
static void game4_reference_sub_506DE0(nox_object_t* player_unit);
static void game4_reference_sub_506E50(nox_object_t* player_unit, wchar2_t* player_name);
static void game4_reference_nox_xxx_voteUptate_506F30(void);
static void game4_reference_sub_506F80(nox_vote_5066D0* vote);
static int game4_reference_sub_507000(nox_vote_5066D0* vote);
static uint32_t game4_reference_vote_quest_player_count_5066D0(void);
static void game4_reference_sub_507090(nox_vote_5066D0* vote);
static void game4_reference_sub_507100(nox_vote_5066D0* vote);
static int game4_reference_sub_507190(int recipient, char active);
static int game4_reference_sub_5071C0(void);
static nox_vote_5066D0* game4_reference_nox_vote_head_5066D0(void);
static void* game4_reference_nox_vote_allocator_5066D0(void);
static size_t game4_reference_nox_vote_record_size_5066D0(void);
static nox_vote_5066D0* game4_reference_nox_vote_next_5066D0(nox_vote_5066D0* vote);
static nox_vote_5066D0* game4_reference_nox_vote_previous_5066D0(nox_vote_5066D0* vote);
static nox_object_team_t* game4_reference_nox_vote_team_5066D0(nox_vote_5066D0* vote);
static uint32_t game4_reference_nox_vote_frame_5066D0(nox_vote_5066D0* vote);
static uint32_t game4_reference_nox_vote_voters_5066D0(nox_vote_5066D0* vote);
static uint8_t game4_reference_nox_vote_count_5066D0(nox_vote_5066D0* vote);
static uint8_t game4_reference_nox_vote_threshold_5066D0(nox_vote_5066D0* vote);
static void game4_reference_nox_vote_set_voters_5066D0(nox_vote_5066D0* vote, uint32_t voters, uint8_t count);

static nox_map_wall_list_node_504290* game4_reference_wall_list_head_504290(void) {
	return (nox_map_wall_list_node_504290*)game4_reference_dword_5d4594_1599532;
}

static void game4_reference_wall_list_set_head_504290(nox_map_wall_list_node_504290* node) {
	game4_reference_dword_5d4594_1599532 = node;
}

static nox_map_waypoint_list_node_5044B0* game4_reference_waypoint_list_head_5044B0(void) {
	return (nox_map_waypoint_list_node_5044B0*)game4_reference_dword_5d4594_1599548;
}

static void game4_reference_waypoint_list_set_head_5044B0(nox_map_waypoint_list_node_5044B0* node) {
	game4_reference_dword_5d4594_1599548 = node;
}

static nox_map_wall_list_node_504290* game4_reference_sub_504290(int8_t x, int8_t y) {
	nox_map_wall_list_node_504290* node = calloc(1, sizeof(*node));
	if (node == NULL) {
		return NULL;
	}
	node->wall = calloc(1, sizeof(*node->wall));
	if (node->wall == NULL) {
		game4_reference_free_bridge(node);
		return NULL;
	}
	node->wall->x = (uint8_t)x;
	node->wall->y = (uint8_t)y;
	node->next = game4_reference_wall_list_head_504290();
	if (node->next != NULL) {
		node->next->previous = node;
	}
	game4_reference_wall_list_set_head_504290(node);
	return node;
}

static nox_map_wall_list_node_504290* game4_reference_nox_xxx_cliWallGet_5042F0(int32_t x, int32_t y) {
	for (nox_map_wall_list_node_504290* node = game4_reference_wall_list_head_504290(); node != NULL; node = node->next) {
		if ((int32_t)node->wall->x == x && (int32_t)node->wall->y == y) {
			return node;
		}
	}
	return NULL;
}

static int32_t game4_reference_sub_504330(int32_t delta_x, int32_t delta_y) {
	for (nox_map_wall_list_node_504290* node = game4_reference_wall_list_head_504290(); node != NULL; node = node->next) {
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

static nox_map_waypoint_list_node_5044B0* game4_reference_sub_5044B0(int32_t index, float x, float y) {
	nox_map_waypoint_list_node_5044B0* node = calloc(1, sizeof(*node));
	if (node == NULL) {
		return NULL;
	}
	node->waypoint = sub_579E70();
	if (node->waypoint == NULL) {
		game4_reference_free_bridge(node);
		return NULL;
	}
	node->next = game4_reference_waypoint_list_head_5044B0();
	if (node->next != NULL) {
		node->next->previous = node;
	}
	game4_reference_waypoint_list_set_head_5044B0(node);

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

static int32_t game4_reference_sub_504560(int32_t delta_x, int32_t delta_y) {
	const float x_offset = (float)delta_x;
	const float y_offset = (float)delta_y;
	for (nox_map_waypoint_list_node_5044B0* node = game4_reference_waypoint_list_head_5044B0(); node != NULL;
		 node = node->next) {
		node->waypoint->pos.field_0 += x_offset;
		node->waypoint->pos.field_4 += y_offset;
		sub_579E90(node->waypoint);
	}
	return 1;
}

static nox_map_wall_504290* game4_reference_nox_map_wall_list_value_504290(nox_map_wall_list_node_504290* node) {
	return node == NULL ? NULL : node->wall;
}

static nox_waypoint_t* game4_reference_nox_map_waypoint_list_value_5044B0(nox_map_waypoint_list_node_5044B0* node) {
	return node == NULL ? NULL : node->waypoint;
}

static void game4_reference_nox_mapgen_free_waypoint_list_503F40(int32_t free_payloads) {
	nox_map_waypoint_list_node_5044B0* node = game4_reference_waypoint_list_head_5044B0();
	while (node != NULL) {
		nox_map_waypoint_list_node_5044B0* next = node->next;
		if (free_payloads && node->waypoint != NULL) {
			game4_reference_free_bridge(node->waypoint);
		}
		game4_reference_free_bridge(node);
		node = next;
	}
	game4_reference_waypoint_list_set_head_5044B0(NULL);
}

static void game4_reference_nox_mapgen_free_tile_list_503F40(void) {
	nox_tile_coord_entry_t* entry = game4_reference_dword_5d4594_1599556;
	while (entry != NULL) {
		// 00504210 walks the list through prev, so the cleanup does the same.
		nox_tile_coord_entry_t* next = entry->prev;
		if (entry->layer != NULL) {
			nox_xxx_tileFreeTile_422200(&entry->layer->subtiles);
			game4_reference_free_bridge(entry->layer);
		}
		game4_reference_free_bridge(entry);
		entry = next;
	}
	game4_reference_dword_5d4594_1599556 = NULL;
}

static void game4_reference_nox_mapgen_free_wall_list_503F40(int32_t free_payloads) {
	nox_map_wall_list_node_504290* node = game4_reference_wall_list_head_504290();
	while (node != NULL) {
		nox_map_wall_list_node_504290* next = node->next;
		if (node->wall != NULL) {
			if (free_payloads && (node->wall->flags & UINT8_C(0x04)) != 0) {
				game4_reference_free_bridge(node->wall->data);
			}
			game4_reference_free_bridge(node->wall);
		}
		game4_reference_free_bridge(node);
		node = next;
	}
	game4_reference_wall_list_set_head_504290(NULL);
}

static void game4_reference_nox_mapgen_adjust_waypoint_506260(void* bounds, float2* position) {
	if (bounds == NULL || position == NULL) {
		return;
	}
	char* map_size = nox_xxx_mapGetWallSize_426A70();
	int4 map_origin;
	sub_428170(bounds, &map_origin);
	position->field_0 = position->field_0 - (float)(23 * *(int32_t*)map_size) + (float)map_origin.field_0 - 11.0f;
	position->field_4 = position->field_4 - (float)(23 * *((int32_t*)map_size + 1)) + (float)map_origin.field_4 - 11.0f;
}

static nox_map_object_list_node_5048A0* game4_reference_map_object_list_head_5048A0(void) {
	return (nox_map_object_list_node_5048A0*)game4_reference_dword_5d4594_1599540;
}

static void game4_reference_map_object_list_set_head_5048A0(nox_map_object_list_node_5048A0* node) {
	game4_reference_dword_5d4594_1599540 = node;
}

static nox_map_object_list_node_5048A0* game4_reference_nox_xxx_unitAddToList_5048A0(nox_object_t* object) {
	nox_map_object_list_node_5048A0* node = calloc(1, sizeof(*node));
	if (node == NULL) {
		return NULL;
	}
	node->previous = NULL;
	node->object = object;
	node->next = game4_reference_map_object_list_head_5048A0();
	if (node->next != NULL) {
		node->next->previous = node;
	}
	game4_reference_map_object_list_set_head_5048A0(node);

	object->object_prev = NULL;
	if (node->next != NULL) {
		object->object_next = node->next->object;
		node->next->object->object_prev = object;
	} else {
		object->object_next = NULL;
	}
	return node;
}

static int32_t game4_reference_sub_504910(int32_t delta_x, int32_t delta_y) {
	const float x_offset = (float)delta_x;
	const float y_offset = (float)delta_y;
	for (nox_map_object_list_node_5048A0* node = game4_reference_map_object_list_head_5048A0(); node != NULL;
		 node = node->next) {
		nox_object_t* object = node->object;
		object->x = object->x + x_offset;
		object = node->object;
		object->y = object->y + y_offset;
		object = node->object;
		const float y = object->y;
		const float x = object->x;
		nox_xxx_createAt_4DAA50(object, NULL, x, y);
		object = node->object;
		object->obj_flags |= UINT32_C(0x80000000);
	}
	return 1;
}

static nox_object_t* game4_reference_sub_504980(void) {
	if (!((dword_5d4594_1599480 == dword_5d4594_3835396 &&
		  (int32_t)dword_5d4594_1599480 != -1 && dword_5d4594_1599476 != 1) ||
		 nox_xxx_mapgenSaveMap_503830((int32_t)dword_5d4594_3835396))) {
		return NULL;
	}
	nox_map_object_list_node_5048A0* node = game4_reference_map_object_list_head_5048A0();
	if (node == NULL) {
		return NULL;
	}
	return node->object;
}

static nox_object_t* game4_reference_sub_5049C0(nox_object_t* object) {
	if (object == NULL) {
		return NULL;
	}
	return object->object_next;
}

static nox_map_object_list_node_5048A0* game4_reference_sub_5049D0(void) {
	return game4_reference_map_object_list_head_5048A0();
}

static nox_map_object_list_node_5048A0* game4_reference_sub_5049E0(nox_map_object_list_node_5048A0* node) {
	if (node == NULL) {
		return NULL;
	}
	return node->next;
}

static int32_t game4_reference_sub_504A10(nox_object_t* object) {
	if (object == NULL) {
		return 0;
	}
	nox_map_object_list_node_5048A0* node = game4_reference_map_object_list_head_5048A0();
	while (node != NULL && node->object != object) {
		node = node->next;
	}
	if (node == NULL) {
		return 0;
	}

	if (node->next != NULL) {
		node->next->previous = node->previous;
	}
	if (node->previous != NULL) {
		node->previous->next = node->next;
	}
	if (node == game4_reference_map_object_list_head_5048A0()) {
		game4_reference_map_object_list_set_head_5048A0(node->next);
	}
	if (object->object_next != NULL) {
		object->object_next->object_prev = object->object_prev;
	}
	if (object->object_prev != NULL) {
		object->object_prev->object_next = object->object_next;
	}
	nox_xxx_objectFreeMem_4E38A0(object);
	game4_reference_free_bridge(node);
	return 1;
}

static nox_object_t* game4_reference_nox_map_object_list_node_object_5048A0(nox_map_object_list_node_5048A0* node) {
	return node->object;
}

static nox_map_object_list_node_5048A0* game4_reference_nox_map_object_list_node_next_5048A0(nox_map_object_list_node_5048A0* node) {
	return node->next;
}

static void game4_reference_nox_map_object_list_node_free_5048A0(nox_map_object_list_node_5048A0* node) {
	game4_reference_free_bridge(node);
}

static nox_playerInfo* game4_reference_vote_player_5066D0(nox_object_t* unit) {
	if (unit == NULL || !(unit->obj_class & 4) || unit->data_update == NULL) {
		return NULL;
	}
	nox_player_update_data_t* update = (nox_player_update_data_t*)unit->data_update;
	return update->player;
}

static nox_object_team_t* game4_reference_vote_object_team_5066D0(nox_object_t* object) {
	return (nox_object_team_t*)&object->field_12;
}

static nox_playerInfo* game4_reference_vote_find_player_5066D0(const wchar2_t* name) {
	if (name == NULL) {
		return NULL;
	}
	for (nox_playerInfo* player = nox_common_playerInfoGetFirst_416EA0(); player != NULL;
		 player = nox_common_playerInfoGetNext_416EE0(player)) {
		if (player->active == 1 && nox_wcscmp(player->name_final, name) == 0) {
			return player;
		}
	}
	return NULL;
}

static nox_vote_5066D0* game4_reference_vote_find_5066D0(uint32_t type, nox_object_t* target) {
	for (nox_vote_5066D0* vote = game4_reference_vote_head_5066D0; vote != NULL; vote = vote->next) {
		if (vote->type == type && (target == NULL || vote->target == target)) {
			return vote;
		}
	}
	return NULL;
}

static char game4_reference_vote_add_voter_5066D0(nox_vote_5066D0* vote, uint32_t mask) {
	uint32_t result = vote->voters;
	if (!(result & mask)) {
		vote->voters |= mask;
		vote->count++;
		result = vote->count;
	}
	return (char)result;
}

static void game4_reference_vote_remove_voter_5066D0(nox_vote_5066D0* vote, uint32_t mask) {
	if (vote == NULL || !(vote->voters & mask)) {
		return;
	}
	vote->voters &= ~mask;
	vote->count--;
	if (vote->count == 0) {
		game4_reference_sub_5067B0(vote);
	}
}

static int game4_reference_nox_xxx_allocVoteArray_5066D0(void) {
	nox_alloc_class* allocator = nox_new_alloc_class("VoteClass", (int)sizeof(nox_vote_5066D0), 64);
	game4_reference_nox_alloc_vote_1599652 = allocator;
	if (allocator == NULL) {
		return 0;
	}
	game4_reference_vote_head_5066D0 = NULL;
	// The original PE32 global cannot represent a native pointer. Keep it
	// cleared for old diagnostic readers while all live users use the head above.
	game4_reference_dword_5d4594_1599656 = 0;
	return 1;
}

static void game4_reference_sub_506700(void) {
	if (game4_reference_nox_alloc_vote_1599652 != NULL) {
		nox_alloc_class_free_all((nox_alloc_class*)game4_reference_nox_alloc_vote_1599652);
	}
	game4_reference_vote_head_5066D0 = NULL;
	game4_reference_dword_5d4594_1599656 = 0;
}

static int game4_reference_sub_506720(void) {
	if (game4_reference_nox_alloc_vote_1599652 != NULL) {
		nox_free_alloc_class((nox_alloc_class*)game4_reference_nox_alloc_vote_1599652);
	}
	game4_reference_nox_alloc_vote_1599652 = NULL;
	game4_reference_vote_head_5066D0 = NULL;
	game4_reference_dword_5d4594_1599656 = 0;
	return 0;
}

static int game4_reference_sub_506740(nox_object_t* player_unit) {
	nox_playerInfo* player = game4_reference_vote_player_5066D0(player_unit);
	if (player == NULL || player->playerInd >= 32) {
		return 0;
	}
	const uint32_t player_mask = UINT32_C(1) << player->playerInd;
	for (nox_vote_5066D0* vote = game4_reference_vote_head_5066D0; vote != NULL;) {
		nox_vote_5066D0* next = vote->next;
		if (vote->voters & player_mask) {
			vote->voters &= ~player_mask;
			vote->count--;
		}
		// GAME.EXE also removes an already-empty record even when this
		// player's bit was not set.
		if (vote->count == 0) {
			game4_reference_sub_5067B0(vote);
		}
		vote = next;
	}
	return 0;
}

static void game4_reference_sub_5067B0(nox_vote_5066D0* vote) {
	if (vote == NULL) {
		return;
	}
	if (vote->type == 2) {
		for (int player_index = 0; player_index < 32; player_index++) {
			if (vote->voters & (UINT32_C(1) << player_index)) {
				game4_reference_nox_xxx_netSendVote_506840(player_index);
			}
		}
	}
	game4_reference_sub_506810(vote);
	if (game4_reference_nox_alloc_vote_1599652 != NULL) {
		nox_alloc_class_free_obj_first((nox_alloc_class*)game4_reference_nox_alloc_vote_1599652, vote);
	}
	if (game4_reference_vote_head_5066D0 == NULL) {
		game4_reference_sub_507190(255, 0);
	}
}

static nox_vote_5066D0* game4_reference_sub_506810(nox_vote_5066D0* vote) {
	if (vote == NULL) {
		return NULL;
	}
	nox_vote_5066D0* result = vote;
	if (vote->next != NULL) {
		vote->next->previous = vote->previous;
	}
	if (vote->previous != NULL) {
		result = vote->next;
		vote->previous->next = vote->next;
	} else {
		game4_reference_vote_head_5066D0 = vote->next;
	}
	return result;
}

static int game4_reference_nox_xxx_netSendVote_506840(int player_index) {
	const uint8_t packet[2] = {0xEE, 7};
	return game4_reference_send_packet_bridge(player_index, packet, sizeof(packet), NULL, 1);
}

static char game4_reference_sub_506870(int type, nox_object_t* player_unit, wchar2_t* player_name) {
	if (player_unit == NULL || !(player_unit->obj_class & 4)) {
		return 0;
	}
	switch (type) {
	case 0:
	case 1:
		return game4_reference_sub_5068E0(type, player_unit, player_name);
	case 2:
		return game4_reference_sub_506B00(type, player_unit) != NULL;
	case 3:
		return game4_reference_sub_506B80(type, player_unit, player_name) != NULL;
	default:
		return 0;
	}
}

static char game4_reference_sub_5068E0(int type, nox_object_t* player_unit, wchar2_t* player_name) {
	const uint32_t minimum_votes = *getMemU32Ptr(0x587000, 229980);
	if (minimum_votes == 0 || minimum_votes > 32 || player_name == NULL) {
		return (char)minimum_votes;
	}
	nox_playerInfo* voter = game4_reference_vote_player_5066D0(player_unit);
	if (voter == NULL || voter->playerInd >= 32) {
		return 0;
	}
	nox_playerInfo* target_player = game4_reference_vote_find_player_5066D0(player_name);
	if (target_player == NULL || target_player->playerInd == 31 || target_player->playerUnit == NULL ||
		target_player->playerUnit == player_unit) {
		return 0;
	}
	nox_object_t* target = target_player->playerUnit;
	const int team_mode = nox_xxx_CheckGameplayFlags_417DA0(4);
	if (team_mode && !nox_xxx_servCompareTeams_419150(
				 game4_reference_vote_object_team_5066D0(player_unit), game4_reference_vote_object_team_5066D0(target))) {
		return 0;
	}
	nox_vote_5066D0* vote = game4_reference_vote_find_5066D0((uint32_t)type, target);
	if (vote == NULL) {
		vote = game4_reference_sub_506A20(type, player_unit);
		if (vote == NULL) {
			return 0;
		}
		vote->target = target;
		vote->team_mode = team_mode ? 1u : 0u;
	}
	return game4_reference_vote_add_voter_5066D0(vote, UINT32_C(1) << voter->playerInd);
}

static nox_vote_5066D0* game4_reference_nox_vote_record_alloc_5066D0(void) {
	if (game4_reference_nox_alloc_vote_1599652 == NULL) {
		return NULL;
	}
	return (nox_vote_5066D0*)nox_alloc_class_new_obj_zero((nox_alloc_class*)game4_reference_nox_alloc_vote_1599652);
}

static nox_vote_5066D0* game4_reference_sub_506A20(int type, nox_object_t* player_unit) {
	if (player_unit == NULL || !(player_unit->obj_class & 4)) {
		return NULL;
	}
	const int was_empty = game4_reference_vote_head_5066D0 == NULL;
	nox_vote_5066D0* vote = game4_reference_nox_vote_record_alloc_5066D0();
	if (vote == NULL) {
		return NULL;
	}
	vote->type = (uint32_t)type;
	vote->frame = gameFrame();
	vote->team = game4_reference_vote_object_team_5066D0(player_unit);
	switch (type) {
	case 0:
	case 1:
		vote->threshold = getMemByte(0x587000, 229980);
		break;
	case 2:
	case 3:
		vote->threshold = 6;
		break;
	default:
		vote->threshold = getMemByte(0x587000, 229984);
		break;
	}
	game4_reference_nox_xxx_voteAddMB_506AD0(vote);
	if (was_empty) {
		game4_reference_sub_507190(255, 1);
	}
	return vote;
}

static nox_vote_5066D0* game4_reference_nox_xxx_voteAddMB_506AD0(nox_vote_5066D0* vote) {
	if (vote == NULL) {
		return NULL;
	}
	vote->previous = NULL;
	vote->next = game4_reference_vote_head_5066D0;
	if (game4_reference_vote_head_5066D0 != NULL) {
		game4_reference_vote_head_5066D0->previous = vote;
	}
	game4_reference_vote_head_5066D0 = vote;
	return vote;
}

static nox_vote_5066D0* game4_reference_sub_506B00(int type, nox_object_t* player_unit) {
	if (nox_server_resetQuestMinVotes_229988 == 0) {
		return NULL;
	}
	nox_playerInfo* voter = game4_reference_vote_player_5066D0(player_unit);
	if (voter == NULL || voter->playerInd >= 32 || voter->field_4792 == 0) {
		return NULL;
	}
	nox_vote_5066D0* vote = game4_reference_vote_find_5066D0((uint32_t)type, NULL);
	if (vote == NULL) {
		vote = game4_reference_sub_506A20(type, player_unit);
		if (vote == NULL) {
			return NULL;
		}
		vote->team_mode = 0;
	}
	game4_reference_vote_add_voter_5066D0(vote, UINT32_C(1) << voter->playerInd);
	return vote;
}

static nox_vote_5066D0* game4_reference_sub_506B80(int type, nox_object_t* player_unit, wchar2_t* player_name) {
	if (nox_server_kickQuestPlayerMinVotes_229992 == 0 || player_name == NULL) {
		return NULL;
	}
	nox_playerInfo* voter = game4_reference_vote_player_5066D0(player_unit);
	if (voter == NULL || voter->playerInd >= 32 || voter->field_4792 == 0) {
		return NULL;
	}
	nox_playerInfo* target_player = game4_reference_vote_find_player_5066D0(player_name);
	if (target_player == NULL || target_player->playerInd == 31 || target_player->field_4792 == 0 ||
		target_player->playerUnit == NULL || target_player->playerUnit == player_unit) {
		return NULL;
	}
	nox_object_t* target = target_player->playerUnit;
	nox_vote_5066D0* vote = game4_reference_vote_find_5066D0((uint32_t)type, target);
	if (vote == NULL) {
		vote = game4_reference_sub_506A20(type, player_unit);
		if (vote == NULL) {
			return NULL;
		}
		vote->target = target;
	}
	game4_reference_vote_add_voter_5066D0(vote, UINT32_C(1) << voter->playerInd);
	return vote;
}

static void game4_reference_sub_506C90(int type, nox_object_t* player_unit, wchar2_t* player_name) {
	if (player_unit == NULL || !(player_unit->obj_class & 4)) {
		return;
	}
	switch (type) {
	case 0:
	case 1:
		game4_reference_sub_506D00(player_unit, player_name);
		break;
	case 2:
		game4_reference_sub_506DE0(player_unit);
		break;
	case 3:
		game4_reference_sub_506E50(player_unit, player_name);
		break;
	default:
		break;
	}
}

static void game4_reference_sub_506D00(nox_object_t* player_unit, wchar2_t* player_name) {
	nox_playerInfo* voter = game4_reference_vote_player_5066D0(player_unit);
	nox_playerInfo* target_player = game4_reference_vote_find_player_5066D0(player_name);
	if (voter == NULL || voter->playerInd >= 32 || target_player == NULL ||
		target_player->playerInd == 31 || target_player->playerUnit == NULL) {
		return;
	}
	const uint32_t mask = UINT32_C(1) << voter->playerInd;
	for (nox_vote_5066D0* vote = game4_reference_vote_head_5066D0; vote != NULL; vote = vote->next) {
		// GAME.EXE routes both remove opcodes 0 and 1 through the type-0 lookup.
		if (vote->type == 0 && vote->target == target_player->playerUnit && (vote->voters & mask)) {
			game4_reference_vote_remove_voter_5066D0(vote, mask);
			return;
		}
	}
}

static void game4_reference_sub_506DE0(nox_object_t* player_unit) {
	nox_playerInfo* voter = game4_reference_vote_player_5066D0(player_unit);
	if (voter == NULL || voter->playerInd >= 32) {
		return;
	}
	const uint32_t mask = UINT32_C(1) << voter->playerInd;
	for (nox_vote_5066D0* vote = game4_reference_vote_head_5066D0; vote != NULL; vote = vote->next) {
		if (vote->type == 2) {
			game4_reference_vote_remove_voter_5066D0(vote, mask);
			return;
		}
	}
}

static void game4_reference_sub_506E50(nox_object_t* player_unit, wchar2_t* player_name) {
	nox_playerInfo* voter = game4_reference_vote_player_5066D0(player_unit);
	nox_playerInfo* target_player = game4_reference_vote_find_player_5066D0(player_name);
	if (voter == NULL || voter->playerInd >= 32 || target_player == NULL ||
		target_player->playerInd == 31 || target_player->playerUnit == NULL) {
		return;
	}
	const uint32_t mask = UINT32_C(1) << voter->playerInd;
	for (nox_vote_5066D0* vote = game4_reference_vote_head_5066D0; vote != NULL; vote = vote->next) {
		if (vote->type == 3 && vote->target == target_player->playerUnit && (vote->voters & mask)) {
			game4_reference_vote_remove_voter_5066D0(vote, mask);
			return;
		}
	}
}

static void game4_reference_nox_xxx_voteUptate_506F30(void) {
	for (nox_vote_5066D0* vote = game4_reference_vote_head_5066D0; vote != NULL;) {
		nox_vote_5066D0* next = vote->next;
		switch (vote->type) {
		case 0:
		case 1:
			game4_reference_sub_506F80(vote);
			break;
		case 2:
			game4_reference_sub_507090(vote);
			break;
		case 3:
			game4_reference_sub_507100(vote);
			break;
		default:
			break;
		}
		vote = next;
	}
}

static void game4_reference_sub_506F80(nox_vote_5066D0* vote) {
	if (vote == NULL || vote->target == NULL || (vote->target->obj_flags & 0x20)) {
		game4_reference_sub_5067B0(vote);
		return;
	}
	vote->team = game4_reference_vote_object_team_5066D0(vote->target);
	if (game4_reference_sub_507000(vote) != 1) {
		return;
	}
	nox_playerInfo* target_player = game4_reference_vote_player_5066D0(vote->target);
	if (target_player != NULL) {
		nox_xxx_playerCallDisconnect_4DEAB0(target_player->playerInd, 4);
		sub_416770(15, target_player->name_final, target_player->serial);
	}
	game4_reference_sub_5067B0(vote);
}

static int game4_reference_sub_507000(nox_vote_5066D0* vote) {
	if (vote == NULL) {
		return 0;
	}
	if (vote->count >= vote->threshold) {
		return 1;
	}
	uint32_t player_count = 0;
	for (nox_object_t* player = nox_xxx_getFirstPlayerUnit_4DA7C0(); player != NULL;
		 player = nox_xxx_getNextPlayerUnit_4DA7F0(player)) {
		if (vote->team_mode != 1 ||
			nox_xxx_servCompareTeams_419150(vote->team, game4_reference_vote_object_team_5066D0(player))) {
			player_count++;
		}
	}
	return vote->count >= (uint32_t)(player_count - 1) && vote->count >= 2;
}

static uint32_t game4_reference_vote_quest_player_count_5066D0(void) {
	uint32_t count = 0;
	for (nox_object_t* unit = nox_xxx_getFirstPlayerUnit_4DA7C0(); unit != NULL;
		 unit = nox_xxx_getNextPlayerUnit_4DA7F0(unit)) {
		nox_playerInfo* player = game4_reference_vote_player_5066D0(unit);
		if (player == NULL) {
			continue;
		}
		if ((!nox_common_gameFlags_check_40A5C0(1) ||
			 !nox_common_getEngineFlag(NOX_ENGINE_FLAG_DISABLE_GRAPHICS_RENDERING) ||
			 player->playerInd != 31) &&
			player->field_4792 == 1) {
			count++;
		}
	}
	return count;
}

static void game4_reference_sub_507090(nox_vote_5066D0* vote) {
	if (vote == NULL || vote->count < game4_reference_vote_quest_player_count_5066D0()) {
		return;
	}
	for (nox_object_t* unit = nox_xxx_getFirstPlayerUnit_4DA7C0(); unit != NULL;
		 unit = nox_xxx_getNextPlayerUnit_4DA7F0(unit)) {
		nox_playerInfo* player = game4_reference_vote_player_5066D0(unit);
		if (player != NULL && player->field_4792 == 1) {
			nox_xxx_playerRespawn_4F7EF0(player->playerUnit);
		}
	}
	nox_game_setQuestStage_4E3CD0(0);
	nox_xxx_mapLoad_4D2450(nox_xxx_getQuestMapFile_4D0F60());
	game4_reference_sub_5067B0(vote);
}

static void game4_reference_sub_507100(nox_vote_5066D0* vote) {
	if (vote == NULL || vote->target == NULL || (vote->target->obj_flags & 0x20)) {
		game4_reference_sub_5067B0(vote);
		return;
	}
	nox_playerInfo* target_player = game4_reference_vote_player_5066D0(vote->target);
	if (target_player == NULL || target_player->field_4792 == 0) {
		game4_reference_sub_5067B0(vote);
		return;
	}
	if (vote->count < vote->threshold) {
		const uint32_t player_count = game4_reference_vote_quest_player_count_5066D0();
		if (player_count <= 1) {
			game4_reference_sub_5067B0(vote);
			return;
		}
		if (!(vote->count >= player_count - 1 && vote->count >= 2)) {
			return;
		}
	}
	sub_4DCFB0(vote->target);
	sub_416770(15, target_player->name_final, target_player->serial);
	game4_reference_sub_5067B0(vote);
}

static int game4_reference_sub_507190(int recipient, char active) {
	const uint8_t packet[3] = {0xEE, 6, (uint8_t)active};
	return game4_reference_send_packet_bridge(recipient, packet, sizeof(packet), NULL, 1);
}

static int game4_reference_sub_5071C0(void) {
	return game4_reference_vote_head_5066D0 != NULL;
}

static nox_vote_5066D0* game4_reference_nox_vote_head_5066D0(void) {
	return game4_reference_vote_head_5066D0;
}

static void* game4_reference_nox_vote_allocator_5066D0(void) {
	return game4_reference_nox_alloc_vote_1599652;
}

static size_t game4_reference_nox_vote_record_size_5066D0(void) {
	return sizeof(nox_vote_5066D0);
}

static nox_vote_5066D0* game4_reference_nox_vote_next_5066D0(nox_vote_5066D0* vote) {
	return vote == NULL ? NULL : vote->next;
}

static nox_vote_5066D0* game4_reference_nox_vote_previous_5066D0(nox_vote_5066D0* vote) {
	return vote == NULL ? NULL : vote->previous;
}

static nox_object_team_t* game4_reference_nox_vote_team_5066D0(nox_vote_5066D0* vote) {
	return vote == NULL ? NULL : vote->team;
}

static uint32_t game4_reference_nox_vote_frame_5066D0(nox_vote_5066D0* vote) {
	return vote == NULL ? 0 : vote->frame;
}

static uint32_t game4_reference_nox_vote_voters_5066D0(nox_vote_5066D0* vote) {
	return vote == NULL ? 0 : vote->voters;
}

static uint8_t game4_reference_nox_vote_count_5066D0(nox_vote_5066D0* vote) {
	return vote == NULL ? 0 : vote->count;
}

static uint8_t game4_reference_nox_vote_threshold_5066D0(nox_vote_5066D0* vote) {
	return vote == NULL ? 0 : vote->threshold;
}

static void game4_reference_nox_vote_set_voters_5066D0(nox_vote_5066D0* vote, uint32_t voters, uint8_t count) {
	if (vote != NULL) {
		vote->voters = voters;
		vote->count = count;
	}
}
#endif
