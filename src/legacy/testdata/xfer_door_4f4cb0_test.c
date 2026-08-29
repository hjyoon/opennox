#include <assert.h>
#include <limits.h>
#include <stddef.h>
#include <stdint.h>

#include "../xfer_door_4f4cb0.h"

struct nox_object_t {
	uintptr_t marker;
};

typedef struct nox_door_update_data_t {
	uint8_t field_0;
	uint8_t lock_code;
	uint8_t reserved_2[2];
	int32_t target_direction;
	int32_t synced_direction;
	int32_t current_direction;
	int32_t tile_x;
	int32_t tile_y;
	uint8_t reserved_24[4];
	uint32_t queued;
	float angular_velocity;
	uint8_t reserved_36[4];
	int16_t fractional_direction;
	uint8_t reserved_42[2];
	uint32_t last_move_frame;
	uint8_t quest_sync;
	uint8_t reserved_49[3];
} nox_door_update_data_t;

typedef int32_t (*door_xfer_fn_4F4CB0)(nox_object_t*, void*);

_Static_assert(CHAR_BIT == 8, "serialized bytes must remain eight bits");
_Static_assert(sizeof(void*) == 4 || sizeof(void*) == 8, "unsupported pointer width");
_Static_assert(sizeof(int32_t) == 4, "DoorXfer result must remain int32");
_Static_assert(sizeof(nox_door_update_data_t) == 52,
	"Door update data must remain 52 bytes");
_Static_assert(offsetof(nox_door_update_data_t, lock_code) == 1,
	"Door lock code must remain at byte 1");
_Static_assert(offsetof(nox_door_update_data_t, target_direction) == 4,
	"Door target direction must remain at byte 4");
_Static_assert(offsetof(nox_door_update_data_t, synced_direction) == 8,
	"Door synchronized direction must remain at byte 8");
_Static_assert(offsetof(nox_door_update_data_t, current_direction) == 12,
	"Door current direction must remain at byte 12");
_Static_assert(offsetof(nox_door_update_data_t, tile_x) == 16,
	"Door tile X must remain at byte 16");
_Static_assert(offsetof(nox_door_update_data_t, tile_y) == 20,
	"Door tile Y must remain at byte 20");
_Static_assert(offsetof(nox_door_update_data_t, queued) == 28,
	"Door queue flag must remain at byte 28");
_Static_assert(offsetof(nox_door_update_data_t, angular_velocity) == 32,
	"Door angular velocity must remain at byte 32");
_Static_assert(offsetof(nox_door_update_data_t, fractional_direction) == 40,
	"Door fractional direction must remain at byte 40");
_Static_assert(offsetof(nox_door_update_data_t, last_move_frame) == 44,
	"Door last-move frame must remain at byte 44");
_Static_assert(offsetof(nox_door_update_data_t, quest_sync) == 48,
	"Door Quest sync byte must remain at byte 48");
_Static_assert(
	_Generic(&nox_xxx_XFerDoor_4F4CB0,
		door_xfer_fn_4F4CB0: 1, default: 0),
	"DoorXfer must preserve two native pointers and its exact int32 result");

static nox_object_t* observed_object;
static void* observed_context;

int32_t nox_xxx_XFerDoor_4F4CB0(nox_object_t* object, void* context) {
	observed_object = object;
	observed_context = context;
	return object != NULL;
}

int main(void) {
	nox_object_t object = {.marker = UINTPTR_MAX};
	void* const context = (void*)(uintptr_t)(UINTPTR_MAX - 15u);
	door_xfer_fn_4F4CB0 const transfer = nox_xxx_XFerDoor_4F4CB0;

	assert(transfer(&object, context) == 1);
	assert(observed_object == &object);
	assert(observed_object->marker == UINTPTR_MAX);
	assert(observed_context == context);

	assert(transfer(NULL, NULL) == 0);
	assert(observed_object == NULL);
	assert(observed_context == NULL);
	return 0;
}
