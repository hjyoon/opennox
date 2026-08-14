// Suppress unrelated Win32-only assertions while parsing the shared headers,
// then assert the fixed record and pointer-native callback/parser boundaries.
#define _Static_assert(...)
#include "../GAME3_2.h"
#include "../GAME3_3.h"
#include "../GAME4_3.h"
#undef _Static_assert

#include <stddef.h>
#include <stdint.h>
#include <string.h>

_Static_assert(sizeof(nox_damage_collide_data_t) == 8, "DamageCollide data size");
_Static_assert(offsetof(nox_damage_collide_data_t, damage) == 0,
	"DamageCollide damage offset");
_Static_assert(offsetof(nox_damage_collide_data_t, damage_type) == 4,
	"DamageCollide damage-type offset");
_Static_assert(offsetof(nox_object_t, health_data) == (sizeof(void*) == 4 ? 556 : 616),
	"object health offset");
_Static_assert(offsetof(nox_object_t, collide_data) == (sizeof(void*) == 4 ? 700 : 776),
	"object collide-data offset");
_Static_assert(offsetof(nox_object_t, func_damage) == (sizeof(void*) == 4 ? 716 : 808),
	"object damage callback offset");
_Static_assert(
	__builtin_types_compatible_p(
		__typeof__(&nox_xxx_collideDamage_4E9430),
		void (*)(nox_object_t*, nox_object_t*, float*)),
	"DamageCollide callback pointer width");
_Static_assert(
	__builtin_types_compatible_p(
		__typeof__(&nox_xxx_collideDamageLoad_536E10),
		int (*)(char*, nox_damage_collide_data_t*)),
	"DamageCollide parser pointer width");

static nox_object_t* seen_source;
static nox_object_t* seen_target;
static float* seen_collision;

void nox_xxx_collideDamage_4E9430(
	nox_object_t* source,
	nox_object_t* target,
	float* collision) {
	seen_source = source;
	seen_target = target;
	seen_collision = collision;
}

int nox_xxx_parseDamageTypeByName_4E0A00(const char* name) {
	return strcmp(name, "MAGIC") == 0 ? -7 : 18;
}

int main(void) {
	nox_object_t source = {0};
	nox_object_t target = {0};
	float collision[2] = {3.5f, -8.25f};
	nox_damage_collide_data_t data = {
		.damage = 9,
		.reserved = {0xa1, 0xb2, 0xc3},
		.damage_type = 4,
	};
	char valid[] = "255 MAGIC";
	char invalid[] = "2 UNKNOWN";

	nox_xxx_collideDamage_4E9430(&source, &target, collision);
	if (seen_source != &source || seen_target != &target || seen_collision != collision) {
		return 1;
	}
	if (!nox_xxx_collideDamageLoad_536E10(valid, &data)) {
		return 2;
	}
	if (data.damage != UINT8_C(255) || data.damage_type != -7 ||
		data.reserved[0] != UINT8_C(0xa1) || data.reserved[1] != UINT8_C(0xb2) ||
		data.reserved[2] != UINT8_C(0xc3)) {
		return 3;
	}
	if (nox_xxx_collideDamageLoad_536E10(invalid, &data) != 0 ||
		data.damage != UINT8_C(2) || data.damage_type != 18) {
		return 4;
	}
	return 0;
}
