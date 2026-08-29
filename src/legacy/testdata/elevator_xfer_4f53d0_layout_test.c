// Compile-only native-width contract for the Elevator transfer restored from
// GAME.EXE 004F53D0.
#define _Static_assert(...) extern int nox_suppressed_static_assert
#include "../GAME3_3.h"
#undef _Static_assert

#include <stddef.h>
#include <stdint.h>

_Static_assert(sizeof(nox_elevator_update_data_t) == 20,
	"Elevator update-data size");
_Static_assert(offsetof(nox_elevator_update_data_t, field_0) == 0,
	"Elevator field-0 offset");
_Static_assert(offsetof(nox_elevator_update_data_t, link_pe32) == 4,
	"Elevator PE32 link offset");
_Static_assert(offsetof(nox_elevator_update_data_t, shaft_extent) == 8,
	"Elevator shaft-extent offset");
_Static_assert(offsetof(nox_elevator_update_data_t, field_3) == 12,
	"Elevator field-3 offset");
_Static_assert(offsetof(nox_elevator_update_data_t, field_4) == 16,
	"Elevator field-4 offset");
_Static_assert(offsetof(nox_object_t, field_34) == (sizeof(void*) == 4 ? 136 : 140),
	"object field-34 offset");
_Static_assert(offsetof(nox_object_t, data_update) == (sizeof(void*) == 4 ? 748 : 872),
	"object update-data offset");

typedef int32_t (*elevator_xfer_fn_4F53D0)(nox_object_t*, void*);

_Static_assert(
	_Generic(&nox_xxx_XFerElevator_4F53D0,
		elevator_xfer_fn_4F53D0: 1, default: 0),
	"ElevatorXfer signature");

static elevator_xfer_fn_4F53D0 const elevator_xfer_signature =
	nox_xxx_XFerElevator_4F53D0;

int main(void) {
	return elevator_xfer_signature == NULL;
}
