package legacy

/*
#include "defs.h"
#include "GAME3_1.h"
#include "client__drawable__update__healup.h"
#include "client__drawable__update__drainup.h"
*/
import "C"

import (
	"github.com/opennox/opennox/v1/client"
	"github.com/opennox/opennox/v1/client/noxrender"
)

var (
	ClientUpdateCharm4CD400     func(*noxrender.Viewport, *client.Drawable) int
	ClientUpdateHeal4CD450      func(*noxrender.Viewport, *client.Drawable) int
	ClientUpdateDrainMana4CD690 func(*noxrender.Viewport, *client.Drawable) int
)

// The drawable-update table retains these C symbols. Forward their native
// viewport and drawable pointers to the same implementations as world updates.
// A dedicated server has no renderer and therefore keeps the drawable alive.
//
//export sub_4CD400
func sub_4CD400(vp *nox_draw_viewport_t, dr *nox_drawable) int32 {
	if ClientUpdateCharm4CD400 == nil {
		return 1
	}
	return int32(ClientUpdateCharm4CD400(asViewport(vp), asDrawable(dr)))
}

//export sub_4CD450
func sub_4CD450(vp *nox_draw_viewport_t, dr *nox_drawable) int32 {
	if ClientUpdateHeal4CD450 == nil {
		return 1
	}
	return int32(ClientUpdateHeal4CD450(asViewport(vp), asDrawable(dr)))
}

//export sub_4CD690
func sub_4CD690(vp *nox_draw_viewport_t, dr *nox_drawable) int32 {
	if ClientUpdateDrainMana4CD690 == nil {
		return 1
	}
	return int32(ClientUpdateDrainMana4CD690(asViewport(vp), asDrawable(dr)))
}
