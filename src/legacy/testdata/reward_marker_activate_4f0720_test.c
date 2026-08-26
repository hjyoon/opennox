// Keep this fixture independent from the Win32-only aggregate legacy headers
// so every supported target frontend can compile the retained public ABI.
#include "../reward_marker_activate_4f0720.h"

#include <limits.h>
#include <stdint.h>

struct nox_object_t {
	uint32_t observed_stage;
	uint32_t guard;
};

typedef nox_object_t* (*reward_marker_activate_fn)(nox_object_t*, uint32_t);

_Static_assert(CHAR_BIT == 8, "bytes must remain eight bits");
_Static_assert(sizeof(uint32_t) == 4, "reward stage must remain an exact dword");
_Static_assert(sizeof(void*) == 4 || sizeof(void*) == 8, "unsupported pointer width");
_Static_assert(
	_Generic(&nox_server_rewardgen_activateMarker_4F0720,
		reward_marker_activate_fn: 1, default: 0),
	"RewardMarkerActivate must use and return native object pointers");

static nox_object_t* observed_marker;

nox_object_t* nox_server_rewardgen_activateMarker_4F0720(
		nox_object_t* marker, uint32_t stage) {
	observed_marker = marker;
	marker->observed_stage = stage;
	return marker;
}

int main(void) {
	nox_object_t marker = {
		.observed_stage = UINT32_C(0),
		.guard = UINT32_C(0xA5A5A5A5),
	};
	reward_marker_activate_fn const activate =
		nox_server_rewardgen_activateMarker_4F0720;
	nox_object_t* const result = activate(&marker, UINT32_C(0xFEDCBA98));

	if (observed_marker != &marker || result != &marker)
		return __LINE__;
	if (marker.observed_stage != UINT32_C(0xFEDCBA98) ||
		marker.guard != UINT32_C(0xA5A5A5A5))
		return __LINE__;
	return 0;
}
