#include <stdint.h>

#include "../GAME4.h"

typedef void (*nox_active_ability_deactivate_4fc440_fn)(nox_object_t*, int32_t);

typedef struct nox_active_ability_deactivate_record_4fc440 {
	int32_t ability;
	nox_object_t* unit;
	uint32_t frame;
	uint32_t active;
	struct nox_active_ability_deactivate_record_4fc440* next;
	struct nox_active_ability_deactivate_record_4fc440* prev;
} nox_active_ability_deactivate_record_4fc440;

#if defined(__clang__) || defined(__GNUC__)
_Static_assert(
	__builtin_types_compatible_p(
		__typeof__(&sub_4FC440),
		nox_active_ability_deactivate_4fc440_fn),
	"active-ability deactivation must retain its native object-pointer/int32 ABI");
#endif

_Static_assert(offsetof(nox_object_t, obj_class) == (sizeof(void*) == 4 ? 8 : 12),
	"active-ability deactivation must read the native Object class field");
_Static_assert(offsetof(nox_object_t, data_update) == (sizeof(void*) == 4 ? 748 : 872),
	"active-ability deactivation must read the native Object update-data pointer");
_Static_assert(offsetof(nox_player_update_data_t, player) ==
	(sizeof(void*) == 4 ? 276 : 336),
	"active-ability deactivation must read the native PlayerUpdate Player pointer");
_Static_assert(offsetof(nox_playerInfo, info) + offsetof(nox_playerInfo2, playerClass) ==
	(sizeof(void*) == 4 ? 2251 : 2255),
	"active-ability deactivation must read the exact Player class byte");
_Static_assert(offsetof(nox_active_ability_deactivate_record_4fc440, ability) == 0,
	"active-ability deactivation must read the signed ability field first");
_Static_assert(offsetof(nox_active_ability_deactivate_record_4fc440, unit) ==
	(sizeof(void*) == 4 ? 4 : 8),
	"active-ability deactivation must preserve the native record unit pointer");
_Static_assert(offsetof(nox_active_ability_deactivate_record_4fc440, active) ==
	(sizeof(void*) == 4 ? 12 : 20),
	"active-ability deactivation must write the complete Active field");
_Static_assert(offsetof(nox_active_ability_deactivate_record_4fc440, next) ==
	(sizeof(void*) == 4 ? 16 : 24),
	"active-ability deactivation must preserve the native record next pointer");
_Static_assert(sizeof(nox_active_ability_deactivate_record_4fc440) ==
	(sizeof(void*) == 4 ? 24 : 40),
	"active-ability deactivation record must retain its native-width layout");

void nox_active_ability_deactivate_4fc440_contract(nox_object_t* unit) {
	sub_4FC440(unit, INT32_MIN);
}
