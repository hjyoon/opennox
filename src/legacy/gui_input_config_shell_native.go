package legacy

/*
#include "GAME2_2.h"
#include "GAME3_1.h"
#include "client__shell__inputcfg__inputcfg.h"
#include "client__shell__optsback.h"
extern nox_gui_animation* nox_wnd_xxx_1522608;
extern nox_window *dword_5d4594_1522604, *dword_5d4594_1522612, *dword_5d4594_1522616;
extern nox_window *dword_5d4594_1522620, *dword_5d4594_1522624, *dword_5d4594_1522628, *dword_5d4594_1522632;
*/
import "C"

import (
	"unsafe"

	"github.com/opennox/opennox/v1/client/gui"
	"github.com/opennox/opennox/v1/legacy/common/ccall"
)

//export sub_4CB880
func sub_4CB880() C.int {
	nox_game_addStateCode_43BDD0(900)
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

	actions := g2ServerChild(root, 910)
	labels := g2ServerChild(root, 911)
	primary := g2ServerChild(root, 912)
	secondary := g2ServerChild(root, 913)
	C.dword_5d4594_1522616, C.dword_5d4594_1522620 = actions, labels
	C.dword_5d4594_1522624, C.dword_5d4594_1522628 = primary, secondary
	if actions == nil || labels == nil || primary == nil || secondary == nil {
		return 0
	}
	list := g2ServerList(actions)
	if list == nil || list.Field_7 == nil || list.Field_8 == nil || list.Field_9 == nil {
		return 0
	}
	nox_xxx_wndSetID_46B080((*nox_window)(list.Field_7), 921)
	nox_xxx_wndSetID_46B080((*nox_window)(list.Field_8), 922)
	nox_xxx_wndSetID_46B080((*nox_window)(list.Field_9), 920)
	nox_xxx_wndSetProc_46B2C0_go(actions, C.sub_4CBF60)
	columns := [...]*nox_window{labels, primary, secondary}
	for _, column := range columns {
		sub_46B120(column, actions)
	}
	nox_xxx_wndSetWindowProc_46B300_go(primary, C.sub_4CC140)
	nox_xxx_wndSetWindowProc_46B300_go(secondary, C.sub_4CC140)
	up, down := g2ServerChild(root, 921), g2ServerChild(root, 922)
	for _, column := range columns {
		g2ServerText(column, 16408, unsafe.Pointer(up), 0)
		g2ServerText(column, 16409, unsafe.Pointer(down), 0)
	}
	C.sub_4CBBF0()
	for id, end := 971, 971+int(C.sub_47DBC0()); id < end; id++ {
		g2ServerEnabled(g2ServerChild(root, id), true)
	}
	if button := g2ServerChild(root, 971+int(nox_client_mousePriKey_430AF0())); button != nil {
		g2ServerEvent(button, 16392, 1, 0)
	}

	prompt := g2ServerChild(root, 980)
	C.dword_5d4594_1522612 = prompt
	if prompt == nil {
		return 0
	}
	sub_46B120(prompt, nil)
	nox_xxx_wndSetProc_46B2C0_go(prompt, C.sub_4CBE70)
	nox_xxx_wndSetWindowProc_46B300_go(prompt, C.sub_4CC170)
	g2ServerHidden(prompt, true)
	if label := g2ServerChild(prompt, 981); label != nil {
		g2ServerText(label, 16385, g2ServerMem(1522636), 0)
	}
	return 1
}

//export sub_4CBB70
func sub_4CBB70() C.int {
	C.sub_4CBD30()
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
