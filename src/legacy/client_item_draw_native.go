package legacy

/*
#include "client__draw__animdraw.h"
#include "client__draw__armordraw.h"
#include "client__draw__basedraw.h"
#include "client__draw__flagdraw.h"
#include "client__draw__magicdrw.h"
#include "client__draw__parse__parse.h"
#include "client__draw__slavedraw.h"
#include "client__draw__staticdraw.h"
#include "client__draw__weapondraw.h"
*/
import "C"

import (
	"unsafe"

	noxflags "github.com/opennox/opennox/v1/common/flags"
)

//export nox_thing_base_draw
func nox_thing_base_draw(viewport *C.int, drawable *C.nox_drawable) C.int {
	C.nox_thing_weapon_draw(viewport, drawable)
	return 1
}

//export nox_thing_armor_draw
func nox_thing_armor_draw(viewport *C.int, drawable *C.nox_drawable) C.int {
	sub_4B96F0(drawable)
	return C.nox_thing_static_draw((*C.uint32_t)(unsafe.Pointer(viewport)), drawable)
}

//export nox_thing_armor_animate_draw
func nox_thing_armor_animate_draw(viewport *C.int, drawable *C.nox_drawable) C.int {
	sub_4B96F0(drawable)
	return C.nox_thing_animate_draw((*C.uint32_t)(unsafe.Pointer(viewport)), drawable)
}

//export nox_thing_flag_draw
func nox_thing_flag_draw(viewport *C.int, drawable *C.nox_drawable) C.int {
	C.nox_thing_weapon_animate_draw(viewport, drawable)
	if noxflags.HasGame(noxflags.GameModeChat) && drawable.flags30&0x1000000 != 0 {
		view := (*C.nox_draw_viewport_t)(unsafe.Pointer(viewport))
		x := int32(view.x1) + int32(drawable.pos.x) - int32(view.field_4)
		y := int32(view.y1) + int32(drawable.pos.y) - int32(int16(drawable.z)) -
			int32(nox_float2int(drawable.field_25)) - int32(view.field_5)
		if team := sub_418A80(sub_4B94E0(drawable)); team != nil {
			var width C.int
			nox_xxx_drawSetTextColor_434390(int32(Get_dword_8531A0_2572()))
			nox_xxx_drawGetStringSize_43F840(nil, &team.name[0], &width, nil, 0)
			nox_xxx_drawString_43F6E0(nil, &team.name[0], x-int32(width)/2, y)
		}
	}
	return 1
}

//export nox_thing_slave_draw
func nox_thing_slave_draw(viewport *C.uint32_t, drawable *C.nox_drawable) C.int {
	data := (*C.nox_static_random_draw_data_t)(drawable.field_76)
	if data != nil && data.images != nil && drawable.field_77 < C.uint32_t(data.count) {
		image := unsafe.Slice(data.images, int(data.count))[drawable.field_77]
		view := (*C.nox_draw_viewport_t)(unsafe.Pointer(viewport))
		nox_xxx_drawObject_4C4770_draw(view, drawable, unsafe.Pointer(image))
	}
	return 1
}

//export nox_things_slave_draw_parse
func nox_things_slave_draw_parse(thing *C.nox_thing, file *C.nox_memfile, attrs *C.char) C.bool {
	thing.draw_func = (*[0]byte)(C.nox_thing_slave_draw)
	data := C.nox_xxx_spriteLoadStaticRandomData_44C000(attrs, file)
	thing.field_5c = unsafe.Pointer(data)
	thing.field_60 = 0
	return C.bool(data != nil)
}

//export nox_thing_drain_mana_draw
func nox_thing_drain_mana_draw() C.int { return 1 }
