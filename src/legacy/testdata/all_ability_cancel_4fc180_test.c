#include "../GAME4.h"

typedef void (*nox_all_ability_cancel_4fc180_fn)(nox_object_t*);

#if defined(__clang__) || defined(__GNUC__)
_Static_assert(
	__builtin_types_compatible_p(
		__typeof__(&nox_xxx_playerCancelAbils_4FC180),
		nox_all_ability_cancel_4fc180_fn),
	"all-ability cancellation must retain its native object-pointer void ABI");
#endif

_Static_assert(offsetof(nox_playerInfo, playerUnit) == 2056,
	"all-ability caller must use the native player-unit field");
_Static_assert(sizeof(((nox_playerInfo*)0)->playerUnit) == sizeof(void*),
	"all-ability caller must not narrow the player-unit pointer");

void nox_all_ability_cancel_4fc180_contract(nox_playerInfo* player) {
	if (player->playerUnit) {
		nox_xxx_playerCancelAbils_4FC180(player->playerUnit);
	}
}
