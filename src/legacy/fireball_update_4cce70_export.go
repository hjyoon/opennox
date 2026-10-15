package legacy

/*
#include "defs.h"
#include "GAME3_1.h"
*/
import "C"

// Keep the registered C names, native pointer arguments, and original power
// mapping. The client supplies its existing Go effect implementation. Like
// other drawable exports, a renderer-free server keeps the drawable alive.
func fireballUpdateNative4CCE70(vp *nox_draw_viewport_t, dr *nox_drawable, power int) int32 {
	if ClientUpdateFireball4CCE70 == nil {
		return 1
	}
	return int32(ClientUpdateFireball4CCE70(asViewport(vp), asDrawable(dr), power))
}

//export sub_4CCE70
func sub_4CCE70(vp *nox_draw_viewport_t, dr *nox_drawable) int32 {
	return fireballUpdateNative4CCE70(vp, dr, 5)
}

//export sub_4CD090
func sub_4CD090(vp *nox_draw_viewport_t, dr *nox_drawable) int32 {
	return fireballUpdateNative4CCE70(vp, dr, 4)
}

//export sub_4CD0C0
func sub_4CD0C0(vp *nox_draw_viewport_t, dr *nox_drawable) int32 {
	return fireballUpdateNative4CCE70(vp, dr, 3)
}

//export sub_4CD0F0
func sub_4CD0F0(vp *nox_draw_viewport_t, dr *nox_drawable) int32 {
	return fireballUpdateNative4CCE70(vp, dr, 2)
}

//export sub_4CD120
func sub_4CD120(vp *nox_draw_viewport_t, dr *nox_drawable) int32 {
	return fireballUpdateNative4CCE70(vp, dr, 1)
}
