package legacy

/*
#include "game2_server_bridge.h"

extern uint32_t dword_587000_129656;
extern void* dword_5d4594_1046356;
extern void* dword_5d4594_1046360;
extern nox_window* dword_5d4594_1046500;
extern nox_window* dword_5d4594_1046524;
extern nox_window* dword_5d4594_1046532;
extern nox_window* dword_5d4594_1046536;
*/
import "C"

import (
	"unsafe"

	noxflags "github.com/opennox/opennox/v1/common/flags"
	"github.com/opennox/opennox/v1/common/memmap"
)

func serverOptionsWindowFile() *C.char {
	lang := nox_strman_get_lang_code()
	if nox_xxx_guiFontHeightMB_43F320(nil) > 10 {
		lang = 2
	}
	if lang < 0 || lang >= 9 {
		lang = 0
	}
	// Use the string slots directly. The original table has four-byte
	// entries and its addresses cannot be read as native pointers.
	return (*C.char)(memmap.PtrOff(0x587000, 129868+16*uintptr(lang)))
}

//export nox_xxx_guiServerOptsLoad_457500
func nox_xxx_guiServerOptsLoad_457500() C.int {
	if nox_gui_xxx_check_446360() != 0 {
		return 1
	}
	if C.dword_5d4594_1046492 != nil {
		nox_xxx_clientPlaySoundSpecial_452D80(231, 100)
		nox_xxx_guiServerOptionsTryHide_4574D0()
		return 1
	}
	if noxflags.HasGame(noxflags.GameHost) {
		sub_459D50(1)
		if C.dword_587000_129656 != 0 {
			nox_common_list_clear_425760(&C.nox_gui_server_options_rules)
		}
	}
	pre := unsafe.Pointer(C.nox_xxx_guiServerOptionsProcPre_4585D0)
	root := nox_new_window_from_file(serverOptionsWindowFile(), pre)
	C.dword_5d4594_1046492 = root
	if root == nil {
		return 0
	}

	nox_draw_setTabWidth_43FE20(100)
	nox_window_setPos_46A9B0(root, int32(C.nox_win_width)-int32(root.width)-10, 0)
	nox_xxx_wndSetWindowProc_46B300_go(root, unsafe.Pointer(C.nox_xxx_guiServerOptionsProc_458590))
	nox_xxx_wndSetDrawFn_46B340_go(root, unsafe.Pointer(C.nox_xxx_windowServerOptionsDrawProc_458500))
	child := func(id int32) *nox_window { return nox_xxx_wndGetChildByID_46B0C0(root, id) }
	C.dword_5d4594_1046512 = child(10101)
	C.dword_5d4594_1046496 = child(10114)
	C.dword_5d4594_1046500 = child(10183)
	C.dword_5d4594_1046504 = child(10197)
	C.dword_5d4594_1046508 = child(10199)
	C.dword_5d4594_1046524 = child(10150)
	C.dword_5d4594_1046516 = child(10134)
	C.dword_5d4594_1046520 = child(10135)
	C.dword_5d4594_1046536 = child(10153)
	nox_gui_winSetFunc96_46B070(child(10331), unsafe.Pointer(C.nox_xxx_options_457AA0))
	nox_gui_winSetFunc96_46B070(child(10333), unsafe.Pointer(C.nox_xxx_options_457B00))

	panels := [...]*nox_window{
		C.dword_5d4594_1046524, C.dword_5d4594_1046532, C.dword_5d4594_1046536,
		C.dword_5d4594_1046500, C.dword_5d4594_1046504, C.dword_5d4594_1046508,
	}
	for _, panel := range panels {
		nox_xxx_wnd_46B280(panel, root)
	}
	for _, panel := range panels {
		nox_xxx_wndSetProc_46B2C0_go(panel, pre)
	}

	*memmap.PtrPtr(0x5D4594, 1046352) = unsafe.Pointer(nox_xxx_gLoadImg_42F970(internCStr("UITabs1")))
	C.dword_5d4594_1046356 = unsafe.Pointer(nox_xxx_gLoadImg_42F970(internCStr("UITabs2")))
	C.dword_5d4594_1046360 = unsafe.Pointer(nox_xxx_gLoadImg_42F970(internCStr("UITabs3")))
	list := C.dword_5d4594_1046496
	sliderImage := unsafe.Pointer(nox_xxx_gLoadImg_42F970(internCStr("UISlider")))
	sliderLit := unsafe.Pointer(nox_xxx_gLoadImg_42F970(internCStr("UISliderLit")))
	panel := C.dword_5d4594_1046500
	slider := nox_xxx_wndGetChildByID_46B0C0(panel, 10182)
	up := nox_xxx_wndGetChildByID_46B0C0(panel, 10180)
	down := nox_xxx_wndGetChildByID_46B0C0(panel, 10181)
	slider.field_100.width, slider.field_100.height = 16, 10
	sub_4B5700(slider, nil, nil, sliderImage, sliderLit, sliderLit)
	nox_xxx_wnd_46B280(slider, list)
	nox_xxx_wnd_46B280(up, list)
	nox_xxx_wnd_46B280(down, list)
	scroll := g2ServerList(list)
	scroll.Field_9, scroll.Field_7, scroll.Field_8 = unsafe.Pointer(slider), unsafe.Pointer(up), unsafe.Pointer(down)

	child(10160).draw_data.field_0 |= 4
	settings := sub_4165D0(0)
	if noxflags.HasGame(noxflags.GameHost) {
		sub_4161E0()
	}
	nox_gui_gamemode_load_457410()
	sub_457B60(settings)
	sub_457A10()
	if noxflags.HasGame(noxflags.GameHost) {
		sub_4165F0(0, 1)
		sub_4165D0(1)
	} else {
		team := child(10159)
		title := g2ServerLoad("servopts.wnd:teams", serverOptionsStringsFile, 1811)
		g2ServerText(team, 16385, unsafe.Pointer(title), ^uintptr(0))
		nox_xxx_wndWddSetTooltip_46B000(&team.draw_data, g2ServerLoad("servopts.wnd:TeamTT", serverOptionsStringsFile, 1812))
		g2ServerHidden(child(10149), false)
		g2ServerHidden(C.dword_5d4594_1046508, true)
		sub_46ACE0(root, 10145, 10146, 1)
	}
	if C.dword_587000_129656 != 0 {
		if sub_4D6F30() != 0 || nox_xxx_isQuest_4D6F50() != 0 {
			nox_server_parseCmdText_443C80(internWStr("execrul OTQuest.rul"), 1)
		} else if noxflags.HasEngine(noxflags.EngineNoRendering) {
			nox_server_parseCmdText_443C80(internWStr("execrul server.rul"), 1)
		}
	}
	C.dword_587000_129656 = 0
	return 1
}
