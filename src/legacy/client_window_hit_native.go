package legacy

/*
#include "GAME2_1.h"
*/
import "C"

//export sub_46B630
func sub_46B630(root *C.nox_window, x, y C.int) *C.nox_window {
	window := root
	for window != nil {
		var next *C.nox_window
		for child := window.field_100; child != nil; child = child.prev {
			left, top := child.off_x, child.off_y
			for parent := child.parent; parent != nil; parent = parent.parent {
				left += parent.off_x
				top += parent.off_y
			}
			if C.intptr_t(x) >= left && C.intptr_t(x) <= left+child.width &&
				C.intptr_t(y) >= top && C.intptr_t(y) <= top+child.height &&
				child.flags&C.NOX_WIN_HIDDEN == 0 {
				next = child
				break
			}
		}
		if next == nil {
			break
		}
		window = next
	}
	return window
}
