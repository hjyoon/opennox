package legacy

/*
#include "game2_server_bridge.h"
#include "client__gui__servopts__advserv.h"
#include "client__gui__servopts__general.h"


static void nox_format_server_audio_cutoff(wchar2_t* out, const wchar2_t* format, int value) {
	nox_swprintf(out, format, value);
}
*/
import "C"

import (
	"encoding/binary"
	"unsafe"

	noxflags "github.com/opennox/opennox/v1/common/flags"
)

const serverGeneralStrings = `C:\NoxPost\src\client\Gui\ServOpts\general.c`

var serverGeneralWindow *nox_window

func serverConnectionLabel(offset uintptr, line int32) *wchar2_t {
	key := (*C.char)(g2ServerPointerSlot(0x587000, offset))
	if key == nil {
		return nil
	}
	return nox_strman_loadString_40F1D0(key, nil, internCStr(serverGeneralStrings), line)
}

//export nox_xxx_gui_4AD320
func nox_xxx_gui_4AD320(parent *C.nox_window) *C.nox_window {
	lang := nox_strman_get_lang_code()
	if nox_xxx_guiFontHeightMB_43F320(nil) > 10 {
		lang = 2
	}
	if serverGeneralWindow != nil {
		return nil
	}
	table := int32(173556)
	if Get_dword_5d4594_2650652() == 1 {
		table = 173596
	}
	path := (*C.char)(g2ServerPointerSlot(0x587000, uintptr(table+4*lang)))
	if path == nil {
		return nil
	}
	proc := unsafe.Pointer(C.nox_xxx_windowServerOptionsGeneralProc_4AD5D0)
	root := nox_new_window_from_file(path, proc)
	serverGeneralWindow = root
	if root == nil {
		return nil
	}
	asWindow(root).AddCleanup(func() {
		if serverGeneralWindow == root {
			serverGeneralWindow = nil
		}
	})
	C.sub_46B120(root, parent)
	nox_xxx_wndSetDrawFn_46B340(root, unsafe.Pointer(C.sub_4AD570))
	if spells := g2ServerChild(root, 10306); spells != nil && noxflags.HasGame(noxflags.GameModeCTF|noxflags.GameModeElimination) {
		g2ServerEnabled(spells, false)
	}
	nox_xxx_wndRetNULL_46A8A0()
	sub_4AD840()
	if Get_dword_5d4594_2650652() == 1 {
		sub_4AD4B0()
		if owner := g2ServerChild(root, 10310); owner != nil {
			nox_xxx_wnd_46B280(owner, root)
			nox_xxx_wndSetProc_46B2C0(owner, proc)
		}
		if list := g2ServerChild(root, 10317); list != nil {
			for offset := uintptr(173540); offset < 173556; offset += 4 {
				if text := serverConnectionLabel(offset, 308); text != nil {
					g2ServerText(list, 16397, unsafe.Pointer(text), ^uintptr(0))
				}
			}
		}
	} else if advanced := g2ServerChild(root, 10319); advanced != nil {
		g2ServerHidden(advanced, true)
	}
	if noxflags.HasEngine(noxflags.EngineNoRendering) {
		if video := g2ServerChild(root, 10304); video != nil {
			g2ServerEnabled(video, false)
		}
	}
	return serverGeneralWindow
}

//export sub_4AD4B0
func sub_4AD4B0() C.int {
	root := serverGeneralWindow
	if root == nil {
		return 0
	}
	list, button := g2ServerChild(root, 10317), g2ServerChild(root, 10316)
	if list == nil || button == nil {
		return 0
	}
	list.height = C.intptr_t(4*(nox_xxx_guiFontHeightMB_43F320(list.draw_data.font)+1) + 2)
	list.end_y = list.off_y + list.height
	var widest C.int
	for offset := uintptr(173540); offset < 173556; offset += 4 {
		if text := serverConnectionLabel(offset, 92); text != nil {
			var width C.int
			nox_xxx_drawGetStringSize_43F840(list.draw_data.font, text, &width, nil, 0)
			widest = max(widest, width)
		}
	}
	list.width = max(button.width, C.intptr_t(widest)+7)
	list.off_x = list.end_x - list.width
	return C.int(button.width)
}

func markServerGeneralChoice(root *C.nox_window, id int) {
	if control := g2ServerChild(root, id); control != nil {
		control.draw_data.field_0 |= 4
	}
}

//export sub_4AD840
func sub_4AD840() C.int {
	root := serverGeneralWindow
	if root == nil {
		return 0
	}
	if nox_server_doPlayersAutoRespawn_40A5F0() != 0 {
		if control := g2ServerChild(root, 10301); control != nil {
			control.draw_data.field_0 |= 4
			if noxflags.HasGame(noxflags.GameModeElimination) {
				g2ServerEnabled(control, false)
			}
		}
	}
	if Get_nox_server_sendMotd_108752() != 0 {
		markServerGeneralChoice(root, 10302)
	}
	if sub_4D0D70() != 0 {
		markServerGeneralChoice(root, 10304)
	}
	if sub_409F40(2) != 0 {
		markServerGeneralChoice(root, 10305)
	}
	if sub_409F40(0x2000) != 0 {
		markServerGeneralChoice(root, 10306)
	}
	if Get_dword_5d4594_2650652() != 1 {
		return 0
	}
	conn := Get_nox_server_connectionType_3596()
	if conn < 1 || conn > 4 {
		return 0
	}
	kind, rate := g2ServerChild(root, 10316), g2ServerChild(root, 10312)
	if kind == nil || rate == nil {
		return 0
	}
	label := serverConnectionLabel(uintptr(173536+4*conn), 391)
	if label == nil {
		return 0
	}
	g2ServerText(kind, 16385, unsafe.Pointer(label), ^uintptr(0))
	return C.int(g2ServerEvent(rate, 16394, uintptr(int64(4-nox_xxx_rateGet_40A6C0())), 0))
}

//export sub_4BE2C0
func sub_4BE2C0(value C.int) C.int {
	field := unsafe.Slice((*byte)(unsafe.Add(sub_416640(), 74)), 4)
	binary.LittleEndian.PutUint32(field, uint32(value))
	dst := (*C.wchar2_t)(g2ServerMem(1316716))
	format := g2ServerLoad("AudCullDesc", `C:\NoxPost\src\client\Gui\ServOpts\advserv.c`, 71)
	C.nox_format_server_audio_cutoff(dst, format, value)
	control := g2ServerChild(serverPerformanceWindow, 2120)
	return C.int(g2ServerText(control, 16385, unsafe.Pointer(dst), ^uintptr(0)))
}
