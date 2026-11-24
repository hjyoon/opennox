package legacy

/*
#include "GAME2_2.h"
#include "GAME3_1.h"
#include "client__shell__inputcfg__inputcfg.h"
#include "client__shell__optsback.h"
extern nox_gui_animation* nox_wnd_xxx_1522608;
extern nox_window *dword_5d4594_1522604, *dword_5d4594_1522612, *dword_5d4594_1522616;
extern nox_window *dword_5d4594_1522632;
*/
import "C"

import (
	"github.com/opennox/opennox/v1/client/gui"
	"github.com/opennox/opennox/v1/legacy/common/ccall"
)

//export sub_4CB880
func sub_4CB880() C.int {
	nox_game_addStateCode_43BDD0(900)
	p := shellInputBindings
	root := nox_new_window_from_file(internCStr("InputCfg.wnd"), C.sub_4CBE70)
	C.dword_5d4594_1522604 = root
	if root == nil {
		return 0
	}
	nox_xxx_wndSetWindowProc_46B300_go(root, C.sub_4A18E0)
	anim := nox_gui_makeAnimation_43C5B0(root, 0, 0, 0, -480, 0, 20, 0, -40)
	C.nox_wnd_xxx_1522608 = anim
	if anim == nil {
		return 0
	}
	a := asGUIAnim(anim)
	a.StateID = 900
	a.Func12Ptr = C.sub_4CBB70
	a.FncDoneOutPtr = C.sub_4CBBB0

	if !p.attachLists(&C.dword_5d4594_1522616, C.sub_4CBF60, C.sub_4CC140) {
		return 0
	}
	sub_4CBBF0()
	p.setMouseButtons()

	prompt := g2ServerChild(root, 980)
	C.dword_5d4594_1522612 = prompt
	if prompt == nil {
		return 0
	}
	sub_46B120(prompt, nil)
	nox_xxx_wndSetProc_46B2C0_go(prompt, C.sub_4CBE70)
	nox_xxx_wndSetWindowProc_46B300_go(prompt, C.sub_4CC170)
	g2ServerHidden(prompt, true)
	p.bindPromptLabel()
	return 1
}

//export sub_4CBB70
func sub_4CBB70() C.int {
	shellInputBindings.apply()
	nox_common_writecfgfile(internCStr("nox.cfg"))
	asGUIAnim(C.nox_wnd_xxx_1522608).SetState(gui.AnimOut)
	sub_43BE40(2)
	nox_xxx_clientPlaySoundSpecial_452D80(923, 100)
	return 1
}

//export sub_4CBBB0
func sub_4CBBB0() C.int {
	anim := C.nox_wnd_xxx_1522608
	next := asGUIAnim(anim).Func13Ptr
	nox_gui_freeAnimation_43C570(anim)
	nox_xxx_windowDestroyMB_46C4E0(C.dword_5d4594_1522604)
	nox_xxx_windowDestroyMB_46C4E0(C.dword_5d4594_1522612)
	C.dword_5d4594_1522604, C.dword_5d4594_1522612, C.dword_5d4594_1522632 = nil, nil, nil
	ccall.CallIntVoid(next)
	return 1
}

//export sub_4CC140
func sub_4CC140(win *nox_window, event C.int, a, b C.uintptr_t) C.int {
	return sub_4C3A60(win, event, a, b)
}

//export sub_4CC170
func sub_4CC170(win *nox_window, event C.int, a, b C.uintptr_t) C.int {
	return shellInputBindings.handleCapture(event, a, b)
}

//export sub_4CC280
func sub_4CC280(key C.uint) C.int { return shellInputBindings.setKey(key) }

//export sub_4CC3C0
func sub_4CC3C0(key C.uint) C.int { return shellInputBindings.setButton(key) }
