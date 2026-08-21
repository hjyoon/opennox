#include "../beast_scroll_award_all_4efd80.h"

typedef void (*nox_beast_scroll_award_all_4efd80_fn)(nox_playerInfo*);

#if defined(__clang__) || defined(__GNUC__)
_Static_assert(
	__builtin_types_compatible_p(
		__typeof__(&nox_xxx_spellAwardAll1_4EFD80),
		nox_beast_scroll_award_all_4efd80_fn),
	"beast-scroll award-all must retain its native player-pointer void ABI");
#endif

void nox_beast_scroll_award_all_4efd80_contract(nox_playerInfo* player) {
	nox_xxx_spellAwardAll1_4EFD80(player);
}
