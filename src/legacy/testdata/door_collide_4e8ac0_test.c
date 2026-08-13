// Suppress unrelated Win32-only assertions while parsing the shared header,
// then assert every fixed-width record and native pointer used by 004E8AC0.
#define _Static_assert(...)
#include "../GAME3_3.h"
#undef _Static_assert

#include <stddef.h>
#include <stdint.h>

#define EXPECT_NATIVE(field, off32, off64) \
	_Static_assert(offsetof(nox_object_t, field) == (sizeof(void*) == 4 ? (off32) : (off64)), \
		"wrong native object offset: " #field)

_Static_assert(sizeof(nox_door_update_data_t) == 52, "DoorUpdate size");
_Static_assert(offsetof(nox_door_update_data_t, lock_code) == 1, "DoorUpdate lock offset");
_Static_assert(offsetof(nox_door_update_data_t, target_direction) == 4, "DoorUpdate target direction offset");
_Static_assert(offsetof(nox_door_update_data_t, synced_direction) == 8, "DoorUpdate synchronized direction offset");
_Static_assert(offsetof(nox_door_update_data_t, current_direction) == 12, "DoorUpdate current direction offset");
_Static_assert(offsetof(nox_door_update_data_t, tile_x) == 16, "DoorUpdate X offset");
_Static_assert(offsetof(nox_door_update_data_t, tile_y) == 20, "DoorUpdate Y offset");
_Static_assert(sizeof(((nox_object_t*)0)->obj_class) == 4, "object class width");
_Static_assert(sizeof(((nox_object_t*)0)->obj_subclass) == 4, "object subclass width");
_Static_assert(sizeof(((nox_object_t*)0)->field_34) == 4, "owner expiry frame width");
_Static_assert(sizeof(((nox_object_t*)0)->inv_holder) == sizeof(void*), "inventory holder pointer width");
_Static_assert(sizeof(((nox_object_t*)0)->owner) == sizeof(void*), "owner pointer width");
_Static_assert(sizeof(((nox_object_t*)0)->data_update) == sizeof(void*), "update-data pointer width");
EXPECT_NATIVE(obj_class, 8, 12);
EXPECT_NATIVE(obj_subclass, 12, 16);
EXPECT_NATIVE(field_34, 136, 140);
EXPECT_NATIVE(inv_holder, 492, 520);
EXPECT_NATIVE(owner, 508, 552);
EXPECT_NATIVE(data_update, 748, 872);

static nox_object_t* seen_door;
static nox_object_t* seen_unit;
static float* seen_collision;

void nox_xxx_collideDoor_4E8AC0(
	nox_object_t* door,
	nox_object_t* unit,
	float* collision) {
	seen_door = door;
	seen_unit = unit;
	seen_collision = collision;
}

static void (*const door_collide_signature)(nox_object_t*, nox_object_t*, float*) =
	nox_xxx_collideDoor_4E8AC0;

int main(void) {
	nox_door_update_data_t update = {
		.lock_code = UINT8_C(2),
		.target_direction = INT32_C(24),
		.synced_direction = INT32_C(16),
		.current_direction = INT32_C(24),
		.tile_x = INT32_C(-4),
		.tile_y = INT32_C(7),
	};
	nox_object_t door = {0};
	nox_object_t unit = {0};
	nox_object_t holder = {0};
	float collision[2] = {12.5f, -3.25f};
	door.obj_class = UINT32_C(0x80);
	door.obj_subclass = UINT32_C(4);
	door.field_34 = UINT32_C(0xfedcba98);
	door.owner = &unit;
	door.data_update = &update;
	unit.inv_holder = &holder;

	door_collide_signature(&door, &unit, collision);
	if (seen_door != &door || seen_unit != &unit || seen_collision != collision) {
		return 1;
	}
	if (door.owner != &unit || door.data_update != &update || unit.inv_holder != &holder) {
		return 2;
	}
	if (update.lock_code != 2 || update.target_direction != 24 ||
		update.synced_direction != 16 || update.current_direction != 24 ||
		update.tile_x != -4 || update.tile_y != 7) {
		return 3;
	}
	door_collide_signature(0, 0, 0);
	if (seen_door != 0 || seen_unit != 0 || seen_collision != 0) {
		return 4;
	}
	return 0;
}
