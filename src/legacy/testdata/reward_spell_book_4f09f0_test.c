// Keep this fixture independent from Win32-only aggregate legacy declarations
// so every supported target frontend can compile the retained public ABIs.
#include "../reward_spell_book_4f09f0.h"

#include <limits.h>
#include <stdint.h>

struct nox_object_t {
	uint32_t observed_stage;
	uint32_t guard;
};

typedef nox_object_t* (*reward_spell_book_fn)(nox_object_t*, uint32_t);
typedef uint32_t (*reward_random_slots_fn)(uint32_t);

_Static_assert(CHAR_BIT == 8, "bytes must remain eight bits");
_Static_assert(sizeof(uint32_t) == 4, "reward scalars must remain exact dwords");
_Static_assert(sizeof(void*) == 4 || sizeof(void*) == 8, "unsupported pointer width");
_Static_assert(
	_Generic(&nox_xxx_rewardSpellBook_4F09F0,
		reward_spell_book_fn: 1, default: 0),
	"reward spell-book creation must use and return native object pointers");
_Static_assert(
	_Generic(&nox_server_rewardGen_pickRandomSlots_4F0B60,
		reward_random_slots_fn: 1, default: 0),
	"reward slot selection must use exact uint32_t input and output");

static nox_object_t* observed_marker;
static uint32_t observed_slot_stage;

nox_object_t* nox_xxx_rewardSpellBook_4F09F0(
		nox_object_t* marker, uint32_t stage) {
	observed_marker = marker;
	marker->observed_stage = stage;
	return marker;
}

uint32_t nox_server_rewardGen_pickRandomSlots_4F0B60(uint32_t stage) {
	observed_slot_stage = stage;
	return stage ^ UINT32_C(0xA5A5A5A5);
}

int main(void) {
	nox_object_t marker = {
		.observed_stage = UINT32_C(0),
		.guard = UINT32_C(0x5A5A5A5A),
	};
	reward_spell_book_fn const create = nox_xxx_rewardSpellBook_4F09F0;
	reward_random_slots_fn const slots = nox_server_rewardGen_pickRandomSlots_4F0B60;

	if (create(&marker, UINT32_C(0xFEDCBA98)) != &marker ||
		observed_marker != &marker || marker.observed_stage != UINT32_C(0xFEDCBA98) ||
		marker.guard != UINT32_C(0x5A5A5A5A)) {
		return __LINE__;
	}
	if (slots(UINT32_C(0x89ABCDEF)) != UINT32_C(0x2C0E684A) ||
		observed_slot_stage != UINT32_C(0x89ABCDEF)) {
		return __LINE__;
	}
	return 0;
}
