#include <assert.h>
#include <limits.h>
#include <stddef.h>
#include <stdint.h>

#include "../xfer_hole_4f51d0.h"

struct nox_object_t {
	uintptr_t marker;
};

typedef struct nox_script_callback_t {
	uint32_t flags;
	int32_t function;
} nox_script_callback_t;

typedef struct nox_hole_collide_data_t {
	nox_script_callback_t script;
	int32_t destination_x;
	int32_t destination_y;
	uint32_t destination_extent;
	uint16_t destination_net_code;
	uint16_t reserved_22;
	uint32_t field_24;
} nox_hole_collide_data_t;

typedef int32_t (*hole_xfer_fn_4F51D0)(nox_object_t*, void*);

_Static_assert(CHAR_BIT == 8, "serialized bytes must remain eight bits");
_Static_assert(sizeof(void*) == 4 || sizeof(void*) == 8,
	"unsupported pointer width");
_Static_assert(sizeof(int32_t) == 4,
	"HoleXfer result must remain int32");
_Static_assert(sizeof(nox_script_callback_t) == 8,
	"script callback must remain eight bytes");
_Static_assert(offsetof(nox_script_callback_t, function) == 4,
	"script function must remain at byte 4");
_Static_assert(sizeof(nox_hole_collide_data_t) == 28,
	"Hole collide data must remain 28 bytes");
_Static_assert(offsetof(nox_hole_collide_data_t, script) == 0,
	"Hole script callback must remain at byte 0");
_Static_assert(offsetof(nox_hole_collide_data_t, destination_x) == 8,
	"Hole destination X must remain at byte 8");
_Static_assert(offsetof(nox_hole_collide_data_t, destination_y) == 12,
	"Hole destination Y must remain at byte 12");
_Static_assert(offsetof(nox_hole_collide_data_t, destination_extent) == 16,
	"Hole destination extent must remain at byte 16");
_Static_assert(offsetof(nox_hole_collide_data_t, destination_net_code) == 20,
	"Hole destination net code must remain at byte 20");
_Static_assert(offsetof(nox_hole_collide_data_t, reserved_22) == 22,
	"Hole reserved slot must remain at byte 22");
_Static_assert(offsetof(nox_hole_collide_data_t, field_24) == 24,
	"Hole field 24 must remain at byte 24");
_Static_assert(
	_Generic(&nox_xxx_XFerHole_4F51D0,
		hole_xfer_fn_4F51D0: 1, default: 0),
	"HoleXfer must preserve two native pointers and its exact int32 result");

static nox_object_t* observed_object;
static void* observed_context;

int32_t nox_xxx_XFerHole_4F51D0(
	nox_object_t* object,
	void* context) {
	observed_object = object;
	observed_context = context;
	return object != NULL;
}

int main(void) {
	nox_object_t object = {.marker = UINTPTR_MAX};
	void* const context = (void*)(uintptr_t)(UINTPTR_MAX - 15u);
	hole_xfer_fn_4F51D0 const transfer = nox_xxx_XFerHole_4F51D0;

	assert(transfer(&object, context) == 1);
	assert(observed_object == &object);
	assert(observed_object->marker == UINTPTR_MAX);
	assert(observed_context == context);

	assert(transfer(NULL, NULL) == 0);
	assert(observed_object == NULL);
	assert(observed_context == NULL);
	return 0;
}
