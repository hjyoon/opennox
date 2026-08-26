// Keep this fixture independent from Win32-only aggregate legacy declarations
// so every supported target frontend can compile the retained public ABI.
#include "../reward_field_guide_4f0d20.h"

#include <limits.h>
#include <stdint.h>

struct nox_object_t {
	uintptr_t observed_self;
	uint32_t observed_stage;
	uint32_t guard;
};

typedef nox_object_t* (*reward_field_guide_fn)(nox_object_t*, uint32_t);

_Static_assert(CHAR_BIT == 8, "bytes must remain eight bits");
_Static_assert(sizeof(void*) == 4 || sizeof(void*) == 8, "unsupported pointer width");
_Static_assert(
	_Generic(&nox_xxx_rewardFieldGuide_4F0D20,
		reward_field_guide_fn: 1, default: 0),
	"reward field-guide creation must use native object pointers and uint32_t stage");

static nox_object_t* observed_marker;

nox_object_t* nox_xxx_rewardFieldGuide_4F0D20(
		nox_object_t* marker, uint32_t stage) {
	observed_marker = marker;
	marker->observed_self = (uintptr_t)marker;
	marker->observed_stage = stage;
	return marker;
}

int main(void) {
	nox_object_t marker = {
		.observed_self = (uintptr_t)0,
		.observed_stage = UINT32_C(0),
		.guard = UINT32_C(0x5A5A5A5A),
	};
	reward_field_guide_fn const create = nox_xxx_rewardFieldGuide_4F0D20;

	if (create(&marker, UINT32_MAX) != &marker || observed_marker != &marker ||
		marker.observed_self != (uintptr_t)&marker ||
		marker.observed_stage != UINT32_MAX ||
		marker.guard != UINT32_C(0x5A5A5A5A)) {
		return __LINE__;
	}
	return 0;
}
