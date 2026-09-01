#include <assert.h>
#include <limits.h>
#include <stdint.h>

#include "../GAME4_3.h"

typedef int32_t (*spell_reward_use_fn)(nox_object_t*, nox_object_t*);

_Static_assert(CHAR_BIT == 8, "bytes must remain eight bits");
_Static_assert(sizeof(int32_t) == 4, "result must remain exact signed int32");
_Static_assert(sizeof(void*) == 4 || sizeof(void*) == 8, "unsupported pointer width");
_Static_assert(
	_Generic(&nox_xxx_useSpellReward_53F9E0,
		spell_reward_use_fn: 1, default: 0),
	"SpellRewardUse must preserve both native object pointers");

static nox_object_t* observed_owner;
static nox_object_t* observed_item;
static unsigned int observed_calls;

int32_t nox_xxx_useSpellReward_53F9E0(
		nox_object_t* owner, nox_object_t* item) {
	observed_owner = owner;
	observed_item = item;
	++observed_calls;
	return INT32_MIN;
}

int main(void) {
	nox_object_t owner = {0};
	nox_object_t item = {0};
	spell_reward_use_fn const use = nox_xxx_useSpellReward_53F9E0;

	assert(use(&owner, &item) == INT32_MIN);
	assert(observed_owner == &owner);
	assert(observed_item == &item);
	assert(use(NULL, NULL) == INT32_MIN);
	assert(observed_owner == NULL);
	assert(observed_item == NULL);
	assert(observed_calls == 2);
	return 0;
}
