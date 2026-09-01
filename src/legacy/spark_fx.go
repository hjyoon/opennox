package legacy

/*
#include "GAME5.h"

static void nox_createSpark_go(float x, float y, int kind, int lifetime,
	float velocity_x, float velocity_y, float z, int owner) {
	(void)nox_xxx_createSpark_54FD80(x, y, kind, lifetime,
		velocity_x, velocity_y, z, owner);
}
*/
import "C"

// Nox_xxx_createSpark_54FD80 retains the legacy particle allocator behind a
// primitive-only bridge. No Go or native object pointer crosses this call.
func Nox_xxx_createSpark_54FD80(
	x, y float32,
	kind, lifetime int,
	velocityX, velocityY, z float32,
	owner int,
) {
	C.nox_createSpark_go(
		C.float(x), C.float(y), C.int(kind), C.int(lifetime),
		C.float(velocityX), C.float(velocityY), C.float(z), C.int(owner),
	)
}
