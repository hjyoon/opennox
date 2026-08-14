// Suppress unrelated Win32-only assertions while parsing the shared headers,
// then assert the native WallReflectSpark callback and observed object fields.
#define _Static_assert(...)
#include "../GAME3_3.h"
#undef _Static_assert

#include <stddef.h>
#include <stdint.h>

_Static_assert(sizeof(nox_projectile_collide_data_t) == 8,
	"WallReflectSpark collide-data size");
_Static_assert(offsetof(nox_projectile_collide_data_t, damage) == 0,
	"WallReflectSpark collide damage offset");
_Static_assert(offsetof(nox_object_t, new_x) == (sizeof(void*) == 4 ? 64 : 68),
	"object new-position offset");
_Static_assert(offsetof(nox_object_t, vel_x) == (sizeof(void*) == 4 ? 80 : 84),
	"object velocity offset");
_Static_assert(offsetof(nox_object_t, collide_data) == (sizeof(void*) == 4 ? 700 : 776),
	"object collide-data offset");
_Static_assert(offsetof(nox_object_t, func_damage) == (sizeof(void*) == 4 ? 716 : 808),
	"object Damage callback offset");
_Static_assert(
	__builtin_types_compatible_p(
		__typeof__(&nox_xxx_collideWallReflectSpark_4EA200),
		void (*)(nox_object_t*, nox_object_t*, float*)),
	"WallReflectSparkCollide callback pointer width");

static nox_object_t* seen_source;
static nox_object_t* seen_target;
static float* seen_collision;

void nox_xxx_collideWallReflectSpark_4EA200(
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

	nox_xxx_collideWallReflectSpark_4EA200(&source, &target, collision);
	if (seen_source != &source || seen_target != &target || seen_collision != collision) {
		return 1;
	}
	return 0;
}
