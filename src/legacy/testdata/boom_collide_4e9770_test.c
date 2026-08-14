// Suppress unrelated Win32-only assertions while parsing the shared headers,
// then assert BoomCollide and direction-helper native pointer boundaries.
#define _Static_assert(...)
#include "../GAME3_3.h"
#include "../GAME4_1.h"
#undef _Static_assert

#include <stddef.h>
#include <stdint.h>

_Static_assert(offsetof(nox_object_t, obj_class) == (sizeof(void*) == 4 ? 8 : 12),
	"object class offset");
_Static_assert(offsetof(nox_object_t, x) == (sizeof(void*) == 4 ? 56 : 60),
	"object position offset");
_Static_assert(offsetof(nox_object_t, vel_x) == (sizeof(void*) == 4 ? 80 : 84),
	"object velocity offset");
_Static_assert(offsetof(nox_object_t, direction1) == (sizeof(void*) == 4 ? 124 : 128),
	"object primary direction offset");
_Static_assert(offsetof(nox_object_t, direction2) == (sizeof(void*) == 4 ? 126 : 130),
	"object secondary direction offset");
_Static_assert(offsetof(nox_object_t, owner) == (sizeof(void*) == 4 ? 508 : 552),
	"object owner offset");
_Static_assert(offsetof(nox_object_t, func_damage) == (sizeof(void*) == 4 ? 716 : 808),
	"object damage callback offset");
_Static_assert(
	__builtin_types_compatible_p(
		__typeof__(&nox_xxx_collideBoom_4E9770),
		void (*)(nox_object_t*, nox_object_t*, float*)),
	"BoomCollide callback pointer width");
_Static_assert(
	__builtin_types_compatible_p(
		__typeof__(&nox_xxx_math_509ED0),
		int (*)(float2*)),
	"direction callback pointer width");

static nox_object_t* seen_source;
static nox_object_t* seen_target;
static float* seen_collision;
static float2* seen_vector;

void nox_xxx_collideBoom_4E9770(
	nox_object_t* source,
	nox_object_t* target,
	float* collision) {
	seen_source = source;
	seen_target = target;
	seen_collision = collision;
}

int nox_xxx_math_509ED0(float2* vector) {
	seen_vector = vector;
	return 0x7f;
}

int main(void) {
	nox_object_t source = {0};
	nox_object_t target = {0};
	float collision[2] = {3.5f, -8.25f};
	float2 vector = {.field_0 = 1.25f, .field_4 = -2.5f};

	nox_xxx_collideBoom_4E9770(&source, &target, collision);
	if (seen_source != &source || seen_target != &target || seen_collision != collision) {
		return 1;
	}
	if (nox_xxx_math_509ED0(&vector) != 0x7f || seen_vector != &vector) {
		return 2;
	}
	if (vector.field_0 != 1.25f || vector.field_4 != -2.5f) {
		return 3;
	}
	return 0;
}
