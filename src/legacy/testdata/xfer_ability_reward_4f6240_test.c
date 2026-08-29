#include <assert.h>
#include <limits.h>
#include <stddef.h>
#include <stdint.h>

#include "../xfer_ability_reward_4f6240.h"

struct nox_object_t {
	uintptr_t marker;
};

typedef int32_t (*ability_reward_xfer_fn_4F6240)(nox_object_t*);

_Static_assert(CHAR_BIT == 8, "serialized bytes must remain eight bits");
_Static_assert(sizeof(void*) == 4 || sizeof(void*) == 8,
	"unsupported pointer width");
_Static_assert(sizeof(int32_t) == 4,
	"AbilityRewardXfer result must remain int32");
_Static_assert(
	_Generic(&nox_xxx_XFerAbilityReward_4F6240,
		ability_reward_xfer_fn_4F6240: 1, default: 0),
	"AbilityRewardXfer must preserve one native pointer and exact int32 result");

static nox_object_t* observed_object;

int32_t nox_xxx_XFerAbilityReward_4F6240(nox_object_t* object) {
	observed_object = object;
	return object != NULL;
}

int main(void) {
	nox_object_t object = {.marker = UINTPTR_MAX};
	ability_reward_xfer_fn_4F6240 const transfer =
		nox_xxx_XFerAbilityReward_4F6240;

	assert(transfer(&object) == 1);
	assert(observed_object == &object);
	assert(observed_object->marker == UINTPTR_MAX);

	assert(transfer(NULL) == 0);
	assert(observed_object == NULL);
	return 0;
}
