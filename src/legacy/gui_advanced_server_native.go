package legacy

/*
#include "cgo_const.h"
#include "GAME2_2.h"
#include "GAME3_1.h"
#include "client__gui__servopts__advserv.h"
#include "client__gui__servopts__objlst.h"
#include "client__gui__servopts__spelllst.h"
#include "client__gui__window.h"
extern uint32_t dword_5d4594_1316704;
extern nox_window *dword_5d4594_1316708, *dword_5d4594_1316712, *dword_5d4594_1316972;
*/
import "C"

import (
	"encoding/binary"
	"unsafe"

	noxflags "github.com/opennox/opennox/v1/common/flags"
	"github.com/opennox/opennox/v1/common/memmap"
	"github.com/opennox/opennox/v1/server"
)

//export nox_xxx_loadAdvancedWnd_4BDC10
func nox_xxx_loadAdvancedWnd_4BDC10(settings *C.int) C.int {
	if C.dword_5d4594_1316708 != nil || settings == nil {
		return 0
	}
	root := nox_new_window_from_file(internCStr("advanced.wnd"), C.nox_xxx_windowAdvancedServProc_4BDDB0)
	C.dword_5d4594_1316708 = root
	if root == nil {
		return 0
	}
	sub_46B120(root, nil)
	sub_46C690(root)
	nox_xxx_wndShowModalMB_46A8C0(root)
	C.nox_xxx_wndSetWindowProc_46B300(root, C.sub_4BDDA0)
	nox_xxx_windowFocus_46B500(root)
	return sub_4BDC70(settings)
}

func setAdvancedRadio(win *nox_window, selected bool) {
	if win == nil {
		return
	}
	if selected {
		win.draw_data.field_0 |= 4
	} else {
		win.draw_data.field_0 &^= 4
	}
}

func applyAdvancedFilters(settings *server.Settings2) {
	sub_453F70((*C.nox_cgo_const_void)(unsafe.Pointer(&settings.Field24)))
	sub_4535E0((*C.int)(unsafe.Pointer(&settings.Field44[0])))
	sub_4535F0(C.int(settings.Field48))
}

//export sub_4BDC70
func sub_4BDC70(settings *C.int) C.int {
	root := C.dword_5d4594_1316708
	if settings == nil || root == nil {
		return 0
	}
	if noxflags.HasGame(noxflags.GameHost) {
		setAdvancedRadio(g2ServerChild(root, 10167), true)
		C.dword_5d4594_1316704 = 0
	} else {
		setAdvancedRadio(g2ServerChild(root, 10164), true)
		if rules := g2ServerChild(root, 10167); rules != nil {
			g2ServerEnabled(rules, false)
		}
		C.dword_5d4594_1316704 = 1
	}
	applyAdvancedFilters((*server.Settings2)(unsafe.Pointer(settings)))
	return sub_4BDD10()
}

//export sub_4BDD10
func sub_4BDD10() C.int {
	root := C.dword_5d4594_1316708
	if root == nil {
		return 0
	}
	var tab *nox_window
	switch C.dword_5d4594_1316704 {
	case 0:
		tab = sub_4CEBA0(root, sub_4165B0())
	case 1:
		tab = C.nox_xxx_guiSpelllistLoad_453850(root)
	case 2:
		tab = C.nox_xxx_guiObjlistLoad_4530C0(root, 0x1000000)
	case 3:
		tab = C.nox_xxx_guiObjlistLoad_4530C0(root, 0x2000000)
	default:
		return 0
	}
	C.dword_5d4594_1316712 = tab
	if tab == nil {
		return 0
	}
	return C.int(nox_xxx_windowFocus_46B500(tab))
}

//export sub_4BDDA0
func sub_4BDDA0(win *nox_window, event C.int, control *nox_window, eventArg C.uintptr_t) C.int {
	return 1
}

