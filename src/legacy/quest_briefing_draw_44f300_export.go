package legacy

/*
#include "quest_briefing_draw_44f300.h"
#include "client__gui__guibrief.h"
#include "common__strman.h"
extern int nox_win_width;
extern int nox_win_height;
*/
import "C"

import (
	"image"
	"unsafe"

	"github.com/opennox/opennox/v1/client"
	"github.com/opennox/opennox/v1/client/gui"
	"github.com/opennox/opennox/v1/client/noxrender"
	"github.com/opennox/opennox/v1/common/memmap"
	"github.com/opennox/opennox/v1/legacy/common/ccall"
)

func questBriefingBlinkSlot44F300() *uint32 {
	return memmap.PtrUint32(0x587000, 123012)
}

func questBriefingDrawNativeHooks44F300(draw *gui.WindowData) questBriefingDrawHooks44F300[*noxrender.Viewport, *client.Drawable, unsafe.Pointer, *uint16] {
	slots := questPreviewSlots44E110()
	return questBriefingDrawHooks44F300[*noxrender.Viewport, *client.Drawable, unsafe.Pointer, *uint16]{
		viewport:       func() *noxrender.Viewport { return GetClient().Viewport() },
		initPreviews:   func() { questPreviewCall44E110() },
		resetParticles: Nox_client_resetScreenParticles_431510,
		hideBook:       func(v int32) int32 { return int32(Nox_xxx_bookHideMB_45ACA0(int(v))) },
		prepare:        Sub_446780,
		dimensions:     func() (int32, int32) { return int32(C.nox_win_width), int32(C.nox_win_height) },
		loadText: func(key, source string, line int32) *uint16 {
			return (*uint16)(unsafe.Pointer(C.nox_strman_loadString_40F1D0(internCStr(key), nil, internCStr(source), C.int(line))))
		},
		titleFont:  func() unsafe.Pointer { return draw.FontC() },
		normalFont: func() unsafe.Pointer { return *questPreviewFontSlot44E110() },
		measure: func(font unsafe.Pointer, text *uint16) int32 {
			var width C.int
			nox_xxx_drawGetStringSize_43F840(font, (*wchar2_t)(unsafe.Pointer(text)), &width, nil, 0)
			return int32(width)
		},
		color: func(c questBriefingColor44F300) uint32 {
			switch c {
			case questBriefingBlack44F300:
				return Get_nox_color_black_2650656()
			case questBriefingOrange44F300:
				return Get_nox_color_orange_2614256()
			default:
				return Get_nox_color_white_2523948()
			}
		},
		setColor: func(c uint32) { nox_xxx_drawSetTextColor_434390(int32(c)) },
		drawText: func(font unsafe.Pointer, text *uint16, x, y int32) int32 {
			return nox_xxx_drawString_43F6E0(font, (*wchar2_t)(unsafe.Pointer(text)), x, y)
		},
		wrapText: func(font unsafe.Pointer, text *uint16, x, y, width, height int32) int32 {
			return nox_xxx_drawStringWrap_43FAF0(font, (*wchar2_t)(unsafe.Pointer(text)), x, y, width, height)
		},
		loadDrawable: func(slot int) *client.Drawable { return *slots[slot] },
		position: func(vp *noxrender.Viewport, dr *client.Drawable, x, y int32) {
			// sub_473A10's coordinate rule, expressed in native fields with
			// the original dword arithmetic instead of writing at dr+12.
			dr.PosVec = image.Pt(int(x+int32(vp.World.Min.X)-int32(vp.Screen.Min.X)),
				int(y+int32(vp.World.Min.Y)-int32(vp.Screen.Min.Y)))
		},
		drawDrawable: func(vp *noxrender.Viewport, dr *client.Drawable) {
			// Keep the actual indirect call contract; CallDraw's additional
			// nil/callability guards are not part of this original body.
			ccall.CallIntPtr2(dr.DrawFuncPtr, vp.C(), dr.C())
		},
		frame:      gameFrame,
		loadBlink:  func() uint32 { return *questBriefingBlinkSlot44F300() },
		storeBlink: func(v uint32) { *questBriefingBlinkSlot44F300() = v },
	}
}

var questBriefingDrawCall44F300 = func(_ *gui.Window, draw *gui.WindowData) int32 {
	return questBriefingDraw44F300(questBriefingDrawNativeHooks44F300(draw))
}

//export nox_client_questBriefingDraw_native_44F300
func nox_client_questBriefingDraw_native_44F300(win *C.nox_window, draw *C.nox_window_data) C.int {
	return C.int(questBriefingDrawCall44F300((*gui.Window)(unsafe.Pointer(win)), (*gui.WindowData)(unsafe.Pointer(draw))))
}

func questBriefingDrawCEntry44F300(win *gui.Window, draw *gui.WindowData) int32 {
	return int32(C.sub_44F300((*C.nox_window)(unsafe.Pointer(win)), (*C.nox_window_data)(unsafe.Pointer(draw))))
}
