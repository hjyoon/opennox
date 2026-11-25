package legacy

/*
#include "GAME1_3.h"
#include "GAME2_1.h"
#include "client__gui__window.h"

extern nox_window* dword_5d4594_825744;
extern nox_window* dword_5d4594_1193712;
extern int nox_win_width, nox_win_height;
int sub_48C9F0(nox_window* win, void* draw);
*/
import "C"

import (
	"unsafe"

	"github.com/opennox/opennox/v1/common/memmap"
)

func clientStatusIcon(image string, imageSlot uintptr, aboveCenter C.int) *nox_window {
	slot := memmap.PtrPtr(0x5D4594, imageSlot)
	*slot = unsafe.Pointer(nox_xxx_gLoadImg_42F970(internCStr(image)))
	window := nox_window_new_go(nil, 136, C.nox_win_width-50, C.nox_win_height/2-aboveCenter, 50, 50, nil)
	nox_xxx_wndSetIcon_46AE60(window, (*nox_video_bag_image_t)(*slot))
	return window
}

//export nox_xxx_guiChatIconLoad_445650
func nox_xxx_guiChatIconLoad_445650() C.int {
	window := clientStatusIcon("ChatIcon", 825748, 50)
	C.dword_5d4594_825744 = window
	if window == nil {
		return 0
	}
	nox_window_set_all_funcs_go(window, nil, unsafe.Pointer(C.nox_xxx_guiChatMode_4456E0), nil)
	const source = `C:\\NoxPost\\src\\client\\Gui\\chaticon.c`
	tooltip := nox_strman_loadString_40F1D0(internCStr("chatmode"), nil, internCStr(source), 73)
	nox_xxx_wndWddSetTooltip_46B000(&window.draw_data, tooltip)
	return 1
}

//export sub_48C980
func sub_48C980() C.int {
	window := clientStatusIcon("ObserverIcon", 1193716, 100)
	C.dword_5d4594_1193712 = window
	if window == nil {
		return 0
	}
	nox_window_set_all_funcs_go(window, nil, unsafe.Pointer(C.sub_48C9F0), nil)
	return 1
}

//export sub_48C9F0
func sub_48C9F0(window *C.nox_window, draw unsafe.Pointer) C.int {
	var x, y C.uint
	nox_client_wndGetPosition_46AA60(window, &x, &y)
	if window != nil && Get_dword_8531A0_2576() != nil {
		const source = `C:\\NoxPost\\src\\client\\Gui\\guiobs.c`
		tooltip := nox_strman_loadString_40F1D0(internCStr("observermode"), nil, internCStr(source), 41)
		if observer := C.dword_5d4594_1193712; observer != nil {
			nox_xxx_wndWddSetTooltip_46B000(&observer.draw_data, tooltip)
		}
		data := &window.draw_data
		nox_client_drawImageAt_47D2C0((*nox_video_bag_image_t)(data.bg_image),
			int32(x)+int32(data.img_px), int32(y)+int32(data.img_py))
	}
	return 1
}
