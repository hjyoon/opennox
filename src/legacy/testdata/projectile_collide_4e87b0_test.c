// Suppress unrelated Win32-only declarations while the shared legacy headers
// are parsed, then restore and assert each C boundary used by 004E87B0.
#define _Static_assert(...)
#include "../GAME3_3.h"
#include "../GAME4_3.h"
#undef _Static_assert

#include <stddef.h>
#include <stdint.h>

_Static_assert(sizeof(nox_projectile_collide_data_t) == 8, "projectile collide-data size");
_Static_assert(offsetof(nox_projectile_collide_data_t, damage) == 0, "projectile damage offset");
_Static_assert(offsetof(nox_projectile_collide_data_t, field_4) == 4, "projectile field_4 offset");
_Static_assert(offsetof(nox_object_t, typ_ind) == (sizeof(void*) == 4 ? 4 : 8),
	"object type-index offset");
_Static_assert(offsetof(nox_object_t, collide_data) == (sizeof(void*) == 4 ? 700 : 776),
	"object collide-data offset");
_Static_assert(offsetof(nox_object_t, func_damage) == (sizeof(void*) == 4 ? 716 : 808),
	"object damage callback offset");
_Static_assert(
	__builtin_types_compatible_p(__typeof__(&sub_536D80), int (*)(char*, void*)),
	"ProjectileCollide parser pointer width");

static nox_object_t* seen_projectile;
static nox_object_t* seen_other;
static float* seen_collision;

void nox_xxx_collideProjectileGeneric_4E87B0(
	nox_object_t* projectile,
	nox_object_t* other,
	float* collision) {
	seen_projectile = projectile;
	seen_other = other;
	seen_collision = collision;
}

static void (*const projectile_signature)(nox_object_t*, nox_object_t*, float*) =
	nox_xxx_collideProjectileGeneric_4E87B0;

static int damage_signature(
	nox_object_t* target,
	nox_object_t* source,
	nox_object_t* attacker,
	int32_t damage,
	int32_t damage_type) {
	return target != 0 && source != 0 && attacker != 0 && damage == -7 && damage_type == 11;
}

int main(void) {
	nox_object_t projectile = {0};
	nox_object_t other = {0};
	nox_projectile_collide_data_t data = {.damage = -7, .field_4 = 0x12345678};
	float collision[2] = {-3.0f, 9.0f};
	projectile.collide_data = &data;
	other.func_damage = damage_signature;
	projectile_signature(&projectile, &other, collision);
	if (seen_projectile != &projectile || seen_other != &other || seen_collision != collision) {
		return 1;
	}
	if (!other.func_damage(&other, &projectile, &projectile, data.damage, 11)) {
		return 2;
	}
	projectile_signature(0, 0, 0);
	if (seen_projectile != 0 || seen_other != 0 || seen_collision != 0) {
		return 3;
	}
	if (data.field_4 != 0x12345678) {
		return 4;
	}
	return 0;
}
