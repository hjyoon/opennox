package legacy

/*
#include "GAME3_1.h"
#include "client__gui__window.h"
extern int nox_win_width, nox_win_height;
extern uint32_t dword_5d4594_1319056;
extern nox_window* dword_5d4594_1319060;
*/
import "C"

import "unsafe"

//export sub_4BFB70
func sub_4BFB70(mode C.int) {
	if C.dword_5d4594_1319056 == 0 && mode == 1 {
		nox_xxx_clientPlaySoundSpecial_452D80(1022, 100)
	}
	C.dword_5d4594_1319056 = C.uint32_t(mode)
}

//export sub_4BFBB0
func sub_4BFBB0(request *C.uint32_t) {
	switch C.dword_5d4594_1319056 {
	case 0:
		if uintptr(unsafe.Pointer(request)) == 1 {
			sub_4BFBF0()
			sub_4BFB70(1)
		}
	case 1:
		if request == nil {
			sub_4BFC70()
			sub_4BFB70(0)
		}
	}
}

//export sub_4BFBF0
func sub_4BFBF0() C.int {
	root := C.dword_5d4594_1319060
	if root == nil {
		return 0
	}
	g2ServerHidden(root, false)
	g2ServerEnabled(root, true)
	var width, height C.int
	C.nox_window_get_size(root, &width, &height)
	nox_window_setPos_46A9B0(root, int32(C.nox_win_width/2-width/2), int32(C.nox_win_height/2-height/2))
	return C.int(nox_xxx_windowFocus_46B500(nil))
}

//export sub_4BFC70
func sub_4BFC70() C.int {
	root := C.dword_5d4594_1319060
	if root == nil {
		return 0
	}
	g2ServerHidden(root, true)
	return C.int(nox_xxx_windowFocus_46B500(nil))
}

//export sub_4BFC90
func sub_4BFC90() C.int {
	root := nox_new_window_from_file(internCStr("SKey.wnd"), C.sub_4BFCD0)
	C.dword_5d4594_1319060 = root
	if root == nil {
		return 0
	}
	sub_4BFB70(0)
	sub_4BFC70()
	return 1
}

//export sub_4BFCD0
func sub_4BFCD0(win *nox_window, event C.int, control *nox_window, arg C.uintptr_t) C.int {
	if event == 16391 {
		id := C.nox_xxx_wndGetID_46B0A0(control)
		nox_xxx_clientPlaySoundSpecial_452D80(766, 100)
		if id == 10803 {
			sub_4BFC70()
		}
	}
	return 0
}

//export sub_4BFD10
func sub_4BFD10() {
	nox_xxx_windowDestroyMB_46C4E0(C.dword_5d4594_1319060)
	C.dword_5d4594_1319060 = nil
	sub_4BFB70(0)
}
