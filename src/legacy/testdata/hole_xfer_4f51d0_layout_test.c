// Compile-only native-width contract for the Hole transfer restored from
// GAME.EXE 004F51D0.
#define _Static_assert(...)
#include "../GAME3_3.h"
#undef _Static_assert

#include <stddef.h>
#include <stdint.h>

_Static_assert(sizeof(nox_hole_collide_data_t) == 28, "Hole collide-data size");
_Static_assert(offsetof(nox_hole_collide_data_t, script) == 0,
	"Hole script offset");
_Static_assert(offsetof(nox_hole_collide_data_t, destination_x) == 8,
	"Hole destination-X offset");
_Static_assert(offsetof(nox_hole_collide_data_t, destination_y) == 12,
	"Hole destination-Y offset");
_Static_assert(offsetof(nox_hole_collide_data_t, destination_extent) == 16,
	"Hole destination-extent offset");
_Static_assert(offsetof(nox_hole_collide_data_t, destination_net_code) == 20,
	"Hole destination-net-code offset");
_Static_assert(offsetof(nox_hole_collide_data_t, reserved_22) == 22,
	"Hole reserved-22 offset");
_Static_assert(offsetof(nox_hole_collide_data_t, field_24) == 24,
	"Hole field-24 offset");
_Static_assert(offsetof(nox_object_t, field_34) == (sizeof(void*) == 4 ? 136 : 140),
	"object field-34 offset");
_Static_assert(offsetof(nox_object_t, collide_data) == (sizeof(void*) == 4 ? 700 : 776),
	"object collide-data offset");
_Static_assert(offsetof(nox_object_t, field_189) == (sizeof(void*) == 4 ? 756 : 888),
	"object field-189 offset");

typedef int32_t (*hole_xfer_fn_4F51D0)(nox_object_t*, void*);

_Static_assert(
	_Generic(&nox_xxx_XFerHole_4F51D0,
		hole_xfer_fn_4F51D0: 1, default: 0),
	"HoleXfer signature");

static hole_xfer_fn_4F51D0 const hole_xfer_signature =
	nox_xxx_XFerHole_4F51D0;

int main(void) {
	return hole_xfer_signature == NULL;
}
