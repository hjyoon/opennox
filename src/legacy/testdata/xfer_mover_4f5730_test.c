#include <assert.h>
#include <limits.h>
#include <stddef.h>
#include <stdint.h>

#include "../xfer_mover_4f5730.h"

struct nox_object_t {
	uintptr_t marker;
};

typedef int32_t (*mover_xfer_fn_4F5730)(nox_object_t*, void*);

_Static_assert(CHAR_BIT == 8, "serialized bytes must remain eight bits");
_Static_assert(sizeof(void*) == 4 || sizeof(void*) == 8,
	"unsupported pointer width");
_Static_assert(sizeof(int32_t) == 4,
	"MoverXfer result must remain int32");
_Static_assert(sizeof(nox_mover_update_data_t) == 36,
	"Mover update data must remain thirty-six bytes");
_Static_assert(offsetof(nox_mover_update_data_t, field_0) == 0,
	"Mover field-0 must remain at byte 0");
_Static_assert(offsetof(nox_mover_update_data_t, field_1) == 4,
	"Mover field-1 must remain at byte 4");
_Static_assert(offsetof(nox_mover_update_data_t, field_2) == 8,
	"Mover field-2 must remain at byte 8");
_Static_assert(offsetof(nox_mover_update_data_t, waypoint_3_pe32) == 12,
	"Mover first PE32 waypoint slot must remain at byte 12");
_Static_assert(offsetof(nox_mover_update_data_t, waypoint_3_index) == 16,
	"Mover first waypoint index must remain at byte 16");
_Static_assert(offsetof(nox_mover_update_data_t, waypoint_5_pe32) == 20,
	"Mover second PE32 waypoint slot must remain at byte 20");
_Static_assert(offsetof(nox_mover_update_data_t, waypoint_5_index) == 24,
	"Mover second waypoint index must remain at byte 24");
_Static_assert(offsetof(nox_mover_update_data_t, target_pe32) == 28,
	"Mover target PE32 slot must remain at byte 28");
_Static_assert(offsetof(nox_mover_update_data_t, target_extent) == 32,
	"Mover target extent must remain at byte 32");
_Static_assert(
	_Generic(&nox_xxx_XFerMover_4F5730,
		mover_xfer_fn_4F5730: 1, default: 0),
	"MoverXfer must preserve two native pointers and its exact int32 result");

static nox_object_t* observed_object;
static void* observed_context;

int32_t nox_xxx_XFerMover_4F5730(
	nox_object_t* object,
	void* context) {
	observed_object = object;
	observed_context = context;
	return object != NULL;
}

int main(void) {
	nox_object_t object = {.marker = UINTPTR_MAX};
	void* const context = (void*)(uintptr_t)(UINTPTR_MAX - 71u);
	mover_xfer_fn_4F5730 const transfer = nox_xxx_XFerMover_4F5730;

	assert(transfer(&object, context) == 1);
	assert(observed_object == &object);
	assert(observed_object->marker == UINTPTR_MAX);
	assert(observed_context == context);

	assert(transfer(NULL, NULL) == 0);
	assert(observed_object == NULL);
	assert(observed_context == NULL);
	return 0;
}
