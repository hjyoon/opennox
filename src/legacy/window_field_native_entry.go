package legacy

/*
#include "GAME2_1.h"
*/
import "C"

import "unsafe"

//export nox_xxx_wndSetOffsetMB_46AE40
func nox_xxx_wndSetOffsetMB_46AE40(win *C.nox_window, x, y C.int) *C.nox_window {
	return (*C.nox_window)(windowFieldOffset46AE40(asWindow(win), int32(x), int32(y)).C())
}

//export nox_xxx_wndSetIcon_46AE60
func nox_xxx_wndSetIcon_46AE60(win *C.nox_window, image *C.nox_video_bag_image_t) C.int {
	return C.int(windowFieldImage(asWindow(win), unsafe.Pointer(image), 1))
}

//export nox_xxx_wndSetIconLit_46AEA0
func nox_xxx_wndSetIconLit_46AEA0(win *C.nox_window, image *C.nox_video_bag_image_t) C.int {
	return C.int(windowFieldImage(asWindow(win), unsafe.Pointer(image), 2))
}

//export sub_46AEC0
func sub_46AEC0(win *C.nox_window, image *C.nox_video_bag_image_t) C.int {
	return C.int(windowFieldImage(asWindow(win), unsafe.Pointer(image), 3))
}

//export sub_46AF40
func sub_46AF40(win *C.nox_window) unsafe.Pointer {
	return windowFieldFont46AF40(asWindow(win))
}
