// Suppress unrelated Win32-only assertions while parsing the shared headers,
// then assert the native Pixie callback and every observed object field.
#define _Static_assert(...)
#include "../GAME3_3.h"
#undef _Static_assert

#include <stddef.h>
#include <stdint.h>

_Static_assert(sizeof(nox_projectile_collide_data_t) == 8,
	"Pixie collide-data size");
_Static_assert(offsetof(nox_projectile_collide_data_t, damage) == 0,
	"Pixie collide damage offset");
_Static_assert(offsetof(nox_object_t, obj_class) == (sizeof(void*) == 4 ? 8 : 12),
	"object class offset");
_Static_assert(offsetof(nox_object_t, obj_flags) == (sizeof(void*) == 4 ? 16 : 20),
	"object flags offset");
_Static_assert(offsetof(nox_object_t, x) == (sizeof(void*) == 4 ? 56 : 60),
	"object position offset");
_Static_assert(offsetof(nox_object_t, new_x) == (sizeof(void*) == 4 ? 64 : 68),
	"object new-position offset");
_Static_assert(offsetof(nox_object_t, vel_x) == (sizeof(void*) == 4 ? 80 : 84),
	"object velocity offset");
_Static_assert(offsetof(nox_object_t, direction1) == (sizeof(void*) == 4 ? 124 : 128),
	"object current-direction offset");
_Static_assert(offsetof(nox_object_t, direction2) == (sizeof(void*) == 4 ? 126 : 130),
	"object next-direction offset");
_Static_assert(offsetof(nox_object_t, owner) == (sizeof(void*) == 4 ? 508 : 552),
	"object owner offset");
_Static_assert(offsetof(nox_object_t, collide_data) == (sizeof(void*) == 4 ? 700 : 776),
	"object collide-data offset");
_Static_assert(offsetof(nox_object_t, func_damage) == (sizeof(void*) == 4 ? 716 : 808),
	"object Damage callback offset");
_Static_assert(
	__builtin_types_compatible_p(
		__typeof__(&nox_xxx_collidePixie_4EA080),
		void (*)(nox_object_t*, nox_object_t*, float*)),
	"PixieCollide callback pointer width");

static nox_object_t* seen_source;
static nox_object_t* seen_target;
static float* seen_collision;

void nox_xxx_collidePixie_4EA080(
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

	nox_xxx_collidePixie_4EA080(&source, &target, collision);
	if (seen_source != &source || seen_target != &target || seen_collision != collision) {
		return 1;
	}
	return 0;
}
