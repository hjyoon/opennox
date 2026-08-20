#include <assert.h>
#include <stdint.h>

#include "../server__object__health.h"

typedef void (*solo_monster_kill_reward_callback_t)(nox_object_t*);
typedef double (*unit_give_xp_callback_t)(nox_object_t*, float);

_Static_assert(sizeof(void*) == 4 || sizeof(void*) == 8, "unsupported pointer width");
_Static_assert(
	_Generic(
		&nox_xxx_soloMonsterKillReward_4EE500_obj_health,
		solo_monster_kill_reward_callback_t: 1,
		default: 0),
	"SoloMonsterKillReward must use one native object pointer");
_Static_assert(
	_Generic(
		&nox_xxx_unitGiveXP_4EF270,
		unit_give_xp_callback_t: 1,
		default: 0),
	"UnitGiveXP must use one native object pointer and one binary32 target");

static nox_object_t* observed_object;

void nox_xxx_soloMonsterKillReward_4EE500_obj_health(nox_object_t* object) {
	observed_object = object;
}

int main(void) {
	uintptr_t storage = UINTPTR_MAX;
	nox_object_t* const object = (nox_object_t*)&storage;
	nox_xxx_soloMonsterKillReward_4EE500_obj_health(object);
	assert(observed_object == object);
	return 0;
}
