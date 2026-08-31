#include <stdint.h>

#include "../GAME4.h"

typedef void (*nox_active_ability_disable_4fc300_fn)(nox_object_t*, int32_t);

#if defined(__clang__) || defined(__GNUC__)
_Static_assert(
	__builtin_types_compatible_p(
		__typeof__(&sub_4FC300),
		nox_active_ability_disable_4fc300_fn),
	"active-ability disable must retain its object-pointer and signed-int32 ABI");
#endif

_Static_assert(offsetof(nox_object_t, data_update) ==
	(sizeof(void*) == 4 ? 748 : 872),
	"active-ability disable must use the native update-data field");
_Static_assert(offsetof(nox_player_update_data_t, harpoon_bolt) ==
	(sizeof(void*) == 4 ? 136 : 160),
	"active-ability disable must use the native Harpoon-bolt field");

void nox_active_ability_disable_4fc300_contract(nox_object_t* unit) {
	sub_4FC300(unit, INT32_MIN);
}
