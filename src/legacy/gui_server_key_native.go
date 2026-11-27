package legacy

/*
#include "GAME3_1.h"
#include "client__gui__window.h"
*/
import "C"

import "unsafe"

var serverKeyWindow *nox_window
var serverKeyMode uint32

//export sub_4BFB70
func sub_4BFB70(mode C.int) {
	if serverKeyMode == 0 && mode == 1 {
		nox_xxx_clientPlaySoundSpecial_452D80(1022, 100)
	}
	serverKeyMode = uint32(mode)
}

//export sub_4BFBB0
func sub_4BFBB0(request *C.uint32_t) {
	serverKeyVisibleRequest(uintptr(unsafe.Pointer(request)))
}

func serverKeyVisibleRequest(request uintptr) {
	switch serverKeyMode {
	case 0:
		if request == 1 {
			sub_4BFBF0()
			sub_4BFB70(1)
		}
	case 1:
		if request == 0 {
			sub_4BFC70()
			sub_4BFB70(0)
		}
	}
}

//export sub_4BFBF0
func sub_4BFBF0() C.int {
	root := serverKeyWindow
	if root == nil {
		return 0
	}
	g2ServerHidden(root, false)
	g2ServerEnabled(root, true)
	var width, height C.int
	nox_window_get_size(root, &width, &height)
	nox_window_setPos_46A9B0(root, int32(clientDisplayStorage.width/2-width/2), int32(clientDisplayStorage.height/2-height/2))
	return C.int(nox_xxx_windowFocus_46B500(nil))
}

//export sub_4BFC70
func sub_4BFC70() C.int {
	root := serverKeyWindow
	if root == nil {
		return 0
	}
	g2ServerHidden(root, true)
	return C.int(nox_xxx_windowFocus_46B500(nil))
}

//export sub_4BFC90
func sub_4BFC90() C.int {
	root := nox_new_window_from_file(internCStr("SKey.wnd"), C.sub_4BFCD0)
	serverKeyWindow = root
	if root == nil {
		return 0
	}
	asWindow(root).AddCleanup(func() {
		if serverKeyWindow == root {
			serverKeyWindow = nil
			serverKeyMode = 0
		}
	})
	sub_4BFB70(0)
	sub_4BFC70()
	return 1
}

//export sub_4BFCD0
func sub_4BFCD0(win *nox_window, event C.int, control *nox_window, arg C.uintptr_t) C.int {
	if event == 16391 {
		id := C.int(nox_xxx_wndGetID_46B0A0(control))
		nox_xxx_clientPlaySoundSpecial_452D80(766, 100)
		if id == 10803 {
			sub_4BFC70()
		}
	}
	return 0
}

//export sub_4BFD10
func sub_4BFD10() {
	nox_xxx_windowDestroyMB_46C4E0(serverKeyWindow)
	serverKeyWindow = nil
	sub_4BFB70(0)
}
