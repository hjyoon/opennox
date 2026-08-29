// Compile-only native-width contract for the ElevatorShaft transfer restored
// from GAME.EXE 004F54A0.
#define _Static_assert(...) extern int nox_suppressed_static_assert
#include "../GAME3_3.h"
#undef _Static_assert

#include <stddef.h>
#include <stdint.h>

_Static_assert(sizeof(nox_elevator_shaft_update_data_t) == 16,
	"ElevatorShaft update-data size");
_Static_assert(offsetof(nox_elevator_shaft_update_data_t, field_0) == 0,
	"ElevatorShaft field-0 offset");
_Static_assert(offsetof(nox_elevator_shaft_update_data_t, link_pe32) == 4,
	"ElevatorShaft PE32 link offset");
_Static_assert(offsetof(nox_elevator_shaft_update_data_t, elevator_extent) == 8,
	"ElevatorShaft elevator-extent offset");
_Static_assert(offsetof(nox_elevator_shaft_update_data_t, field_3) == 12,
	"ElevatorShaft field-3 offset");
_Static_assert(offsetof(nox_object_t, field_34) == (sizeof(void*) == 4 ? 136 : 140),
	"object field-34 offset");
_Static_assert(offsetof(nox_object_t, data_update) == (sizeof(void*) == 4 ? 748 : 872),
	"object update-data offset");

typedef int32_t (*elevator_shaft_xfer_fn_4F54A0)(nox_object_t*, void*);

_Static_assert(
	_Generic(&nox_xxx_XFerElevatorShaft_4F54A0,
		elevator_shaft_xfer_fn_4F54A0: 1, default: 0),
	"ElevatorShaftXfer signature");

static elevator_shaft_xfer_fn_4F54A0 const elevator_shaft_xfer_signature =
	nox_xxx_XFerElevatorShaft_4F54A0;

int main(void) {
	return elevator_shaft_xfer_signature == NULL;
}
