// Suppress unrelated Win32-only assertions while parsing the shared headers,
// then assert the native callback, object, and shared parser ABI.
#define _Static_assert(...)
#include "../GAME3_3.h"
#include "../GAME4_3.h"
#undef _Static_assert

#include <limits.h>
#include <stddef.h>
#include <stdint.h>

_Static_assert(sizeof(nox_projectile_collide_data_t) == 8,
	"WallReflectCollide data size");
_Static_assert(offsetof(nox_projectile_collide_data_t, damage) == 0,
	"WallReflectCollide damage offset");
_Static_assert(offsetof(nox_projectile_collide_data_t, field_4) == 4,
	"WallReflectCollide field_4 offset");
_Static_assert(offsetof(nox_object_t, x) == (sizeof(void*) == 4 ? 56 : 60),
	"object position offset");
_Static_assert(offsetof(nox_object_t, new_x) == (sizeof(void*) == 4 ? 64 : 68),
	"object new-position offset");
_Static_assert(offsetof(nox_object_t, vel_x) == (sizeof(void*) == 4 ? 80 : 84),
	"object velocity offset");
_Static_assert(offsetof(nox_object_t, func_collide) == (sizeof(void*) == 4 ? 696 : 768),
	"object Collide callback offset");
_Static_assert(offsetof(nox_object_t, collide_data) == (sizeof(void*) == 4 ? 700 : 776),
	"object collide-data offset");
_Static_assert(offsetof(nox_object_t, func_damage) == (sizeof(void*) == 4 ? 716 : 808),
	"object Damage callback offset");
_Static_assert(
	__builtin_types_compatible_p(
		__typeof__(&nox_xxx_collideSulphurShot2_4E9D80),
		void (*)(nox_object_t*, nox_object_t*, float*)),
	"WallReflectCollide callback pointer width");
_Static_assert(
	__builtin_types_compatible_p(
		__typeof__(&nox_xxx_collideSulphurShot_4E9E50),
		void (*)(nox_object_t*, nox_object_t*, float*)),
	"YellowStarShotCollide callback pointer width");
_Static_assert(
	__builtin_types_compatible_p(__typeof__(&sub_536D80), int (*)(char*, void*)),
	"shared projectile parser pointer width");

static nox_object_t* seen_source;
static nox_object_t* seen_target;
static float* seen_collision;
static int seen_callback;

void nox_xxx_collideSulphurShot2_4E9D80(
	nox_object_t* source,
	nox_object_t* target,
	float* collision) {
	seen_source = source;
	seen_target = target;
	seen_collision = collision;
	seen_callback = 1;
}

void nox_xxx_collideSulphurShot_4E9E50(
	nox_object_t* source,
	nox_object_t* target,
	float* collision) {
	seen_source = source;
	seen_target = target;
	seen_collision = collision;
	seen_callback = 2;
}

int main(void) {
	nox_object_t source = {0};
	nox_object_t target = {0};
	float collision[2] = {3.5f, -8.25f};
	nox_projectile_collide_data_t data = {.damage = 7, .field_4 = 0x12345678};
	char positive[] = "2147483647 trailing";
	char negative[] = "-2147483648";
	char invalid[] = "invalid";

	nox_xxx_collideSulphurShot2_4E9D80(&source, &target, collision);
	if (seen_callback != 1 || seen_source != &source || seen_target != &target ||
		seen_collision != collision) {
		return 1;
	}
	nox_xxx_collideSulphurShot_4E9E50(&source, &target, collision);
	if (seen_callback != 2 || seen_source != &source || seen_target != &target ||
		seen_collision != collision) {
		return 2;
	}
	if (sub_536D80(positive, &data) != 1 || data.damage != INT32_MAX ||
		data.field_4 != 0x12345678) {
		return 3;
	}
	if (sub_536D80(negative, &data) != 1 || data.damage != INT32_MIN ||
		data.field_4 != 0x12345678) {
		return 4;
	}
	if (sub_536D80(invalid, &data) != 1 || data.damage != INT32_MIN ||
		data.field_4 != 0x12345678) {
		return 5;
	}
	return 0;
}
