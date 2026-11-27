#ifndef NOX_MAPGEN_DECOR_RECORDS_H
#define NOX_MAPGEN_DECOR_RECORDS_H

#include <stddef.h>
#include <stdint.h>

// These map generator records keep their PE32 layouts. Link fields are
// registry tokens for native allocations.
typedef struct nox_mapgen_decor_entry_legacy {
	uint32_t type;
	char name[60];
	uint32_t density_enabled;
	int32_t min_count;
	int32_t max_count;
	float density;
	uint32_t choices_token;
	uint32_t foreach_token;
	uint32_t next_token;
	uint32_t shuffle_next_token;
	uint32_t shuffle_prev_token;
} nox_mapgen_decor_entry_legacy;

typedef struct nox_mapgen_decor_set_legacy {
	int32_t min_count;
	int32_t max_count;
	uint32_t entries_token;
	uint32_t entry_count;
	uint32_t shares_entries;
	uint32_t next_token;
} nox_mapgen_decor_set_legacy;

typedef struct nox_mapgen_decor_definition_legacy {
	char name[60];
	uint32_t room_type;
	uint8_t constraint_mask;
	uint8_t occur_limit;
	uint8_t exhausted;
	uint8_t must_occur;
	uint8_t occurred;
	uint8_t reserved[3];
	uint32_t frequency;
	int32_t min_room_size;
	int32_t max_room_size;
	uint32_t wall_floors_token;
	uint32_t wall_floor_count;
	uint32_t decor_sets_token;
	uint32_t decor_set_count;
	char door_name[60];
	char double_door_name[60];
	uint32_t next_token;
} nox_mapgen_decor_definition_legacy;

typedef struct nox_mapgen_decor_fill_legacy {
	uint32_t type;
	char object_name[60];
	char tile_name[60];
	uint32_t next_token;
} nox_mapgen_decor_fill_legacy;

typedef struct nox_mapgen_wall_floor_legacy {
	char wall_name[60];
	char floor_name[60];
	uint32_t fills_token;
	uint32_t next_token;
} nox_mapgen_wall_floor_legacy;

typedef struct nox_mapgen_item_set_entry_legacy {
	char lookup_name[60];
	char object_name[60];
	uint32_t modifier_tokens[4];
	uint32_t modifier_counts[4];
	uint32_t next_token;
} nox_mapgen_item_set_entry_legacy;

_Static_assert(sizeof(nox_mapgen_decor_entry_legacy) == 100,
			   "wrong mapgen decor-entry record size");
_Static_assert(offsetof(nox_mapgen_decor_entry_legacy, choices_token) == 80,
			   "wrong mapgen decor-entry choices-token offset");
_Static_assert(offsetof(nox_mapgen_decor_entry_legacy, foreach_token) == 84,
			   "wrong mapgen decor-entry FOREACH-token offset");
_Static_assert(offsetof(nox_mapgen_decor_entry_legacy, next_token) == 88,
			   "wrong mapgen decor-entry next-token offset");
_Static_assert(offsetof(nox_mapgen_decor_entry_legacy, shuffle_next_token) == 92,
			   "wrong mapgen decor-entry shuffle-next offset");
_Static_assert(offsetof(nox_mapgen_decor_entry_legacy, shuffle_prev_token) == 96,
			   "wrong mapgen decor-entry shuffle-prev offset");
_Static_assert(sizeof(nox_mapgen_decor_set_legacy) == 24,
			   "wrong mapgen decor-set record size");
_Static_assert(offsetof(nox_mapgen_decor_set_legacy, next_token) == 20,
			   "wrong mapgen decor-set next-token offset");
_Static_assert(sizeof(nox_mapgen_decor_definition_legacy) == 224,
			   "wrong mapgen decor-definition record size");
_Static_assert(offsetof(nox_mapgen_decor_definition_legacy, wall_floors_token) == 84,
			   "wrong mapgen decor-definition wall/floor-token offset");
_Static_assert(offsetof(nox_mapgen_decor_definition_legacy, decor_sets_token) == 92,
			   "wrong mapgen decor-definition decor-set-token offset");
_Static_assert(offsetof(nox_mapgen_decor_definition_legacy, next_token) == 220,
			   "wrong mapgen decor-definition next-token offset");
_Static_assert(sizeof(nox_mapgen_item_set_entry_legacy) == 156, "wrong mapgen item-set record size");
_Static_assert(offsetof(nox_mapgen_item_set_entry_legacy, modifier_tokens) == 120,
			   "wrong mapgen item-set modifier-token offset");
_Static_assert(offsetof(nox_mapgen_item_set_entry_legacy, modifier_counts) == 136,
			   "wrong mapgen item-set modifier-count offset");
_Static_assert(offsetof(nox_mapgen_item_set_entry_legacy, next_token) == 152,
			   "wrong mapgen item-set next-token offset");
_Static_assert(sizeof(nox_mapgen_decor_fill_legacy) == 128,
			   "wrong mapgen decor-fill record size");
_Static_assert(offsetof(nox_mapgen_decor_fill_legacy, next_token) == 124,
			   "wrong mapgen decor-fill next-token offset");
_Static_assert(sizeof(nox_mapgen_wall_floor_legacy) == 128,
			   "wrong mapgen wall/floor record size");
_Static_assert(offsetof(nox_mapgen_wall_floor_legacy, fills_token) == 120,
			   "wrong mapgen wall/floor fill-token offset");
_Static_assert(offsetof(nox_mapgen_wall_floor_legacy, next_token) == 124,
			   "wrong mapgen wall/floor next-token offset");

#endif // NOX_MAPGEN_DECOR_RECORDS_H
