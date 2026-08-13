// Suppress unrelated Win32-only declarations while the shared legacy headers
// are parsed, then restore and assert every C field exposed by 004E8460.
#define _Static_assert(...)
#include "../GAME3_3.h"
#undef _Static_assert

#include <stddef.h>
#include <stdint.h>

#define EXPECT_NATIVE(field, off32, off64) \
	_Static_assert(offsetof(nox_object_t, field) == (sizeof(void*) == 4 ? (off32) : (off64)), \
		"wrong native object offset: " #field)

_Static_assert(sizeof(((nox_object_t*)0)->obj_class) == 4, "object class width");
_Static_assert(sizeof(((nox_object_t*)0)->obj_flags) == 4, "object flags width");
_Static_assert(sizeof(((nox_object_t*)0)->new_x) == 4, "new X width");
_Static_assert(sizeof(((nox_object_t*)0)->new_y) == 4, "new Y width");
_Static_assert(sizeof(((nox_object_t*)0)->prev_x) == 4, "previous X width");
_Static_assert(sizeof(((nox_object_t*)0)->prev_y) == 4, "previous Y width");
_Static_assert(sizeof(((nox_object_t*)0)->vel_x) == 4, "velocity X width");
_Static_assert(sizeof(((nox_object_t*)0)->vel_y) == 4, "velocity Y width");
_Static_assert(sizeof(((nox_object_t*)0)->mass) == 4, "mass width");
_Static_assert(sizeof(((nox_object_t*)0)->buffs) == 4, "buff mask width");
_Static_assert(sizeof(((nox_object_t*)0)->buffs_dur[0]) == 2, "buff duration width");
_Static_assert(sizeof(((nox_object_t*)0)->buffs_power[0]) == 1, "buff power width");
_Static_assert(sizeof(((nox_object_t*)0)->health_data) == sizeof(void*), "health pointer width");
_Static_assert(sizeof(((nox_object_t*)0)->func_damage) == sizeof(void*), "damage callback width");
_Static_assert(sizeof(((nox_object_t*)0)->data_update) == sizeof(void*), "update data pointer width");

EXPECT_NATIVE(obj_class, 8, 12);
EXPECT_NATIVE(obj_flags, 16, 20);
EXPECT_NATIVE(new_x, 64, 68);
EXPECT_NATIVE(new_y, 68, 72);
EXPECT_NATIVE(prev_x, 72, 76);
EXPECT_NATIVE(prev_y, 76, 80);
EXPECT_NATIVE(vel_x, 80, 84);
EXPECT_NATIVE(vel_y, 84, 88);
EXPECT_NATIVE(mass, 120, 124);
EXPECT_NATIVE(buffs, 340, 344);
EXPECT_NATIVE(buffs_dur, 344, 348);
EXPECT_NATIVE(buffs_power, 408, 412);
EXPECT_NATIVE(health_data, 556, 616);
EXPECT_NATIVE(func_damage, 716, 808);
EXPECT_NATIVE(data_update, 748, 872);

_Static_assert(sizeof(((nox_player_update_data_t*)0)->collision_wall) == sizeof(void*),
	"collision wall pointer width");
_Static_assert(offsetof(nox_player_update_data_t, collision_wall) ==
	(sizeof(void*) == 4 ? 296 : 360), "collision wall offset");

static nox_object_t* seen_player;
static nox_object_t* seen_other;
static float* seen_collision;

void nox_xxx_collidePlayer_4E8460(nox_object_t* player, nox_object_t* other, float* collision) {
	seen_player = player;
	seen_other = other;
	seen_collision = collision;
}

static void (*const player_signature)(nox_object_t*, nox_object_t*, float*) =
	nox_xxx_collidePlayer_4E8460;

int main(void) {
	nox_object_t player = {0};
	nox_object_t other = {0};
	float collision[2] = {-3.0f, 9.0f};
	player_signature(&player, &other, collision);
	if (seen_player != &player || seen_other != &other || seen_collision != collision) {
		return 1;
	}
	player_signature(0, 0, 0);
	if (seen_player != 0 || seen_other != 0 || seen_collision != 0) {
		return 2;
	}
	return 0;
}
