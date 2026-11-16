package legacy

/*
#include "GAME2_1.h"
*/
import "C"

import (
	"unsafe"

	"github.com/opennox/opennox/v1/client"
)

func drawablePredicateStatic(dr *client.Drawable) bool {
	return dr != nil && dr.DrawFuncPtr != nil && (dr.DrawFuncPtr == Get_nox_thing_static_draw() || dr.DrawFuncPtr == Get_nox_thing_static_random_draw())
}

func drawablePredicateResult(value bool) C.int {
	if value {
		return 1
	}
	return 0
}

//export nox_xxx_sprite_4756E0_drawable
func nox_xxx_sprite_4756E0_drawable(owner *C.nox_drawable) C.int {
	dr := (*client.Drawable)(unsafe.Pointer(owner))
	return drawablePredicateResult(client.DrawableBackWallQueue(dr, drawablePredicateStatic(dr)))
}

//export nox_xxx_sprite_475740_drawable
func nox_xxx_sprite_475740_drawable(owner *C.nox_drawable) C.int {
	dr := (*client.Drawable)(unsafe.Pointer(owner))
	return drawablePredicateResult(client.DrawableFrontQueue(dr, drawablePredicateStatic(dr)))
}

//export nox_xxx_sprite_4757A0_drawable
func nox_xxx_sprite_4757A0_drawable(owner *C.nox_drawable) C.int {
	return drawablePredicateResult(client.DrawableOverlayQueue((*client.Drawable)(unsafe.Pointer(owner))))
}

//export sub_4757D0_drawable
func sub_4757D0_drawable(owner *C.nox_drawable) C.int {
	return drawablePredicateResult(client.DrawableWorldQueue((*client.Drawable)(unsafe.Pointer(owner))))
}
