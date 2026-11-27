package legacy

/*
#include "client__gui__window.h"
#include "GAME2_1.h"
#include "GAME2_3.h"
#include "GAME3.h"
#include "GAME3_1.h"
*/
import "C"

import (
	"unsafe"

	"github.com/opennox/opennox/v1/client/gui"
)

var optionsRoot *nox_window
var optionsAudioChecks [3]*nox_window

func OptionsShowCEntry4AA6B0() int32 { return int32(C.nox_game_showOptions_4AA6B0()) }

func OptionsNewWindow4AA6B0(name string) *gui.Window {
	return asWindow(nox_new_window_from_file(internCStr(name), C.sub_4AABE0))
}

func OptionsStoreRoot4AA6B0(win *gui.Window) {
	optionsRoot = (*nox_window)(unsafe.Pointer(win))
}

func OptionsNewAdvanced4AA6B0(win *gui.Window) int32 {
	return int32(C.nox_client_advVideoOpts_New_4CB590((*nox_window)(unsafe.Pointer(win))))
}

func OptionsStoreAnimation4AA6B0(anim *gui.Anim) {
	optionsAnimation = anim
}

func OptionsStartOut4AA6B0() unsafe.Pointer { return C.sub_4AA9C0 }
func OptionsDoneOut4AA6B0() unsafe.Pointer  { return C.sub_4AAA10 }

func OptionsStoreCheckbox4AA6B0(channel int32, win *gui.Window) {
	ptr := (*nox_window)(unsafe.Pointer(win))
	switch channel {
	case 0:
		optionsAudioChecks[0] = ptr
	case 1:
		optionsAudioChecks[1] = ptr
	default:
		optionsAudioChecks[2] = ptr
	}
}

func OptionsCheckbox4AA6B0(channel int32) *gui.Window {
	switch channel {
	case 0:
		return asWindow(optionsAudioChecks[0])
	case 1:
		return asWindow(optionsAudioChecks[1])
	default:
		return asWindow(optionsAudioChecks[2])
	}
}

// GAME.EXE passes root to this three-byte no-op; the retained C no-op has
// no parameters. Keep the root reload at the caller, without showing a modal.
func OptionsReturnNull4AA6B0(_ *gui.Window) { C.nox_xxx_wndRetNULL_46A8A0() }
func OptionsInitVideo4AA6B0()               { C.sub_4AAA70() }
