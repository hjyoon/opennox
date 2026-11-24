package legacy

/*
#include "GAME2_3.h"

*/
import "C"

import (
	"unsafe"

	"github.com/opennox/opennox/v1/client/noxrender"
	noxflags "github.com/opennox/opennox/v1/common/flags"
)

//export sub_48DCF0
func sub_48DCF0(view *nox_draw_viewport_t) {
	vp := asViewport(view)
	render := GetClient().R2()
	fontHeight := C.int32_t(render.FontHeight(nil))
	order := C.int32_t(0)
	for bubble := chatBubbleHead; bubble != nil; {
		next := bubble.next
		bubble.order = order
		order++
		setChatBubblePosition(bubble, vp, render, fontHeight)
		if gameFrame() > uint32(bubble.expire_frame) {
			nox_chat_bubble_unlink_48DCF0(bubble)
			nox_alloc_class_free_obj_first(chatBubblePool(), unsafe.Pointer(bubble))
		}
		bubble = next
	}
	for bubble := chatBubbleHead; bubble != nil; bubble = bubble.next {
		sub_48E240(0, (*C.uint32_t)(unsafe.Pointer(bubble)))
	}
}

func setChatBubblePosition(bubble *C.nox_chat_bubble, vp *noxrender.Viewport, render Render2, fontHeight C.int32_t) {
	wire := C.uint(bubble.net_code)
	code := int32(nox_xxx_netClearHighBit_578B30(wire))
	if nox_xxx_netTestHighBit_578B70(wire) != 0 {
		bubble.drawable = nox_xxx_netSpriteByCodeStatic_45A720(code)
	} else {
		bubble.drawable = nox_xxx_netSpriteByCodeDynamic_45A6F0(code)
	}
	if dr := asDrawable(bubble.drawable); dr != nil {
		bubble.drawable_x, bubble.drawable_y = C.uint16_t(dr.PosVec.X), C.uint16_t(dr.PosVec.Y)
	}
	x := C.int32_t(vp.Screen.Min.X + int(bubble.drawable_x) - vp.World.Min.X)
	y := C.int32_t(vp.Screen.Min.Y + int(bubble.drawable_y) - vp.World.Min.Y)
	size := render.GetStringSizeWrapped(nil, GoWString(&bubble.text[0]), 128)
	bubble.width, bubble.height = min(C.int32_t(size.X), 128), C.int32_t(size.Y)
	x -= bubble.width / 2
	y -= 64 + bubble.height
	bubble.x, bubble.y, bubble.visible, bubble.draw_tail = x, y, 1, 1

	if noxflags.HasGame(noxflags.GameModeCoop) &&
		(int(x) < vp.Screen.Min.X || int(x) > vp.Screen.Max.X ||
			int(y-64) < vp.Screen.Min.Y || int(y-64) > vp.Screen.Max.Y) {
		bubble.visible, bubble.draw_tail = 0, 0
		return
	}

	if left := C.int32_t(vp.Screen.Min.X) + fontHeight; x < left {
		x, bubble.draw_tail = left, 0
	} else if right := C.int32_t(vp.Screen.Max.X); x+bubble.width+fontHeight > right {
		x, bubble.draw_tail = right-bubble.width-fontHeight, 0
	}
	if top := C.int32_t(vp.Screen.Min.Y) + 2*fontHeight + 2; y < top {
		y, bubble.draw_tail = top, 0
	} else if bottom := C.int32_t(vp.Screen.Max.Y); y+bubble.height+fontHeight > bottom {
		y, bubble.draw_tail = bottom-bubble.height-fontHeight, 0
	}
	box := C.int4{field_0: C.int(x), field_4: C.int(y), field_8: C.int(x + bubble.width), field_C: C.int(y + bubble.height)}
	sub_48E000(&box, &bubble.draw_tail)
	bubble.x, bubble.y = C.int32_t(box.field_0), C.int32_t(box.field_4)
}
