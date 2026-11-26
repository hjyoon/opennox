package legacy

/*
#include "GAME3_1.h"
*/
import "C"

import (
	"unsafe"

	"github.com/opennox/opennox/v1/common/memmap"
)

var voteOverlayRoot *nox_window

//export sub_4C3390
func sub_4C3390() C.int {
	image := nox_xxx_gLoadImg_42F970(internCStr("VoteInProgress"))
	*memmap.PtrPtr(0x5D4594, 1321220) = unsafe.Pointer(image)
	root := nox_window_new(nil, 136, clientDisplayStorage.width-50, clientDisplayStorage.height/2-100, 50, 50, nil)
	storeWindowReference(&voteOverlayRoot, root)
	nox_xxx_wndSetIcon_46AE60(root, image)
	nox_window_set_all_funcs(root, nil, C.sub_4C3410, nil)
	nox_window_set_hidden(root, 1)
	return 1
}

//export sub_4C3410
func sub_4C3410(win *nox_window, draw unsafe.Pointer) C.int {
	pos := asWindow(win).GlobalPos()
	d := &win.draw_data
	nox_client_drawImageAt_47D2C0((*nox_video_bag_image_t)(d.bg_image), int32(pos.X)+int32(d.img_px), int32(pos.Y)+int32(d.img_py))
	return 1
}

//export sub_4C3460
func sub_4C3460(hidden C.int) C.int {
	return C.int(nox_window_set_hidden(voteOverlayRoot, int32(hidden)))
}

//export sub_4C34A0
func sub_4C34A0() C.int {
	result := nox_xxx_windowDestroyMB_46C4E0(voteOverlayRoot)
	voteOverlayRoot = nil
	return C.int(result)
}
