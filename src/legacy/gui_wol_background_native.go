package legacy

/*
#include "GAME1_2.h"

*/
import "C"

import "github.com/opennox/opennox/v1/client/gui"

//export sub_438E30
func sub_438E30(win *C.nox_window, _ *C.nox_window_data) C.int {
	var ux, uy C.uint
	C.nox_client_wndGetPosition_46AA60(win, &ux, &uy)
	x, y := int32(ux), int32(uy)

	draw := &win.draw_data
	image := draw.bg_image
	if draw.field_0&6 != 0 {
		image = draw.hl_image
	}
	nox_client_drawImageAt_47D2C0((*nox_video_bag_image_t)(image), x+int32(draw.img_px), y+int32(draw.img_py))

	for child := win.field_100; child != nil; child = child.prev {
		if child.flags&C.NOX_WIN_HIDDEN != 0 || gui.StyleFlags(child.draw_data.style) != gui.StyleStaticText {
			continue
		}
		text := (*C.nox_staticText_data)(child.widget_data)
		if text == nil {
			continue
		}
		nox_xxx_drawSetTextColor_434390(int32(renderPalette.white))
		nox_xxx_drawStringStyle_43F7B0(child.draw_data.font, text.text, x+int32(child.off_x), y+int32(child.off_y))
	}
	return 1
}
