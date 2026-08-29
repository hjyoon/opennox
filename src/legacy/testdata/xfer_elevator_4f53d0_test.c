#include <assert.h>
#include <limits.h>
#include <stddef.h>
#include <stdint.h>

#include "../xfer_elevator_4f53d0.h"

struct nox_object_t {
	uintptr_t marker;
};

typedef int32_t (*elevator_xfer_fn_4F53D0)(nox_object_t*, void*);

_Static_assert(CHAR_BIT == 8, "serialized bytes must remain eight bits");
_Static_assert(sizeof(void*) == 4 || sizeof(void*) == 8,
	"unsupported pointer width");
_Static_assert(sizeof(int32_t) == 4,
	"ElevatorXfer result must remain int32");
_Static_assert(sizeof(nox_elevator_update_data_t) == 20,
	"Elevator update data must remain twenty bytes");
_Static_assert(offsetof(nox_elevator_update_data_t, field_0) == 0,
	"Elevator field-0 must remain at byte 0");
_Static_assert(offsetof(nox_elevator_update_data_t, link_pe32) == 4,
	"Elevator PE32 link slot must remain at byte 4");
_Static_assert(offsetof(nox_elevator_update_data_t, shaft_extent) == 8,
	"Elevator shaft extent must remain at byte 8");
_Static_assert(offsetof(nox_elevator_update_data_t, field_3) == 12,
	"Elevator field-3 must remain at byte 12");
_Static_assert(offsetof(nox_elevator_update_data_t, field_4) == 16,
	"Elevator field-4 must remain at byte 16");
_Static_assert(
	_Generic(&nox_xxx_XFerElevator_4F53D0,
		elevator_xfer_fn_4F53D0: 1, default: 0),
	"ElevatorXfer must preserve two native pointers and its exact int32 result");

static nox_object_t* observed_object;
static void* observed_context;

int32_t nox_xxx_XFerElevator_4F53D0(
	nox_object_t* object,
	void* context) {
	observed_object = object;
	observed_context = context;
	return object != NULL;
}

int main(void) {
	nox_object_t object = {.marker = UINTPTR_MAX};
	void* const context = (void*)(uintptr_t)(UINTPTR_MAX - 31u);
	elevator_xfer_fn_4F53D0 const transfer =
		nox_xxx_XFerElevator_4F53D0;

	assert(transfer(&object, context) == 1);
	assert(observed_object == &object);
	assert(observed_object->marker == UINTPTR_MAX);
	assert(observed_context == context);

	assert(transfer(NULL, NULL) == 0);
	assert(observed_object == NULL);
	assert(observed_context == NULL);
	return 0;
}