//export nox_xxx_windowAdvancedServProc_4BDDB0
func nox_xxx_windowAdvancedServProc_4BDDB0(win *nox_window, event C.int, control *nox_window, eventArg C.uintptr_t) C.int {
	if event != 16391 || control == nil {
		return 1
	}
	id := int(C.nox_xxx_wndGetID_46B0A0(control))
	nox_xxx_clientPlaySoundSpecial_452D80(766, 100)
	if id == 10148 {
		if settings := (*server.Settings2)(unsafe.Pointer(sub_4165B0())); settings != nil {
			settings.Field24 = *(*server.Settings3)(unsafe.Pointer(sub_453F90()))
			settings.Field44 = *(*[4]byte)(unsafe.Pointer(sub_453600()))
			settings.Field48 = uint32(sub_453610())
		}
		sub_4BDF30()
		return 1
	}
	if id < 10164 || id > 10167 {
		return 1
	}
	if tab := C.dword_5d4594_1316712; tab != nil {
		C.nox_xxx_windowDestroyMB_46C4E0(tab)
		C.dword_5d4594_1316712 = nil
	}
	C.dword_5d4594_1316704 = C.uint32_t((id - 10163) % 4)
	sub_4BDD10()
	return 1
}

//export sub_4BDF30
func sub_4BDF30() C.int {
	root := C.dword_5d4594_1316708
	if root == nil {
		return 0
	}
	C.nox_xxx_wnd_46C6E0(root)
	C.nox_xxx_windowDestroyMB_46C4E0(root)
	C.dword_5d4594_1316708 = nil
	C.dword_5d4594_1316712 = nil
	return C.int(nox_xxx_windowFocus_46B500(nil))
}

//export sub_4BDF70
func sub_4BDF70(settings *C.int) C.int {
	if C.dword_5d4594_1316708 == nil {
		return 0
	}
	return sub_4BDF90(settings)
}

//export sub_4BDF90
func sub_4BDF90(settings *C.int) C.int {
	if settings == nil {
		return 0
	}
	applyAdvancedFilters((*server.Settings2)(unsafe.Pointer(settings)))
	switch C.dword_5d4594_1316704 {
	case 1:
		return sub_454120()
	case 2, 3:
		return sub_453750()
	default:
		return 0
	}
}

const (
	advancedPerfDrawMode  = 58
	advancedPerfLockMode  = 62
	advancedPerfRateMode  = 66
	advancedPerfRateValue = 70
)

func performanceRecord(ptr unsafe.Pointer) []byte {
	return unsafe.Slice((*byte)(ptr), advancedPerfRateValue+4)
}

//export sub_4BDFD0
func sub_4BDFD0() C.int {
	settings := sub_416640()
	lang := nox_strman_get_lang_code()
	if nox_xxx_guiFontHeightMB_43F320(nil) > 10 {
		lang = 2
	}
	resource := (*C.char)(g2ServerPointerSlot(0x587000, uintptr(180048+4*lang)))
	root := nox_new_window_from_file(resource, C.sub_4BE330)
	if root == nil {
		return 0
	}
	C.dword_5d4594_1316972 = root
	sub_46B120(root, nil)
	sub_46C690(root)
	nox_xxx_wndShowModalMB_46A8C0(root)
	nox_xxx_windowFocus_46B500(root)
	C.nox_xxx_wndSetWindowProc_46B300(root, C.sub_4BE320)
	host := g2ServerChild(nil, 10100)
	var x, y C.uint
	C.nox_gui_getWindowOffs_46AA20(host, &x, &y)
	nox_window_setPos_46A9B0(root, int32(x)+15, int32(y)+80)
	if panel := g2ServerChild(root, 2104); panel != nil {
		C.nox_xxx_wnd_46B280(panel, root)
		C.nox_xxx_wndSetProc_46B2C0(panel, C.sub_4BE330)
	}
	return sub_4BE120(settings)
}

//export sub_4BE120
func sub_4BE120(ptr unsafe.Pointer) C.int {
	root := C.dword_5d4594_1316972
	if ptr == nil || root == nil {
		return 0
	}
	settings := performanceRecord(ptr)
	setAdvancedRadio(g2ServerChild(root, 2102), binary.LittleEndian.Uint32(settings[advancedPerfDrawMode:]) != 0)
	setAdvancedRadio(g2ServerChild(root, 2103), binary.LittleEndian.Uint32(settings[advancedPerfLockMode:]) != 0)
	entry := g2ServerChild(root, 2110)
	if entry != nil {
		g2ServerIntText(entry, int32(binary.LittleEndian.Uint32(settings[advancedPerfRateValue:])), ^uintptr(0))
	}
	mode := binary.LittleEndian.Uint32(settings[advancedPerfRateMode:])
	for id := 2106; id <= 2109; id++ {
		setAdvancedRadio(g2ServerChild(root, id), false)
	}
	if mode <= 3 {
		setAdvancedRadio(g2ServerChild(root, 2106+int(mode)), true)
	}
	if entry != nil {
		g2ServerEnabled(entry, mode == 3)
	}
	return 0
}

