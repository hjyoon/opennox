#include <assert.h>
#include <limits.h>
#include <stdint.h>

#include "../GAME4_3.h"
#include "../server__ability__ability.h"

struct nox_object_t {
	uintptr_t marker;
};

typedef int32_t (*ability_reward_fn)(nox_object_t*, int32_t, int32_t);
typedef int32_t (*ability_reward_use_fn)(nox_object_t*, nox_object_t*);

_Static_assert(CHAR_BIT == 8, "bytes must remain eight bits");
_Static_assert(sizeof(int32_t) == 4, "ability scalars must remain exact signed int32");
_Static_assert(sizeof(void*) == 4 || sizeof(void*) == 8, "unsupported pointer width");
_Static_assert(
	_Generic(&nox_xxx_abilityRewardServ_4FB9C0_ability,
		ability_reward_fn: 1, default: 0),
	"ability reward must use a native object pointer and signed int32 scalars");
_Static_assert(
	_Generic(&nox_xxx_useAbilityReward_53FAE0,
		ability_reward_use_fn: 1, default: 0),
	"ability reward use must preserve both native object pointers");

static nox_object_t* observed_unit;
static nox_object_t* observed_item;
static int32_t observed_ability;
static int32_t observed_reward_arg;
static unsigned int observed_reward_calls;
static unsigned int observed_use_calls;

int32_t nox_xxx_abilityRewardServ_4FB9C0_ability(
		nox_object_t* unit, int32_t ability, int32_t reward_arg) {
	observed_unit = unit;
	observed_ability = ability;
	observed_reward_arg = reward_arg;
	++observed_reward_calls;
	return reward_arg;
}

int32_t nox_xxx_useAbilityReward_53FAE0(
		nox_object_t* owner, nox_object_t* item) {
	observed_unit = owner;
	observed_item = item;
	++observed_use_calls;
	return INT32_MIN;
}

int main(void) {
	nox_object_t owner = {.marker = UINTPTR_MAX};
	nox_object_t item = {.marker = UINTPTR_MAX - UINTPTR_C(1)};
	ability_reward_fn const reward = nox_xxx_abilityRewardServ_4FB9C0_ability;
	ability_reward_use_fn const use = nox_xxx_useAbilityReward_53FAE0;

	assert(reward(&owner, INT32_MAX, INT32_MIN) == INT32_MIN);
	assert(observed_unit == &owner);
	assert(observed_ability == INT32_MAX);
	assert(observed_reward_arg == INT32_MIN);
	assert(reward(NULL, INT32_MIN, INT32_MAX) == INT32_MAX);
	assert(observed_unit == NULL);
	assert(observed_reward_calls == 2);

	assert(use(&owner, &item) == INT32_MIN);
	assert(observed_unit == &owner);
	assert(observed_item == &item);
	assert(use(NULL, NULL) == INT32_MIN);
	assert(observed_unit == NULL);
	assert(observed_item == NULL);
	assert(observed_use_calls == 2);
	return 0;
}
