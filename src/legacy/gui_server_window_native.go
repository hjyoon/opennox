package legacy

/*
#include "game2_server_bridge.h"

*/
import "C"

import (
	"unsafe"

	noxflags "github.com/opennox/opennox/v1/common/flags"
	"github.com/opennox/opennox/v1/common/memmap"
	"github.com/opennox/opennox/v1/legacy/common/alloc"
)

type serverOptionsChildWindows struct {
	mapList     *nox_window
	mapPanel    *nox_window
	pane10197   *nox_window
	pane10199   *nox_window
	lessonLimit *nox_window
	timeLimit   *nox_window
	pane10150   *nox_window
	accessTab   *nox_window
	playersTab  *nox_window
	pane10153   *nox_window
	generalTab  *nox_window
}

var serverOptionsChildren serverOptionsChildWindows
var serverOptionsTabImage2, serverOptionsTabImage3 unsafe.Pointer

var serverSharedWindows, _ = alloc.New(C.nox_game2_server_windows_t{})

//export nox_game2_server_windows_storage
func nox_game2_server_windows_storage() *C.nox_game2_server_windows_t { return serverSharedWindows }

func serverOptionsWindowFile() *C.char {
	lang := nox_strman_get_lang_code()
	if nox_xxx_guiFontHeightMB_43F320(nil) > 10 {
		lang = 2
	}
	if lang < 0 || lang >= 9 {
		lang = 0
	}
	return (*C.char)(*memmap.PtrPtr(0x587000, 129760+4*uintptr(lang)))
}

//export nox_xxx_guiServerOptsLoad_457500
func nox_xxx_guiServerOptsLoad_457500() C.int {
	if nox_gui_xxx_check_446360() != 0 {
		return 1
	}
	if serverSharedWindows.options != nil {
		nox_xxx_clientPlaySoundSpecial_452D80(231, 100)
		nox_xxx_guiServerOptionsTryHide_4574D0()
		return 1
	}
	if noxflags.HasGame(noxflags.GameHost) {
		sub_459D50(1)
		if legacyWords.dword_587000_129656 != 0 {
			nox_common_list_clear_425760(serverOptionsRules)
		}
	}
	pre := unsafe.Pointer(C.nox_xxx_guiServerOptionsProcPre_4585D0)
	root := nox_new_window_from_file(serverOptionsWindowFile(), pre)
	serverSharedWindows.options = root
	if root == nil {
		return 0
	}

	nox_draw_setTabWidth_43FE20(100)
	nox_window_setPos_46A9B0(root, int32(C.nox_win_width)-int32(root.width)-10, 0)
	nox_xxx_wndSetWindowProc_46B300(root, unsafe.Pointer(C.nox_xxx_guiServerOptionsProc_458590))
	nox_xxx_wndSetDrawFn_46B340(root, unsafe.Pointer(C.nox_xxx_windowServerOptionsDrawProc_458500))
	child := func(id int32) *nox_window { return nox_xxx_wndGetChildByID_46B0C0(root, id) }
	serverOptionsNameInput = child(10101)
	serverOptionsChildren.mapList = child(10114)
	serverOptionsChildren.mapPanel = child(10183)
	serverOptionsChildren.pane10197 = child(10197)
	serverOptionsChildren.pane10199 = child(10199)
	serverOptionsChildren.pane10150 = child(10150)
	serverOptionsChildren.lessonLimit = child(10134)
	serverOptionsChildren.timeLimit = child(10135)
	serverOptionsChildren.pane10153 = child(10153)
	nox_gui_winSetFunc96_46B070(child(10331), unsafe.Pointer(C.nox_xxx_options_457AA0))
	nox_gui_winSetFunc96_46B070(child(10333), unsafe.Pointer(C.nox_xxx_options_457B00))

	panels := [...]*nox_window{
		serverOptionsChildren.pane10150, serverOptionsChildren.playersTab, serverOptionsChildren.pane10153,
		serverOptionsChildren.mapPanel, serverOptionsChildren.pane10197, serverOptionsChildren.pane10199,
	}
	for _, panel := range panels {
		nox_xxx_wnd_46B280(panel, root)
	}
	for _, panel := range panels {
		nox_xxx_wndSetProc_46B2C0(panel, pre)
	}

	*memmap.PtrPtr(0x5D4594, 1046352) = unsafe.Pointer(nox_xxx_gLoadImg_42F970(internCStr("UITabs1")))
	serverOptionsTabImage2 = unsafe.Pointer(nox_xxx_gLoadImg_42F970(internCStr("UITabs2")))
	serverOptionsTabImage3 = unsafe.Pointer(nox_xxx_gLoadImg_42F970(internCStr("UITabs3")))
	slider := unsafe.Pointer(nox_xxx_gLoadImg_42F970(internCStr("UISlider")))
	sliderLit := unsafe.Pointer(nox_xxx_gLoadImg_42F970(internCStr("UISliderLit")))
	attachServerListScrollbar(serverOptionsChildren.mapPanel, serverOptionsChildren.mapList, 10182, 10180, 10181, slider, sliderLit)

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
		g2ServerHidden(serverOptionsChildren.pane10199, true)
		sub_46ACE0(root, 10145, 10146, 1)
	}
	if legacyWords.dword_587000_129656 != 0 {
		if sub_4D6F30() != 0 || nox_xxx_isQuest_4D6F50() != 0 {
			nox_server_parseCmdText_443C80(internWStr("execrul OTQuest.rul"), 1)
		} else if noxflags.HasEngine(noxflags.EngineNoRendering) {
			nox_server_parseCmdText_443C80(internWStr("execrul server.rul"), 1)
		}
	}
	legacyWords.dword_587000_129656 = 0
	return 1
}
