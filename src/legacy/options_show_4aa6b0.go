package legacy

/*
#include "client__gui__window.h"
#include "GAME2_1.h"
#include "GAME2_3.h"
#include "GAME3.h"
#include "GAME3_1.h"
extern nox_window* dword_5d4594_1309720;
extern nox_window* dword_5d4594_1309728;
extern nox_window* dword_5d4594_1309732;
extern nox_window* dword_5d4594_1309736;
extern nox_gui_animation* nox_wnd_xxx_1309740;
*/
import "C"

import (
	"unsafe"

	"github.com/opennox/opennox/v1/client/gui"
)

func OptionsShowCEntry4AA6B0() int32 { return int32(C.nox_game_showOptions_4AA6B0()) }

func OptionsNewWindow4AA6B0(name string) *gui.Window {
	return asWindow(C.nox_new_window_from_file(internCStr(name), C.sub_4AABE0))
}

func OptionsStoreRoot4AA6B0(win *gui.Window) {
	C.dword_5d4594_1309720 = (*nox_window)(unsafe.Pointer(win))
}

func OptionsNewAdvanced4AA6B0(win *gui.Window) int32 {
	return int32(C.nox_client_advVideoOpts_New_4CB590((*nox_window)(unsafe.Pointer(win))))
}

func OptionsStoreAnimation4AA6B0(anim *gui.Anim) {
	C.nox_wnd_xxx_1309740 = (*C.nox_gui_animation)(unsafe.Pointer(anim))
}

func OptionsStartOut4AA6B0() unsafe.Pointer { return C.sub_4AA9C0 }
func OptionsDoneOut4AA6B0() unsafe.Pointer  { return C.sub_4AAA10 }

func OptionsStoreCheckbox4AA6B0(channel int32, win *gui.Window) {
	ptr := (*nox_window)(unsafe.Pointer(win))
	switch channel {
	case 0:
		C.dword_5d4594_1309728 = ptr
	case 1:
		C.dword_5d4594_1309732 = ptr
	default:
		C.dword_5d4594_1309736 = ptr
	}
}

func OptionsCheckbox4AA6B0(channel int32) *gui.Window {
	switch channel {
	case 0:
		return asWindow(C.dword_5d4594_1309728)
	case 1:
		return asWindow(C.dword_5d4594_1309732)
	default:
		return asWindow(C.dword_5d4594_1309736)
	}
}

// GAME.EXE passes root to this three-byte no-op; the retained C no-op has
// no parameters. Keep the root reload at the caller, without showing a modal.
func OptionsReturnNull4AA6B0(_ *gui.Window) { C.nox_xxx_wndRetNULL_46A8A0() }
func OptionsInitVideo4AA6B0()               { C.sub_4AAA70() }
