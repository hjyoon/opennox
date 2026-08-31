#include "../GAME4.h"

#include <limits.h>

typedef int32_t (*nox_active_ability_membership_4fc250_fn)(nox_object_t*, int32_t);

#if defined(__clang__) || defined(__GNUC__)
_Static_assert(
	__builtin_types_compatible_p(
		__typeof__(&nox_common_playerIsAbilityActive_4FC250),
		nox_active_ability_membership_4fc250_fn),
	"active-ability membership must retain its native object-pointer/int32 ABI");
#endif

_Static_assert(offsetof(nox_object_t, obj_class) == (sizeof(void*) == 4 ? 8 : 12),
	"active-ability membership must read the native Object class field");
_Static_assert(offsetof(nox_object_t, data_update) == (sizeof(void*) == 4 ? 748 : 872),
	"active-ability membership must read the native Object update-data pointer");
_Static_assert(offsetof(nox_player_update_data_t, player) ==
	(sizeof(void*) == 4 ? 276 : 336),
	"active-ability membership must read the native PlayerUpdate Player pointer");
_Static_assert(offsetof(nox_playerInfo, info) + offsetof(nox_playerInfo2, playerClass) ==
	(sizeof(void*) == 4 ? 2251 : 2255),
	"active-ability membership must read the exact Player class byte");

int32_t nox_active_ability_membership_4fc250_contract(nox_object_t* unit) {
	return nox_common_playerIsAbilityActive_4FC250(unit, INT32_MIN);
}
