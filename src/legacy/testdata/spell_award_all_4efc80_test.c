#include "../spell_award_all_4efc80.h"

typedef void (*nox_spell_award_all_4efc80_fn)(nox_playerInfo*);

#if defined(__clang__) || defined(__GNUC__)
_Static_assert(
	__builtin_types_compatible_p(
		__typeof__(&nox_xxx_spellAwardAll2_4EFC80),
		nox_spell_award_all_4efc80_fn),
	"spell award-all must retain its native player-pointer void ABI");
#endif

void nox_spell_award_all_4efc80_contract(nox_playerInfo* player) {
	nox_xxx_spellAwardAll2_4EFC80(player);
}
