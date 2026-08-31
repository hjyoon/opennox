#include "../GAME4.h"

#include <limits.h>

typedef int32_t (*nox_active_ability_value_4fc3e0_fn)(nox_object_t*, int32_t);

typedef struct nox_active_ability_value_record_4fc3e0 {
	int32_t ability;
	nox_object_t* unit;
	uint32_t frame;
	uint32_t active;
	struct nox_active_ability_value_record_4fc3e0* next;
	struct nox_active_ability_value_record_4fc3e0* prev;
} nox_active_ability_value_record_4fc3e0;

#if defined(__clang__) || defined(__GNUC__)
_Static_assert(
	__builtin_types_compatible_p(
		__typeof__(&nox_xxx_probablyWarcryCheck_4FC3E0),
		nox_active_ability_value_4fc3e0_fn),
	"active-ability value lookup must retain its native object-pointer/int32 ABI");
#endif

_Static_assert(offsetof(nox_object_t, obj_class) == (sizeof(void*) == 4 ? 8 : 12),
	"active-ability value lookup must read the native Object class field");
_Static_assert(offsetof(nox_object_t, data_update) == (sizeof(void*) == 4 ? 748 : 872),
	"active-ability value lookup must read the native Object update-data pointer");
_Static_assert(offsetof(nox_player_update_data_t, player) ==
	(sizeof(void*) == 4 ? 276 : 336),
	"active-ability value lookup must read the native PlayerUpdate Player pointer");
_Static_assert(offsetof(nox_playerInfo, info) + offsetof(nox_playerInfo2, playerClass) ==
	(sizeof(void*) == 4 ? 2251 : 2255),
	"active-ability value lookup must read the exact Player class byte");
_Static_assert(offsetof(nox_active_ability_value_record_4fc3e0, ability) == 0,
	"active-ability value lookup must read the signed ability field first");
_Static_assert(offsetof(nox_active_ability_value_record_4fc3e0, unit) ==
	(sizeof(void*) == 4 ? 4 : 8),
	"active-ability value lookup must preserve the native record unit pointer");
_Static_assert(offsetof(nox_active_ability_value_record_4fc3e0, active) ==
	(sizeof(void*) == 4 ? 12 : 20),
	"active-ability value lookup must read the complete Active field");
_Static_assert(offsetof(nox_active_ability_value_record_4fc3e0, next) ==
	(sizeof(void*) == 4 ? 16 : 24),
	"active-ability value lookup must preserve the native record next pointer");
_Static_assert(sizeof(nox_active_ability_value_record_4fc3e0) ==
	(sizeof(void*) == 4 ? 24 : 40),
	"active-ability value record must retain its native-width layout");

int32_t nox_active_ability_value_4fc3e0_contract(nox_object_t* unit) {
	return nox_xxx_probablyWarcryCheck_4FC3E0(unit, INT32_MIN);
}
