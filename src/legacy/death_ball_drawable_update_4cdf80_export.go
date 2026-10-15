package legacy

/*
#include "defs.h"
#include "GAME3_1.h"
#include "GAME3_2.h"
#include "client__drawable__update__dball.h"
*/
import "C"

// Retain the registered C names and spark counts without narrowing native
// viewport or drawable pointers. A renderer-free server keeps the drawable
// alive, matching the other drawable update exports.
func deathBallSparksNative4CDF80(vp *nox_draw_viewport_t, dr *nox_drawable, count int) int32 {
	if ClientUpdateDeathBallSparks4CDF80 == nil {
		return 1
	}
	return int32(ClientUpdateDeathBallSparks4CDF80(asViewport(vp), asDrawable(dr), count))
}

//export nox_xxx_updDrawDBall_4CDF80
func nox_xxx_updDrawDBall_4CDF80(vp *nox_draw_viewport_t, dr *nox_drawable) int32 {
	return deathBallSparksNative4CDF80(vp, dr, 3)
}

//export sub_4CE0A0
func sub_4CE0A0(vp *nox_draw_viewport_t, dr *nox_drawable) int32 {
	return deathBallSparksNative4CDF80(vp, dr, 1)
}

//export nox_xxx_updDrawDBallCharge_4CE0C0
func nox_xxx_updDrawDBallCharge_4CE0C0(vp *nox_draw_viewport_t, dr *nox_drawable) int32 {
	if ClientUpdateDeathBallCharge4CE0C0 == nil {
		return 1
	}
	return int32(ClientUpdateDeathBallCharge4CE0C0(asViewport(vp), asDrawable(dr)))
}

//export sub_4CA650
func sub_4CA650(vp *nox_draw_viewport_t, dr *nox_drawable) int32 {
	if ClientUpdateLinearOrb4CA650 == nil {
		return 1
	}
	return int32(ClientUpdateLinearOrb4CA650(asViewport(vp), asDrawable(dr)))
}
