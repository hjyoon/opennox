// Suppress unrelated Win32-only declarations while the shared header is
// parsed, then restore and assert every C boundary consumed by 004E8910.
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
_Static_assert(sizeof(((nox_object_t*)0)->typ_ind) == 2, "native TypeInd storage width");
_Static_assert(sizeof(((nox_object_t*)0)->obj_class) == 4, "native class storage width");
_Static_assert(sizeof(((nox_object_t*)0)->inv_next_item) == sizeof(void*), "inventory next pointer width");
_Static_assert(sizeof(((nox_object_t*)0)->inv_first_item) == sizeof(void*), "inventory first pointer width");
_Static_assert(sizeof(((nox_object_t*)0)->owner) == sizeof(void*), "owner pointer width");
_Static_assert(sizeof(((nox_object_t*)0)->data_update) == sizeof(void*), "update-data pointer width");
EXPECT_NATIVE(typ_ind, 4, 8);
EXPECT_NATIVE(obj_class, 8, 12);
EXPECT_NATIVE(inv_next_item, 496, 528);
EXPECT_NATIVE(inv_first_item, 504, 544);
EXPECT_NATIVE(owner, 508, 552);
EXPECT_NATIVE(data_update, 748, 872);

static nox_object_t* seen_unit;
static nox_object_t* seen_door;

nox_object_t* nox_xxx_doorGetSomeKey_4E8910(nox_object_t* unit, nox_object_t* door) {
	seen_unit = unit;
	seen_door = door;
	return unit;
}

static nox_object_t* (*const door_key_signature)(nox_object_t*, nox_object_t*) =
	nox_xxx_doorGetSomeKey_4E8910;

int main(void) {
	nox_object_t unit = {0};
	nox_object_t door = {0};
	nox_object_t key = {0};
	nox_door_update_data_t update = {0};
	unit.typ_ind = UINT16_C(7);
	unit.obj_class = UINT32_C(0x4);
	unit.inv_first_item = &key;
	key.inv_next_item = 0;
	door.data_update = &update;
	update.lock_code = UINT8_C(2);

	if (door_key_signature(&unit, &door) != &unit || seen_unit != &unit || seen_door != &door) {
		return 1;
	}
	if (unit.typ_ind != UINT16_C(7) || unit.inv_first_item != &key ||
		door.data_update != &update || update.lock_code != UINT8_C(2)) {
		return 2;
	}
	if (door_key_signature(0, 0) != 0 || seen_unit != 0 || seen_door != 0) {
		return 3;
	}
	return 0;
}
