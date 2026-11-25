package legacy

/*
#include "client__draw__glowdraw.h"

extern uint32_t dword_5d4594_1313532, dword_5d4594_1313536, dword_5d4594_1313540;
*/
import "C"

import (
	"unsafe"

	noxcolor "github.com/opennox/libs/color"
	"github.com/opennox/opennox/v1/client"
	"github.com/opennox/opennox/v1/client/noxrender"
	"github.com/opennox/opennox/v1/common/memmap"
)

var DrawLifetimeSparkle func(*client.Drawable, *noxrender.Viewport, noxcolor.RGBA5551, noxcolor.RGBA5551) int

//export sub_4B6770
func sub_4B6770(vp *C.int, dr *C.nox_drawable, bright, dim C.int) C.int {
	if DrawLifetimeSparkle == nil {
		return 0
	}
	return C.int(DrawLifetimeSparkle(asDrawable(dr), (*noxrender.Viewport)(unsafe.Pointer(vp)),
		noxcolor.RGBA5551(bright), noxcolor.RGBA5551(dim)))
}

//export nox_thing_red_spark_draw
func nox_thing_red_spark_draw(vp *C.uint32_t, dr *C.nox_drawable) C.int {
	return sub_4B6970(vp, dr, C.int(C.dword_5d4594_1313532), C.int(memmap.Int32(0x5D4594, 1313528)))
}

//export nox_thing_blue_spark_draw
func nox_thing_blue_spark_draw(vp *C.uint32_t, dr *C.nox_drawable) C.int {
	return sub_4B6970(vp, dr, C.int(C.dword_5d4594_1313540), C.int(C.dword_5d4594_1313536))
}

//export nox_thing_cyan_spark_draw
func nox_thing_cyan_spark_draw(vp *C.uint32_t, dr *C.nox_drawable) C.int {
	return sparkDrawFromPalette(vp, dr, 1313544)
}

//export nox_thing_green_spark_draw
func nox_thing_green_spark_draw(vp *C.uint32_t, dr *C.nox_drawable) C.int {
	return sparkDrawFromPalette(vp, dr, 1313580)
}

//export nox_thing_yellow_spark_draw
func nox_thing_yellow_spark_draw(vp *C.uint32_t, dr *C.nox_drawable) C.int {
	return sub_4B6970(vp, dr, C.int(C.dword_5d4594_1313532), C.int(memmap.Int32(0x5D4594, 1313576)))
}

//export nox_thing_violet_spark_draw
func nox_thing_violet_spark_draw(vp *C.uint32_t, dr *C.nox_drawable) C.int {
	return sparkDrawFromPalette(vp, dr, 1313552)
}

//export nox_thing_death_ball_spark_draw
func nox_thing_death_ball_spark_draw(vp *C.uint32_t, dr *C.nox_drawable) C.int {
	return sparkDrawFromPalette(vp, dr, 1313568)
}

//export nox_thing_white_spark_draw
func nox_thing_white_spark_draw(vp *C.uint32_t, dr *C.nox_drawable) C.int {
	return sub_4B6970(vp, dr, C.int(Get_nox_color_white_2523948()), C.int(C.dword_5d4594_1313540))
}

func sparkDrawFromPalette(vp *C.uint32_t, dr *C.nox_drawable, offset uintptr) C.int {
	bright := memmap.Int32(0x5D4594, offset+4)
	dim := memmap.Int32(0x5D4594, offset)
	return sub_4B6970(vp, dr, C.int(bright), C.int(dim))
}

//export nox_thing_particle_draw
func nox_thing_particle_draw(vp *C.nox_draw_viewport_t, dr *C.nox_drawable) C.int {
	return C.int(nox_thing_particle_draw_native(vp, dr))
}
