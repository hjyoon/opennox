#ifndef OPENNOX_WORLD_NATIVE_STORAGE_H
#define OPENNOX_WORLD_NATIVE_STORAGE_H

#include "defs.h"

enum { NOX_WORLD_TILE_CAPACITY = 128 };

typedef struct {
	char * archive_records;
	uint32_t archive_record_count;
	uint32_t objects_placed;
	uint32_t selected_archive_index;
	uint32_t merge_walls;
	uint32_t repeat_walls;
	uint32_t item_set_template;
	uint32_t hall_tracking_head;
	uint32_t room_head;
	uint32_t farthest_room;
	uint32_t room_rank_word_2487584;
	uint32_t object_record_head;
	uint32_t object_record_cursor;
	uint32_t north_prefab;
} nox_archive_mapgen_state_t;
_Static_assert(offsetof(nox_archive_mapgen_state_t, archive_record_count) == sizeof(void*), "archive count offset");
nox_archive_mapgen_state_t* nox_archive_mapgen_storage(void);

typedef struct {
	obj_5D4594_2650668_t ** tile_coordinates;
	obj_5D4594_754088_t * focus_primary;
	obj_5D4594_754088_t * focus_secondary;
	void * quickbar_data;
	void * secret_wall_head;
	uintptr_t local_player_word;
	int focus_primary_count;
	int focus_secondary_count;
} nox_world_owner_state_t;
_Static_assert(offsetof(nox_world_owner_state_t, focus_primary_count) == 6 * sizeof(void*), "focus word offset");
_Static_assert(sizeof(nox_world_owner_state_t) == 6 * sizeof(void*) + 2 * sizeof(uint32_t), "world owner state size");
nox_world_owner_state_t* nox_world_owner_storage(void);

typedef struct {
	uint32_t global_record_word_60044;
	uint32_t buff_iteration_count;
	uint32_t settings_scalar;
	uint32_t death_message_word_1203836;
	uint32_t death_message_word_1203840;
	uint32_t nearby_server_count;
	uint32_t map_effect_word_1313740;
	uint32_t map_effect_word_1313788;
	uint32_t chakram_selected_word;
	unsigned int script_pointer_word;
	uint32_t object_trace_word;
	uint32_t server_subflags;
	uint32_t team_entries_limit;
	uint32_t local_player_color;
	uint32_t next_polygon_index;
} nox_world_word_state_t;
_Static_assert(sizeof(nox_world_word_state_t) == 15 * sizeof(uint32_t), "world word widths");
nox_world_word_state_t* nox_world_word_storage(void);

#endif
