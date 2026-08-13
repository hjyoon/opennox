// Suppress unrelated Win32-only declarations while the shared legacy headers
// are parsed, then restore and assert every field used by 004D6A20/004E8340/
// 004E8390 on the selected target.
#define _Static_assert(...)
#include "../GAME3_2.h"
#include "../GAME3_3.h"
#undef _Static_assert

#include <limits.h>
#include <stddef.h>
#include <stdint.h>

_Static_assert(sizeof(nox_point) == 8, "door target point size");
_Static_assert(offsetof(nox_point, x) == 0, "door target X offset");
_Static_assert(offsetof(nox_point, y) == 4, "door target Y offset");
_Static_assert(sizeof(nox_door_update_data_t) == 52, "DoorUpdate size");
_Static_assert(offsetof(nox_door_update_data_t, lock_code) == 1, "DoorUpdate lock offset");
_Static_assert(offsetof(nox_door_update_data_t, target_direction) == 4, "DoorUpdate target direction offset");
_Static_assert(offsetof(nox_door_update_data_t, synced_direction) == 8, "DoorUpdate synchronized direction offset");
_Static_assert(offsetof(nox_door_update_data_t, current_direction) == 12, "DoorUpdate current direction offset");
_Static_assert(offsetof(nox_door_update_data_t, tile_x) == 16, "DoorUpdate X offset");
_Static_assert(offsetof(nox_door_update_data_t, tile_y) == 20, "DoorUpdate Y offset");
_Static_assert(offsetof(nox_door_update_data_t, quest_sync) == 48, "DoorUpdate Quest-sync offset");
_Static_assert(sizeof(((nox_object_t*)0)->obj_class) == 4, "object class width");
_Static_assert(sizeof(((nox_object_t*)0)->extent) == 4, "object extent width");
_Static_assert(offsetof(nox_object_t, obj_class) == (sizeof(void*) == 4 ? 8 : 12), "object class offset");
_Static_assert(offsetof(nox_object_t, extent) == (sizeof(void*) == 4 ? 40 : 44), "object extent offset");
_Static_assert(offsetof(nox_object_t, data_update) == (sizeof(void*) == 4 ? 748 : 872),
	"object update-data offset");

static int32_t send_result;
static int32_t sent_recipient;
static nox_object_t* sent_object;
static uint8_t sent_packet[4];
static int32_t quest_result;
static int sync_calls;

int32_t sub_4D6A20(int32_t recipient, nox_object_t* object) {
	const uint16_t extent = (uint16_t)object->extent;
	sent_recipient = recipient;
	sent_object = object;
	sent_packet[0] = UINT8_C(0xf0);
	sent_packet[1] = UINT8_C(0x0f);
	sent_packet[2] = (uint8_t)extent;
	sent_packet[3] = (uint8_t)(extent >> 8);
	return send_result;
}

int32_t sub_4E8390(nox_object_t* door) {
	nox_door_update_data_t* update = (nox_door_update_data_t*)door->data_update;
	update->quest_sync = UINT8_C(1);
	++sync_calls;
	return sub_4D6A20(INT32_C(255), door);
}

void nox_xxx_fnFindCloseDoors_4E8340(nox_object_t* door, nox_point* target) {
	if (((uint8_t)door->obj_class & UINT8_C(0x80)) == 0) {
		return;
	}
	nox_door_update_data_t* update = (nox_door_update_data_t*)door->data_update;
	if (update->tile_x != target->x) {
		return;
	}
	if (update->tile_y != target->y) {
		return;
	}
	update->lock_code = UINT8_C(0);
	if (quest_result != 0) {
		(void)sub_4E8390(door);
	}
}

static void (*const close_signature)(nox_object_t*, nox_point*) = nox_xxx_fnFindCloseDoors_4E8340;
static int32_t (*const sync_signature)(nox_object_t*) = sub_4E8390;
static int32_t (*const packet_signature)(int32_t, nox_object_t*) = sub_4D6A20;

int main(void) {
	nox_object_t door = {0};
	nox_door_update_data_t update = {0};
	nox_point target = {.x = INT32_C(-17), .y = INT32_MAX};
	door.obj_class = UINT32_C(0xa5000080);
	door.extent = UINT32_C(0xa5a5bcde);
	door.data_update = &update;
	update.lock_code = UINT8_C(4);
	update.tile_x = INT32_C(-17);
	update.tile_y = INT32_MAX;
	update.quest_sync = UINT8_C(0x7a);

	quest_result = INT32_C(0);
	close_signature(&door, &target);
	if (update.lock_code != 0 || update.quest_sync != UINT8_C(0x7a) || sync_calls != 0) {
		return 1;
	}

	update.lock_code = UINT8_MAX;
	quest_result = INT32_MIN;
	send_result = INT32_MIN;
	close_signature(&door, &target);
	if (update.lock_code != 0 || update.quest_sync != 1 || sync_calls != 1) {
		return 2;
	}
	if (sent_recipient != 255 || sent_object != &door || sent_packet[0] != UINT8_C(0xf0) ||
		sent_packet[1] != UINT8_C(0x0f) || sent_packet[2] != UINT8_C(0xde) || sent_packet[3] != UINT8_C(0xbc)) {
		return 3;
	}
	if (sync_signature(&door) != INT32_MIN || packet_signature(INT32_MIN, &door) != INT32_MIN) {
		return 4;
	}

	update.lock_code = UINT8_C(3);
	target.x = INT32_C(-16);
	close_signature(&door, &target);
	if (update.lock_code != UINT8_C(3) || sync_calls != 2) {
		return 5;
	}
	return 0;
}
