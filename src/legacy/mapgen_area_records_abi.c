#include <stddef.h>

#include "GAME4_2.h"
#include "mapgen_decor_records.h"

_Static_assert(sizeof(nox_mapgen_occupied_rect_legacy) == 28,
			   "wrong mapgen occupied-rectangle record size");
_Static_assert(offsetof(nox_mapgen_occupied_rect_legacy, next_token) == 24,
			   "wrong mapgen occupied-rectangle next-token offset");
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

// GAME.EXE returns the processed record count in a pointer-sized register.
extern uintptr_t nox_mapgenLoadAreaRecordsNative(void);
char* sub_526950(void) {
	return (char*)nox_mapgenLoadAreaRecordsNative();
}
