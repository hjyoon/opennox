package legacy

/*
#include "client__gui__window.h"
*/
import "C"

//export sub_46ACE0
func sub_46ACE0(root *C.nox_window, first, last, hidden C.int) {
	for id := first; id <= last; id++ {
		win := nox_xxx_wndGetChildByID_46B0C0(root, int32(id))
		nox_window_set_hidden(win, int32(hidden))
	}
}

//export sub_46AD20
func sub_46AD20(root *C.nox_window, first, last, enabled C.int) {
	for id := first; id <= last; id++ {
		win := nox_xxx_wndGetChildByID_46B0C0(root, int32(id))
		nox_xxx_wnd_46ABB0_go(win, int32(enabled))
	}
}

//export nox_gui_getWindowOffs_46AA20
func nox_gui_getWindowOffs_46AA20(win *C.nox_window, x, y *C.uint) C.int {
	if win == nil {
		*x, *y = 0, 0
		return -2
	}
	*x, *y = C.uint(win.off_x), C.uint(win.off_y)
	return 0
}

//export nox_client_wndGetPosition_46AA60
func nox_client_wndGetPosition_46AA60(win *C.nox_window, x, y *C.uint) C.int {
	if win == nil {
		return -2
	}
	*x, *y = C.uint(win.off_x), C.uint(win.off_y)
	for parent := win.parent; parent != nil; parent = parent.parent {
		*x += C.uint(parent.off_x)
		*y += C.uint(parent.off_y)
	}
	return 0
}

//export nox_xxx_wndPointInWnd_46AAB0
func nox_xxx_wndPointInWnd_46AAB0(win *C.nox_window, x, y C.int) C.bool {
	var left, top C.uint
	var width, height C.int
	nox_client_wndGetPosition_46AA60(win, &left, &top)
	nox_window_get_size(win, &width, &height)
	return C.bool(x >= C.int(left) && x <= C.int(left)+width &&
		y >= C.int(top) && y <= C.int(top)+height)
}

//export sub_46AB20
func sub_46AB20(win *C.nox_window, width, height C.int) C.int {
	if win == nil {
		return -2
	}
	win.width, win.height = C.intptr_t(width), C.intptr_t(height)
	win.end_x, win.end_y = win.off_x+win.width, win.off_y+win.height
	nox_window_call_field_94_fnc(win, 16388, uintptr(width), uintptr(height), nil, 0)
	return 0
}

//export nox_window_get_size
func nox_window_get_size(win *C.nox_window, width, height *C.int) C.int {
	if win == nil {
		*width, *height = 0, 0
		return -2
	}
	*width, *height = C.int(win.width), C.int(win.height)
	return 0
}

//export nox_xxx_wnd_46ABB0
func nox_xxx_wnd_46ABB0(win *C.nox_window, enabled C.int) C.int {
	return C.int(nox_xxx_wnd_46ABB0_go(win, int32(enabled)))
}

//export nox_xxx_wnd_46AD60
func nox_xxx_wnd_46AD60(win *C.nox_window, flags C.int) C.int {
	if win == nil {
		return -2
	}
	previous := win.flags
	win.flags |= C.nox_window_flags(flags)
	return C.int(previous)
}

//export nox_xxx_wndClearFlag_46AD80
func nox_xxx_wndClearFlag_46AD80(win *C.nox_window, flags C.int) C.int {
	if win == nil {
		return -2
	}
	previous := win.flags
	win.flags &^= C.nox_window_flags(flags)
	return C.int(previous)
}

//export nox_window_is_child
func nox_window_is_child(parent, child *C.nox_window) C.int {
	if parent == nil || child == nil {
		return 0
	}
	for ancestor := child.parent; ancestor != nil; ancestor = ancestor.parent {
		if ancestor == parent {
			return 1
		}
	}
	return 0
}

//export nox_xxx_wnd_46B280
func nox_xxx_wnd_46B280(win, owner *C.nox_window) C.int {
	if win == nil {
		return -2
	}
	if owner == nil {
		owner = win
	}
	win.draw_data.win = owner
	return 0
}
