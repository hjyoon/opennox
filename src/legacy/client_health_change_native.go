package legacy

/*
#include "GAME2_3.h"
#include "common/alloc/classes/alloc_class.h"

*/
import "C"

import (
	"image"
	"strconv"
	"unsafe"

	noxcolor "github.com/opennox/libs/color"

	"github.com/opennox/opennox/v1/common/memmap"
)

//export nox_xxx_allocArrayHealthChanges_49A5F0
func nox_xxx_allocArrayHealthChanges_49A5F0() C.int {
	pool := nox_new_alloc_class(internCStr("HealthChange"), C.sizeof_nox_health_change, 32)
	clientDisplayStorage.health_change_allocator = unsafe.Pointer(pool)
	if pool == nil {
		return 0
	}
	clientDisplayStorage.health_change_font = nox_xxx_guiFontPtrByName_43F360(internCStr("numbers"))
	return 1
}

//export sub_49A630
func sub_49A630() {
	nox_alloc_class_free_all((*C.nox_alloc_class)(clientDisplayStorage.health_change_allocator))
	clientDisplayStorage.health_change_head = nil
}

//export nox_xxx_cliAddHealthChange_49A650
func nox_xxx_cliAddHealthChange_49A650(code C.int, delta C.short) *C.nox_health_change {
	pool := (*C.nox_alloc_class)(clientDisplayStorage.health_change_allocator)
	change := (*C.nox_health_change)(nox_alloc_class_new_obj_zero(pool))
	if change == nil {
		return nil
	}
	change.drawable_id, change.delta = C.uint32_t(code), C.int16_t(delta)
	change.frame = C.uint32_t(nox_xxx_bookGet_430B40_get_mouse_prev_seq())
	change.next, change.prev = healthChangeHead(), nil
	if change.next != nil {
		change.next.prev = change
	}
	clientDisplayStorage.health_change_head = unsafe.Pointer(change)
	return change.next
}

func healthChangeHead() *C.nox_health_change {
	return (*C.nox_health_change)(clientDisplayStorage.health_change_head)
}

//export sub_49A6A0
func sub_49A6A0(view *nox_draw_viewport_t, drawable *nox_drawable) {
	frame := uint32(nox_xxx_bookGet_430B40_get_mouse_prev_seq())
	damageColor := uint32(renderPalette.yellow)
	if unsafe.Pointer(drawable) == *memmap.PtrPtr(0x852978, 8) {
		damageColor = memmap.Uint32(0x85B3FC, 940)
	}
	for change := healthChangeHead(); change != nil; {
		next := change.next
		age := frame - uint32(change.frame)
		if age > 30 {
			sub_49A880(change)
		} else if change.drawable_id == drawable.field_32 {
			dr := asDrawable(drawable)
			point := asViewport(view).ToScreenPos(dr.Pos())
			point.Y -= int(2*age) + int(int16(dr.ZVal)) + int(int32(dr.ZSizeMax))
			point = image.Pt(int(int32(point.X)), int(int32(point.Y)))
			value := int(change.delta)
			color := damageColor
			if value > 0 {
				color = uint32(worldWordStorage.local_player_color)
			}
			if value < 0 {
				value = -value
			}
			drawHealthChangeNumber(point, strconv.Itoa(value), color)
		}
		change = next
	}
}

var healthChangeOutline = [...]image.Point{{X: -1, Y: -1}, {X: -1, Y: 1}, {X: 1, Y: -1}, {X: 1, Y: 1}}

func drawHealthChangeNumber(point image.Point, text string, color uint32) {
	r := GetClient().R2()
	font := r.GetFonts().AsFont(clientDisplayStorage.health_change_font)
	point.X -= r.GetStringSizeWrapped(font, text, 0).X / 2
	r.Data().SetTextColor(noxcolor.RGBA5551(renderPalette.black))
	for _, offset := range healthChangeOutline {
		r.DrawString(font, text, point.Add(offset))
	}
	r.Data().SetTextColor(noxcolor.RGBA5551(color))
	r.DrawString(font, text, point)
}

//export sub_49A880
func sub_49A880(change *C.nox_health_change) {
	if change.prev != nil {
		change.prev.next = change.next
	} else {
		clientDisplayStorage.health_change_head = unsafe.Pointer(change.next)
	}
	if change.next != nil {
		change.next.prev = change.prev
	}
	nox_alloc_class_free_obj_first((*C.nox_alloc_class)(clientDisplayStorage.health_change_allocator), unsafe.Pointer(change))
}

//export sub_49A8C0
func sub_49A8C0() C.int {
	nox_free_alloc_class((*C.nox_alloc_class)(clientDisplayStorage.health_change_allocator))
	clientDisplayStorage.health_change_allocator, clientDisplayStorage.health_change_head, clientDisplayStorage.health_change_font = nil, nil, nil
	return 0
}
