// Suppress unrelated Win32-only assertions while parsing the shared header,
// then assert SparkExplosionCollide's native pointer and one-byte data ABI.
#define _Static_assert(...)
#include "../GAME3_3.h"
#include "../GAME4_3.h"
#undef _Static_assert

#include <stddef.h>
#include <stdint.h>

_Static_assert(sizeof(nox_spark_explosion_collide_data_t) == 1,
	"SparkExplosionCollide data size");
_Static_assert(offsetof(nox_spark_explosion_collide_data_t, power) == 0,
	"SparkExplosionCollide power offset");
_Static_assert(offsetof(nox_object_t, x) == (sizeof(void*) == 4 ? 56 : 60),
	"object position offset");
_Static_assert(offsetof(nox_object_t, direction1) == (sizeof(void*) == 4 ? 124 : 128),
	"object direction offset");
_Static_assert(offsetof(nox_object_t, collide_data) == (sizeof(void*) == 4 ? 700 : 776),
	"object collide-data offset");
_Static_assert(offsetof(nox_object_t, func_damage) == (sizeof(void*) == 4 ? 716 : 808),
	"object damage callback offset");
_Static_assert(
	__builtin_types_compatible_p(
		__typeof__(&nox_xxx_fireballCollide_4E9AC0),
		void (*)(nox_object_t*, nox_object_t*, float*)),
	"SparkExplosionCollide callback pointer width");
_Static_assert(
	__builtin_types_compatible_p(
		__typeof__(&nox_xxx_collideSparkExplosionLoad_536DE0),
		int (*)(char*, nox_spark_explosion_collide_data_t*)),
	"SparkExplosionCollide parser data width");

static nox_object_t* seen_source;
static nox_object_t* seen_target;
static float* seen_collision;

void nox_xxx_fireballCollide_4E9AC0(
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
	nox_spark_explosion_collide_data_t data = {0};
	char positive[] = "511 trailing";
	char negative[] = "-2";
	char invalid[] = "invalid";

	nox_xxx_fireballCollide_4E9AC0(&source, &target, collision);
	if (seen_source != &source || seen_target != &target || seen_collision != collision) {
		return 1;
	}
	if (nox_xxx_collideSparkExplosionLoad_536DE0(positive, &data) != 1 || data.power != 255) {
		return 2;
	}
	if (nox_xxx_collideSparkExplosionLoad_536DE0(negative, &data) != 1 || data.power != 254) {
		return 3;
	}
	if (nox_xxx_collideSparkExplosionLoad_536DE0(invalid, &data) != 1 ||
		data.power != (uint8_t)(uintptr_t)invalid) {
		return 4;
	}
	return 0;
}
