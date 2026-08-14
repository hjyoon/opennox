// Suppress unrelated Win32-only assertions while parsing the shared headers,
// then assert only HarpoonCollide's native object, player-data and callback ABI.
#define _Static_assert(...)
#include "../GAME3_3.h"
#undef _Static_assert

#include <stddef.h>
#include <stdint.h>

_Static_assert(sizeof(nox_harpoon_collide_data_t) ==
	(sizeof(void*) == 4 ? 8 : 16), "Harpoon collide-data size");
_Static_assert(offsetof(nox_harpoon_collide_data_t, field_0) == 0,
	"Harpoon collide field-zero offset");
_Static_assert(offsetof(nox_harpoon_collide_data_t, owner) == sizeof(void*),
	"Harpoon collide owner offset");

_Static_assert(offsetof(nox_player_update_data_t, harpoon_target) ==
	(sizeof(void*) == 4 ? 132 : 152), "Player Harpoon target offset");
_Static_assert(offsetof(nox_player_update_data_t, harpoon_bolt) ==
	(sizeof(void*) == 4 ? 136 : 160), "Player Harpoon bolt offset");
_Static_assert(offsetof(nox_player_update_data_t, harpoon_field_35) ==
	(sizeof(void*) == 4 ? 140 : 168), "Player Harpoon field-35 offset");
_Static_assert(offsetof(nox_player_update_data_t, harpoon_target_x) ==
	(sizeof(void*) == 4 ? 144 : 172), "Player Harpoon target-X offset");
_Static_assert(offsetof(nox_player_update_data_t, harpoon_target_y) ==
	(sizeof(void*) == 4 ? 148 : 176), "Player Harpoon target-Y offset");
_Static_assert(offsetof(nox_player_update_data_t, harpoon_frame) ==
	(sizeof(void*) == 4 ? 152 : 180), "Player Harpoon frame offset");

_Static_assert(offsetof(nox_object_t, obj_class) ==
	(sizeof(void*) == 4 ? 8 : 12), "object class offset");
_Static_assert(offsetof(nox_object_t, obj_flags) ==
	(sizeof(void*) == 4 ? 16 : 20), "object flags offset");
_Static_assert(offsetof(nox_object_t, new_x) ==
	(sizeof(void*) == 4 ? 64 : 68), "object new-X offset");
_Static_assert(offsetof(nox_object_t, new_y) ==
	(sizeof(void*) == 4 ? 68 : 72), "object new-Y offset");
_Static_assert(offsetof(nox_object_t, owner) ==
	(sizeof(void*) == 4 ? 508 : 552), "object owner offset");
_Static_assert(offsetof(nox_object_t, func_damage) ==
	(sizeof(void*) == 4 ? 716 : 808), "object damage callback offset");
_Static_assert(offsetof(nox_object_t, data_update) ==
	(sizeof(void*) == 4 ? 748 : 872), "object update-data offset");

_Static_assert(__builtin_types_compatible_p(
	__typeof__(&nox_xxx_collideHarpoon_4EB6A0),
	void (*)(nox_object_t*, nox_object_t*, float*)),
	"Harpoon collide callback pointer width");

static nox_object_t* seen_source;
static nox_object_t* seen_target;
static float* seen_collision;

void nox_xxx_collideHarpoon_4EB6A0(
	nox_object_t* source,
	nox_object_t* target,
	float* collision) {
	seen_source = source;
	seen_target = target;
	seen_collision = collision;
}

int main(void) {
	nox_object_t owner = {0};
	nox_object_t source = {0};
	nox_object_t target = {0};
	nox_harpoon_collide_data_t collide = {
		.field_0 = UINT32_C(0x89abcdef),
		.owner = &owner,
	};
	nox_player_update_data_t player = {
		.harpoon_target = &target,
		.harpoon_bolt = &source,
		.harpoon_field_35 = UINT32_C(0x12345678),
		.harpoon_target_x = 12.5f,
		.harpoon_target_y = -4.25f,
		.harpoon_frame = UINT32_C(9876),
	};
	float collision[2] = {3.5f, -8.25f};

	source.collide_data = &collide;
	owner.data_update = &player;
	nox_xxx_collideHarpoon_4EB6A0(&source, &target, collision);
	if (seen_source != &source || seen_target != &target ||
		seen_collision != collision || collide.owner != &owner) {
		return 1;
	}
	if (collide.field_0 != UINT32_C(0x89abcdef) ||
		player.harpoon_target != &target || player.harpoon_bolt != &source ||
		player.harpoon_field_35 != UINT32_C(0x12345678) ||
		player.harpoon_target_x != 12.5f || player.harpoon_target_y != -4.25f ||
		player.harpoon_frame != UINT32_C(9876)) {
		return 2;
	}
	return 0;
}
