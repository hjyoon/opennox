#include <assert.h>
#include <limits.h>
#include <stddef.h>
#include <stdint.h>

#include "../xfer_elevator_shaft_4f54a0.h"

struct nox_object_t {
	uintptr_t marker;
};

typedef int32_t (*elevator_shaft_xfer_fn_4F54A0)(nox_object_t*, void*);

_Static_assert(CHAR_BIT == 8, "serialized bytes must remain eight bits");
_Static_assert(sizeof(void*) == 4 || sizeof(void*) == 8,
	"unsupported pointer width");
_Static_assert(sizeof(int32_t) == 4,
	"ElevatorShaftXfer result must remain int32");
_Static_assert(sizeof(nox_elevator_shaft_update_data_t) == 16,
	"ElevatorShaft update data must remain sixteen bytes");
_Static_assert(offsetof(nox_elevator_shaft_update_data_t, field_0) == 0,
	"ElevatorShaft field-0 must remain at byte 0");
_Static_assert(offsetof(nox_elevator_shaft_update_data_t, link_pe32) == 4,
	"ElevatorShaft PE32 link slot must remain at byte 4");
_Static_assert(offsetof(nox_elevator_shaft_update_data_t, elevator_extent) == 8,
	"ElevatorShaft elevator extent must remain at byte 8");
_Static_assert(offsetof(nox_elevator_shaft_update_data_t, field_3) == 12,
	"ElevatorShaft field-3 must remain at byte 12");
_Static_assert(
	_Generic(&nox_xxx_XFerElevatorShaft_4F54A0,
		elevator_shaft_xfer_fn_4F54A0: 1, default: 0),
	"ElevatorShaftXfer must preserve two native pointers and its exact int32 result");

static nox_object_t* observed_object;
static void* observed_context;

int32_t nox_xxx_XFerElevatorShaft_4F54A0(
	nox_object_t* object,
	void* context) {
	observed_object = object;
	observed_context = context;
	return object != NULL;
}

int main(void) {
	nox_object_t object = {.marker = UINTPTR_MAX};
	void* const context = (void*)(uintptr_t)(UINTPTR_MAX - 63u);
	elevator_shaft_xfer_fn_4F54A0 const transfer =
		nox_xxx_XFerElevatorShaft_4F54A0;

	assert(transfer(&object, context) == 1);
	assert(observed_object == &object);
	assert(observed_object->marker == UINTPTR_MAX);
	assert(observed_context == context);

	assert(transfer(NULL, NULL) == 0);
	assert(observed_object == NULL);
	assert(observed_context == NULL);
	return 0;
}
