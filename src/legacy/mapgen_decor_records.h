#ifndef NOX_MAPGEN_DECOR_RECORDS_H
#define NOX_MAPGEN_DECOR_RECORDS_H

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

#endif // NOX_MAPGEN_DECOR_RECORDS_H
