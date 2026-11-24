package legacy

/*
#include "GAME2_1.h"
#include "GAME3_1.h"
#include "client__gui__guiinput.h"
#include "noxstring.h"
extern int nox_win_width, nox_win_height;
extern nox_window *dword_5d4594_1321224, *dword_5d4594_1321228, *dword_5d4594_1321232;
extern nox_window *dword_5d4594_1321236, *dword_5d4594_1321240, *dword_5d4594_1321244;
extern nox_window *dword_5d4594_1321248, *dword_5d4594_1321252;
extern nox_window *dword_5d4594_1522612, *dword_5d4594_1522632, *dword_5d4594_1522624, *dword_5d4594_1522628;
*/
import "C"

import (
	"unsafe"

	"github.com/opennox/opennox/v1/common/memmap"
)

//export sub_4C3500
func sub_4C3500() C.int {
	root := nox_new_window_from_file(internCStr("yesno.wnd"), nil)
	C.dword_5d4594_1321224 = root
	if root != nil {
		nox_window_setPos_46A9B0(root, int32(C.nox_win_width-320)/2, int32(C.nox_win_height-240)/2)
	}
	nox_window_set_hidden(root, 1)
	return C.int(bool2int(root != nil))
}

//export sub_4C35B0
func sub_4C35B0(cancel C.int) C.int {
	if wndIsShown_nox_xxx_wndIsShown_46ACC0(C.dword_5d4594_1321228) != 0 {
		return 0
	}
	sub_413A00(0)
	if cancel != 0 {
		g2ServerHidden(C.dword_5d4594_1321228, true)
	} else {
		C.sub_4C3620()
		nox_common_writecfgfile(internCStr("nox.cfg"))
		g2ServerHidden(C.dword_5d4594_1321228, true)
		C.sub_472280()
		sub_4ADA40()
	}
	return 1
}

//export sub_4C3A60
func sub_4C3A60(win *nox_window, event C.int, a, b C.uintptr_t) C.int {
	if event == 19 || event == 20 {
		return 0
	}
	return nox_xxx_wndListboxProcWithoutData10_4A28E0(win, event, a, b)
}

type inputBindingPanel struct {
	prompt       **nox_window
	selected     **nox_window
	columns      [2]**nox_window
	keyEmpty     uintptr
	restoreFocus bool
}

var gameInputBindings = inputBindingPanel{
	prompt:   &C.dword_5d4594_1321232,
	selected: &C.dword_5d4594_1321252,
	columns:  [2]**nox_window{&C.dword_5d4594_1321244, &C.dword_5d4594_1321248},
	keyEmpty: 185444,
}

var shellInputBindings = inputBindingPanel{
	prompt:       &C.dword_5d4594_1522612,
	selected:     &C.dword_5d4594_1522632,
	columns:      [2]**nox_window{&C.dword_5d4594_1522624, &C.dword_5d4594_1522628},
	keyEmpty:     187824,
	restoreFocus: true,
}

func (p inputBindingPanel) hideCapture() {
	nox_xxx_windowFocus_46B500(nil)
	nox_xxx_wnd_46C6E0(*p.prompt)
	g2ServerHidden(*p.prompt, true)
	if p.restoreFocus {
		nox_gui_input_cfg_restore_focus()
	}
}

//export sub_4C3EB0
func sub_4C3EB0(win *nox_window, event C.int, a, b C.uintptr_t) C.int {
	return gameInputBindings.handleCapture(event, a, b)
}

func (p inputBindingPanel) handleCapture(event C.int, a, b C.uintptr_t) C.int {
	var button C.uint
	switch event {
	case 6, 7:
		button = 0x10000
	case 10, 11:
		button = 0x10002
	case 14, 15:
		button = 0x10001
	case 19:
		button = 0x10003
	case 20:
		button = 0x10004
	case 21:
		if a == 1 {
			if b == 2 {
				p.hideCapture()
				if selection := *p.selected; selection != nil {
					g2ServerEvent(selection, 16403, ^uintptr(0), 0)
				}
			}
			return 1
		}
		if b == 1 && p.setKey(C.uint(a)) != 0 {
			button = 0x10000
		} else {
			return 0
		}
	default:
		return 0
	}
	p.setButton(button)
	p.hideCapture()
	return 1
}

func clearInputBinding(list *nox_window, name *wchar2_t, empty unsafe.Pointer) {
	data := g2ServerList(list)
	if data == nil {
		return
	}
	count := int(data.Field_11_0)
	for i := 0; i < count; i++ {
		current := (*wchar2_t)(unsafe.Pointer(g2ServerEvent(list, 16406, uintptr(i), 0)))
		if current != nil && C.nox_wcscmp(current, name) == 0 {
			g2ServerText(list, 16407, empty, uintptr(i))
		}
	}
}

func (p inputBindingPanel) set(name *wchar2_t, offset uintptr) {
	for i, column := range p.columns {
		clearInputBinding(*column, name, memmap.PtrOff(0x587000, offset+4*uintptr(i)))
	}
	index := int32(g2ServerEvent(*p.selected, 16404, 0, 0))
	g2ServerText(*p.selected, 16407, unsafe.Pointer(name), uintptr(index))
	g2ServerEvent(*p.selected, 16403, ^uintptr(0), 0)
	*p.selected = nil
}

func (p inputBindingPanel) setKey(key C.uint) C.int {
	name := nox_xxx_keybind_titleByKeyZero_42EA00(key)
	if name == nil {
		return 0
	}
	if *p.selected != nil {
		p.set(name, p.keyEmpty)
	}
	return 1
}

func (p inputBindingPanel) setButton(key C.uint) C.int {
	name := nox_xxx_keybind_titleByKey_42EA00(key)
	if *p.selected == nil {
		return 1
	}
	if name == nil {
		return 0
	}
	p.set(name, p.keyEmpty+8)
	return 1
}

//export sub_4C3FC0
func sub_4C3FC0(key C.uint) C.int { return gameInputBindings.setKey(key) }

//export sub_4C4100
func sub_4C4100(key C.uint) C.int { return gameInputBindings.setButton(key) }

//export sub_4C4220
func sub_4C4220() C.int {
	nox_xxx_windowDestroyMB_46C4E0(C.dword_5d4594_1321228)
	nox_xxx_windowDestroyMB_46C4E0(C.dword_5d4594_1321232)
	C.dword_5d4594_1321228, C.dword_5d4594_1321232 = nil, nil
	C.dword_5d4594_1321236, C.dword_5d4594_1321240 = nil, nil
	C.dword_5d4594_1321244, C.dword_5d4594_1321248 = nil, nil
	return 0
}

//export sub_4C4260
func sub_4C4260() {
	C.sub_4C3B70()
	nox_xxx_wndShowModalMB_46A8C0(C.dword_5d4594_1321228)
	sub_413A00(1)
}

//export sub_4C4280
func sub_4C4280() C.int {
	return C.int(bool2int(wndIsShown_nox_xxx_wndIsShown_46ACC0(C.dword_5d4594_1321228) == 0))
}
