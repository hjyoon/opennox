// Suppress unrelated Win32-only assertions while parsing the shared header,
// then assert every native-width field and ABI value used by 004E8DF0.
#define _Static_assert(...)
#include "../GAME3_3.h"
#undef _Static_assert

#include <stddef.h>
#include <stdint.h>

#define EXPECT_NATIVE(field, off32, off64) \
	_Static_assert(offsetof(nox_object_t, field) == (sizeof(void*) == 4 ? (off32) : (off64)), \
		"wrong native object offset: " #field)

_Static_assert(sizeof(uintptr_t) == sizeof(void*), "uintptr width");
_Static_assert(sizeof(((nox_object_t*)0)->obj_class) == 4, "object class width");
_Static_assert(sizeof(((nox_object_t*)0)->field_32) == 4, "pickup frame width");
_Static_assert(sizeof(((nox_object_t*)0)->data_update) == sizeof(void*), "update-data pointer width");
EXPECT_NATIVE(obj_class, 8, 12);
EXPECT_NATIVE(field_32, 128, 132);
EXPECT_NATIVE(data_update, 748, 872);

_Static_assert(sizeof(((nox_player_update_data_t*)0)->movement_flags) == 4, "movement flags width");
_Static_assert(offsetof(nox_player_update_data_t, movement_flags) ==
	(sizeof(void*) == 4 ? 240 : 300), "movement flags offset");

static nox_object_t* seen_item;
static nox_object_t* seen_unit;
static float* seen_collision;
static uintptr_t next_result;

uintptr_t nox_xxx_collidePickup_4E8DF0(
	nox_object_t* item,
	nox_object_t* unit,
	float* collision) {
	seen_item = item;
	seen_unit = unit;
	seen_collision = collision;
	return next_result;
}

static uintptr_t (*const pickup_signature)(nox_object_t*, nox_object_t*, float*) =
	nox_xxx_collidePickup_4E8DF0;

int main(void) {
	nox_player_update_data_t update = {0};
	nox_object_t item = {0};
	nox_object_t unit = {0};
	float collision[2] = {2.5f, -7.25f};
	item.field_32 = UINT32_C(0xfedcba98);
	unit.obj_class = UINT32_C(4);
	unit.data_update = &update;
	update.movement_flags = UINT32_C(0x10203041);

	next_result = (uintptr_t)&unit;
	if (pickup_signature(&item, &unit, collision) != (uintptr_t)&unit) {
		return 1;
	}
	if (seen_item != &item || seen_unit != &unit || seen_collision != collision) {
		return 2;
	}
	if (item.field_32 != UINT32_C(0xfedcba98) || unit.obj_class != UINT32_C(4) ||
		unit.data_update != &update || update.movement_flags != UINT32_C(0x10203041)) {
		return 3;
	}

	next_result = (uintptr_t)0;
	if (pickup_signature(0, 0, 0) != (uintptr_t)0) {
		return 4;
	}
	if (seen_item != 0 || seen_unit != 0 || seen_collision != 0) {
		return 5;
	}
	next_result = (uintptr_t)1;
	if (pickup_signature(&item, &unit, collision) != (uintptr_t)1) {
		return 6;
	}
	return 0;
}
