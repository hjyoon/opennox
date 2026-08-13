// Suppress unrelated Win32-only declarations while the shared legacy headers
// are parsed, then restore and assert each C boundary used by 004E8880.
#define _Static_assert(...)
#include "../GAME3_3.h"
#include "../GAME4_3.h"
#undef _Static_assert

#include <stddef.h>
#include <stdint.h>

#define EXPECT_NATIVE(field, off32, off64) \
	_Static_assert(offsetof(nox_object_t, field) == (sizeof(void*) == 4 ? (off32) : (off64)), \
		"wrong native object offset: " #field)

_Static_assert(sizeof(nox_projectile_collide_data_t) == 8, "projectile collide-data size");
_Static_assert(offsetof(nox_projectile_collide_data_t, damage) == 0, "projectile damage offset");
_Static_assert(offsetof(nox_projectile_collide_data_t, field_4) == 4, "projectile field_4 offset");
_Static_assert(sizeof(((nox_object_t*)0)->new_x) == 4, "new X width");
_Static_assert(sizeof(((nox_object_t*)0)->new_y) == 4, "new Y width");
_Static_assert(sizeof(((nox_object_t*)0)->collide_data) == sizeof(void*), "collide-data pointer width");
_Static_assert(sizeof(((nox_object_t*)0)->func_damage) == sizeof(void*), "damage callback width");
EXPECT_NATIVE(new_x, 64, 68);
EXPECT_NATIVE(new_y, 68, 72);
EXPECT_NATIVE(collide_data, 700, 776);
EXPECT_NATIVE(func_damage, 716, 808);
_Static_assert(
	__builtin_types_compatible_p(__typeof__(&sub_536D80), int (*)(char*, void*)),
	"ProjectileSparkCollide parser pointer width");

static nox_object_t* seen_projectile;
static nox_object_t* seen_other;
static float* seen_collision;

void nox_xxx_collideProjectileSpark_4E8880(
	nox_object_t* projectile,
	nox_object_t* other,
	float* collision) {
	seen_projectile = projectile;
	seen_other = other;
	seen_collision = collision;
}

static void (*const projectile_spark_signature)(nox_object_t*, nox_object_t*, float*) =
	nox_xxx_collideProjectileSpark_4E8880;

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
	projectile.new_x = 69.0f;
	projectile.new_y = -46.0f;
	projectile.collide_data = &data;
	other.func_damage = damage_signature;
	projectile_spark_signature(&projectile, &other, collision);
	if (seen_projectile != &projectile || seen_other != &other || seen_collision != collision) {
		return 1;
	}
	if (!other.func_damage(&other, &projectile, &projectile, data.damage, 11)) {
		return 2;
	}
	projectile_spark_signature(0, 0, 0);
	if (seen_projectile != 0 || seen_other != 0 || seen_collision != 0) {
		return 3;
	}
	if (data.field_4 != 0x12345678 || projectile.new_x != 69.0f || projectile.new_y != -46.0f) {
		return 4;
	}
	return 0;
}
