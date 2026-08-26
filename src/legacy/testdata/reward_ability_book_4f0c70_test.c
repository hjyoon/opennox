// Keep this fixture independent from Win32-only aggregate legacy declarations
// so every supported target frontend can compile the retained public ABI.
#include "../reward_ability_book_4f0c70.h"

#include <limits.h>
#include <stdint.h>

struct nox_object_t {
	uintptr_t observed_self;
	uint32_t guard;
};

typedef nox_object_t* (*reward_ability_book_fn)(nox_object_t*);

_Static_assert(CHAR_BIT == 8, "bytes must remain eight bits");
_Static_assert(sizeof(void*) == 4 || sizeof(void*) == 8, "unsupported pointer width");
_Static_assert(
	_Generic(&nox_xxx_rewardAbilityBook_4F0C70,
		reward_ability_book_fn: 1, default: 0),
	"reward ability-book creation must use and return native object pointers");

static nox_object_t* observed_marker;

nox_object_t* nox_xxx_rewardAbilityBook_4F0C70(nox_object_t* marker) {
	observed_marker = marker;
	marker->observed_self = (uintptr_t)marker;
	return marker;
}

int main(void) {
	nox_object_t marker = {
		.observed_self = (uintptr_t)0,
		.guard = UINT32_C(0x5A5A5A5A),
	};
	reward_ability_book_fn const create = nox_xxx_rewardAbilityBook_4F0C70;

	if (create(&marker) != &marker || observed_marker != &marker ||
		marker.observed_self != (uintptr_t)&marker ||
		marker.guard != UINT32_C(0x5A5A5A5A)) {
		return __LINE__;
	}
	return 0;
}
