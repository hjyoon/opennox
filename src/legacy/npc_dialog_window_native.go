package legacy

/*
#include "GAME2_2.h"
#include "GAME2_3.h"
#include "GAME3.h"
#include "client__gui__window.h"

*/
import "C"

import "unsafe"

var npcDialogWindow *nox_window

//export sub_479950
func sub_479950() C.int {
	if wndIsShown_nox_xxx_wndIsShown_46ACC0(npcDialogWindow) != 0 {
		return 0
	}
	npcDialogSendResponse(0)
	return 1
}

//export sub_4799A0
func sub_4799A0() C.int {
	*g2ServerWord(1107052) = uint32(nox_color_rgb_4344A0(240, 128, 64))
	root := nox_new_window_from_file(internCStr("Dialog.wnd"), C.nox_xxx_guiDialog_479B00)
	npcDialogWindow = root
	asWindow(root).AddCleanup(func() {
		if npcDialogWindow == root {
			npcDialogWindow, clientUIWords.dialog_active = nil, 0
		}
	})
	if root == nil {
		return 0
	}
	nox_xxx_wndSetWindowProc_46B300(root, C.sub_479BE0)
	nox_xxx_wndSetDrawFn_46B340(root, C.sub_479CB0)
	nox_gui_winSetFunc96_46B070(root, C.sub_479D00)

	slider := nox_xxx_wndGetChildByID_46B0C0(root, 3904)
	up := nox_xxx_wndGetChildByID_46B0C0(root, 3903)
	down := nox_xxx_wndGetChildByID_46B0C0(root, 3902)
	list := nox_xxx_wndGetChildByID_46B0C0(root, 3901)
	done := nox_xxx_wndGetChildByID_46B0C0(root, 3906)
	data := g2ServerList(list)
	if slider == nil || up == nil || down == nil || done == nil || data == nil {
		nox_xxx_windowDestroyMB_46C4E0(root)
		npcDialogWindow = nil
		return 0
	}
	lit := unsafe.Pointer(Nox_xxx_gLoadImg("UISliderLit").C())
	normal := unsafe.Pointer(Nox_xxx_gLoadImg("UISlider").C())
	C.sub_4B5700(slider, nil, nil, normal, lit, lit)
	slider.draw_data.win, up.draw_data.win, down.draw_data.win = list, list, list
	data.Field_9, data.Field_7, data.Field_8 = unsafe.Pointer(slider), unsafe.Pointer(up), unsafe.Pointer(down)
	nox_xxx_wndSetDrawFn_46B340(done, C.sub_479C40)
	nox_xxx_wnd_46ABB0_go(root, 0)
	nox_window_set_hidden(root, 1)
	clientUIWords.dialog_active = 0
	return 1
}

//export sub_479BE0
func sub_479BE0(win *nox_window, event C.int, arg, value C.uintptr_t) C.int {
	return 1
}

//export sub_479C40
func sub_479C40(win *nox_window, draw *C.nox_window_data) C.int {
	frame := byte(nox_xxx_bookGet_430B40_get_mouse_prev_seq())
	if sub_44D930() == 0 && frame&0x7f < 30 && frame&8 != 0 {
		window := asWindow(win)
		origin, size := window.GlobalPos(), window.Size()
		sub_49CD30(C.int(origin.X), C.int(origin.Y), C.int(size.X), C.int(size.Y)-2, C.int(*g2ServerWord(1107052)), 4)
	}
	return C.int(nox_xxx_wndButtonDrawNoImg_4A81D0(win, draw))
}

//export sub_479CB0
func sub_479CB0(win *nox_window, draw *C.nox_window_data) C.int {
	nox_client_drawImageAt_47D2C0((*nox_video_bag_image_t)(draw.bg_image), int32(clientDisplayStorage.width)-C.NOX_DEFAULT_WIDTH,
		int32(clientDisplayStorage.height)-C.NOX_DEFAULT_HEIGHT)
	return 1
}

//export sub_479D00
func sub_479D00() C.int {
	return 1
}

//export sub_479D10
func sub_479D10() C.int {
	nox_xxx_windowDestroyMB_46C4E0(npcDialogWindow)
	npcDialogWindow, clientUIWords.dialog_active = nil, 0
	return 0
}
