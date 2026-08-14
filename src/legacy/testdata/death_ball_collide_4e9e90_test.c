// Suppress unrelated Win32-only assertions while parsing the shared headers,
// then assert the native DeathBall callback and the fields it observes.
#define _Static_assert(...)
#include "../GAME3_3.h"
#include "../GAME4_3.h"
#undef _Static_assert

#include <stddef.h>
#include <stdint.h>

_Static_assert(offsetof(nox_object_t, obj_class) == (sizeof(void*) == 4 ? 8 : 12),
	"object class offset");
_Static_assert(offsetof(nox_object_t, x) == (sizeof(void*) == 4 ? 56 : 60),
	"object position offset");
_Static_assert(offsetof(nox_object_t, new_x) == (sizeof(void*) == 4 ? 64 : 68),
	"object new-position offset");
_Static_assert(offsetof(nox_object_t, prev_x) == (sizeof(void*) == 4 ? 72 : 76),
	"object previous-position offset");
_Static_assert(offsetof(nox_object_t, vel_x) == (sizeof(void*) == 4 ? 80 : 84),
	"object velocity offset");
_Static_assert(offsetof(nox_object_t, func_damage) == (sizeof(void*) == 4 ? 716 : 808),
	"object Damage callback offset");
_Static_assert(offsetof(nox_object_t, data_update) == (sizeof(void*) == 4 ? 748 : 872),
	"object update-data offset");
_Static_assert(sizeof(nox_door_update_data_t) == 52,
	"Door update-data size");
_Static_assert(offsetof(nox_door_update_data_t, current_direction) == 12,
	"Door current-direction offset");
_Static_assert(
	__builtin_types_compatible_p(
		__typeof__(&nox_xxx_collideDeathBall_4E9E90),
		void (*)(nox_object_t*, nox_object_t*, float*)),
	"DeathBallCollide callback pointer width");

static nox_object_t* seen_source;
static nox_object_t* seen_target;
static float* seen_collision;

void nox_xxx_collideDeathBall_4E9E90(
	nox_object_t* source,
	nox_object_t* target,
	float* collision) {
	seen_source = source;
	seen_target = target;
	seen_collision = collision;
}

int main(void) {
	nox_object_t source = {0};
	nox_object_t target = {0};
	float collision[2] = {3.5f, -8.25f};

	nox_xxx_collideDeathBall_4E9E90(&source, &target, collision);
	if (seen_source != &source || seen_target != &target || seen_collision != collision) {
		return 1;
	}
	return 0;
}