//export sub_4BE320
func sub_4BE320(win *nox_window, event C.int, arg1, arg2 C.uintptr_t) C.int {
	return 1
}

func performanceRate(entry *nox_window) (uint32, bool) {
	text := (*wchar2_t)(unsafe.Pointer(g2ServerEvent(entry, 16413, 0, 0)))
	if text == nil || *text == 0 {
		return 0, false
	}
	return uint32(max(0, int32(C.nox_wcstol(text, nil, 10)))), true
}

//export sub_4BE330
func sub_4BE330(win *nox_window, event C.int, arg1, arg2 C.uintptr_t) C.int {
	settings := performanceRecord(sub_416640())
	if event > 0x4007 {
		switch event {
		case 16393:
			C.sub_4BE2C0(C.int(arg2))
			nox_xxx_gameSetAudioFadeoutMb_501AC0(int32(arg2))
		case 16415:
			if entry := (*nox_window)(unsafe.Pointer(uintptr(arg1))); entry != nil {
				if value, ok := performanceRate(entry); ok && C.nox_xxx_wndGetID_46B0A0(entry) == 2110 {
					binary.LittleEndian.PutUint32(settings[advancedPerfRateValue:], value)
				}
			}
		}
		return 1
	}
	if event == 16387 {
		entry := g2ServerChild(C.dword_5d4594_1316972, int(int32(arg2)))
		if entry == nil || uint16(arg1) == 1 {
			return 0
		}
		if value, ok := performanceRate(entry); ok && arg2 == 2110 {
			binary.LittleEndian.PutUint32(settings[advancedPerfRateValue:], value)
		}
		return 1
	}
	control := (*nox_window)(unsafe.Pointer(uintptr(arg1)))
	if event != 16391 || control == nil {
		return 1
	}
	id := int(C.nox_xxx_wndGetID_46B0A0(control))
	nox_xxx_clientPlaySoundSpecial_452D80(766, 100)
	switch id {
	case 2102:
		field := settings[advancedPerfDrawMode:]
		binary.LittleEndian.PutUint32(field, binary.LittleEndian.Uint32(field)^1)
	case 2103:
		field := settings[advancedPerfLockMode:]
		binary.LittleEndian.PutUint32(field, binary.LittleEndian.Uint32(field)^1)
	case 2106, 2107, 2108, 2109:
		if entry := g2ServerChild(C.dword_5d4594_1316972, 2110); entry != nil {
			g2ServerEnabled(entry, id == 2109)
		}
		binary.LittleEndian.PutUint32(settings[advancedPerfRateMode:], uint32(id-2106))
		sub_40A6A0(1)
	case 2130:
		sub_4BE610()
	}
	return 1
}

//export sub_4BE610
func sub_4BE610() C.int {
	root := C.dword_5d4594_1316972
	if root == nil {
		return 0
	}
	C.nox_xxx_wnd_46C6E0(root)
	C.nox_xxx_windowDestroyMB_46C4E0(root)
	C.dword_5d4594_1316972 = nil
	return C.int(nox_xxx_windowFocus_46B500(nil))
}

//export sub_4BE800
func sub_4BE800(value C.int) C.int {
	*g2ServerWord(1316980) = uint32(value)
	return value
}

//export sub_4BE810
func sub_4BE810(a, b, c C.int, d C.char) C.char {
	*g2ServerWord(1316988) = uint32(a)
	*g2ServerWord(1316984) = uint32(b)
	*g2ServerWord(1316992) = uint32(c)
	*memmap.PtrUint8(0x5D4594, 1316996) = byte(d)
	return d
}
