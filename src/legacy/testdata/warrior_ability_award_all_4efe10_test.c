#include "../warrior_ability_award_all_4efe10.h"

typedef void (*nox_warrior_ability_award_all_4efe10_fn)(nox_playerInfo*);

#if defined(__clang__) || defined(__GNUC__)
_Static_assert(
	__builtin_types_compatible_p(
		__typeof__(&nox_xxx_spellAwardAll3_4EFE10),
		nox_warrior_ability_award_all_4efe10_fn),
	"warrior-ability award-all must retain its native player-pointer void ABI");
#endif

void nox_warrior_ability_award_all_4efe10_contract(nox_playerInfo* player) {
	nox_xxx_spellAwardAll3_4EFE10(player);
}
