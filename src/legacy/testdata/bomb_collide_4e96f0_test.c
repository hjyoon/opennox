// Suppress unrelated Win32-only assertions while parsing the shared headers,
// then assert the fixed collide record and pointer-native callback boundaries.
#define _Static_assert(...)
#include "../GAME3_3.h"
#undef _Static_assert

#include <stddef.h>
#include <stdint.h>

_Static_assert(sizeof(nox_bomb_collide_data_t) == 8,
	"BombCollide data size");
_Static_assert(offsetof(nox_bomb_collide_data_t, reserved) == 0,
	"BombCollide reserved offset");
_Static_assert(offsetof(nox_object_t, obj_class) == (sizeof(void*) == 4 ? 8 : 12),
	"object class offset");
_Static_assert(offsetof(nox_object_t, obj_flags) == (sizeof(void*) == 4 ? 16 : 20),
	"object flags offset");
_Static_assert(offsetof(nox_object_t, field_12) == (sizeof(void*) == 4 ? 48 : 52),
	"object team offset");
_Static_assert(offsetof(nox_object_t, owner) == (sizeof(void*) == 4 ? 508 : 552),
	"object owner offset");
_Static_assert(offsetof(nox_object_t, data_update) == (sizeof(void*) == 4 ? 748 : 872),
	"object update-data offset");
_Static_assert(
	__builtin_types_compatible_p(
		__typeof__(&nox_xxx_collideBomb_4E96F0),
		void (*)(nox_object_t*, nox_object_t*, float*)),
	"BombCollide callback pointer width");
_Static_assert(
	__builtin_types_compatible_p(
		__typeof__(&nox_xxx_unitsHaveSameTeam_4EC520),
		int (*)(nox_object_t*, nox_object_t*)),
	"same-team callback pointer width");

static nox_object_t* seen_bomb;
static nox_object_t* seen_other;
static float* seen_collision;

void nox_xxx_collideBomb_4E96F0(
	nox_object_t* bomb,
	nox_object_t* other,
	float* collision) {
	seen_bomb = bomb;
	seen_other = other;
	seen_collision = collision;
}

int main(void) {
	nox_object_t bomb = {0};
	nox_object_t other = {0};
	float collision[2] = {3.5f, -8.25f};
	nox_bomb_collide_data_t data = {
		.reserved = {UINT8_C(1), UINT8_C(2), UINT8_C(3), UINT8_C(4),
			UINT8_C(5), UINT8_C(6), UINT8_C(7), UINT8_C(8)},
	};

	nox_xxx_collideBomb_4E96F0(&bomb, &other, collision);
	if (seen_bomb != &bomb || seen_other != &other || seen_collision != collision) {
		return 1;
	}
	for (size_t i = 0; i < sizeof(data.reserved); ++i) {
		if (data.reserved[i] != i + 1) {
			return 2;
		}
	}
	return 0;
}
