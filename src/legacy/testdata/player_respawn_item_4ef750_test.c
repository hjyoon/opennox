// Suppress unrelated Win32-only declarations while the shared structures are
// parsed, then restore and assert every C boundary consumed by 004EF750.
#define _Static_assert(...)
#include "../GAME3_3.h"
#undef _Static_assert

#include <limits.h>
#include <stddef.h>
#include <stdint.h>

#define EXPECT_NATIVE(field, off32, off64) \
	_Static_assert(offsetof(nox_object_t, field) == (sizeof(void*) == 4 ? (off32) : (off64)), \
		"wrong native object offset: " #field)

typedef nox_object_t* (*player_respawn_item_fn)(
	nox_object_t*, char*, nox_modifier_attrs_t*, int32_t, int32_t);

_Static_assert(CHAR_BIT == 8, "bytes must remain eight bits");
_Static_assert(sizeof(void*) == 4 || sizeof(void*) == 8, "unsupported pointer width");
_Static_assert(sizeof(int32_t) == 4, "placement arguments must remain exact int32");
_Static_assert(sizeof(((nox_object_t*)0)->obj_class) == 4, "native class storage width");
_Static_assert(sizeof(((nox_object_t*)0)->obj_flags) == 4, "native flag storage width");
_Static_assert(sizeof(((nox_object_t*)0)->func_init) == sizeof(void*), "initializer pointer width");
_Static_assert(sizeof(((nox_object_t*)0)->data_update) == sizeof(void*), "update-data pointer width");
_Static_assert(sizeof(nox_modifier_attrs_t) == (sizeof(void*) == 4 ? 20 : 40), "modifier attributes size");
_Static_assert(offsetof(nox_modifier_attrs_t, field_16) == 4 * sizeof(void*), "modifier field_16 offset");
EXPECT_NATIVE(obj_class, 8, 12);
EXPECT_NATIVE(obj_flags, 16, 20);
EXPECT_NATIVE(func_init, 688, 752);
EXPECT_NATIVE(data_update, 748, 872);
_Static_assert(
	_Generic(&nox_xxx_playerRespawnItem_4EF750, player_respawn_item_fn: 1, default: 0),
	"respawn-item creation must use native pointers and exact int32 scalars");

static nox_object_t* observed_player;
static char* observed_type_id;
static nox_modifier_attrs_t* observed_attrs;
static int32_t observed_a4;
static int32_t observed_a5;
static nox_object_t* next_result;
static unsigned int observed_calls;

nox_object_t* nox_xxx_playerRespawnItem_4EF750(
		nox_object_t* player,
		char* type_id,
		nox_modifier_attrs_t* attrs,
		int32_t a4,
		int32_t a5) {
	observed_player = player;
	observed_type_id = type_id;
	observed_attrs = attrs;
	observed_a4 = a4;
	observed_a5 = a5;
	++observed_calls;
	return next_result;
}

static void check_call(
		player_respawn_item_fn create_item,
		nox_object_t* player,
		char* type_id,
		nox_modifier_attrs_t* attrs,
		int32_t a4,
		int32_t a5,
		nox_object_t* result) {
	next_result = result;
	if (create_item(player, type_id, attrs, a4, a5) != result ||
		observed_player != player || observed_type_id != type_id ||
		observed_attrs != attrs || observed_a4 != a4 || observed_a5 != a5) {
		__builtin_trap();
	}
}

int main(void) {
	nox_object_t player = {0};
	nox_object_t item = {0};
	nox_modifier_attrs_t attrs = {0};
	char type_id[] = "Longsword";
	player_respawn_item_fn const create_item = nox_xxx_playerRespawnItem_4EF750;

	check_call(create_item, &player, type_id, &attrs, INT32_MIN, INT32_MAX, &item);
	check_call(create_item, NULL, NULL, NULL, INT32_MAX, INT32_MIN, NULL);
	if (observed_calls != 2) {
		return 1;
	}
	return 0;
}
