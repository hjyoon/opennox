package legacy

/*
#include "GAME2_3.h"
#include "client__gui__chathelp.h"
#include "client__gui__conntype.h"
#include "client__gui__servopts__guiserv.h"
#include "client__gui__window.h"


*/
import "C"

import (
	"slices"
	"unsafe"

	noxflags "github.com/opennox/opennox/v1/common/flags"
	"github.com/opennox/opennox/v1/common/memmap"
	"github.com/opennox/opennox/v1/legacy/common/alloc"
)

type gameSetupWindowSlots struct {
	help, connection *nox_window
}

var gameSetupWindows, _ = alloc.New(gameSetupWindowSlots{})

func showGameSetupDialog(win *nox_window) {
	sub_46B120(win, nil)
	nox_xxx_wndShowModalMB_46A8C0(win)
	sub_46C690(win)
	nox_xxx_windowFocus_46B500(win)
}

//export nox_xxx_cliShowHelpGui_49C560
func nox_xxx_cliShowHelpGui_49C560() {
	lang := nox_strman_get_lang_code()
	if nox_xxx_guiFontHeightMB_43F320(nil) > 10 {
		lang = 2
	}
	windowPath := (*C.char)(*memmap.PtrPtr(0x587000, 164512+4*uintptr(lang)))
	win := nox_new_window_from_file(windowPath, C.nox_xxx_wnd_49C760)
	storeWindowReference(&gameSetupWindows.help, win)
	if win == nil {
		return
	}
	showGameSetupDialog(win)
	nox_window_setPos_46A9B0(win, int32((int(clientDisplayStorage.width)-int(win.width))/2), int32((int(clientDisplayStorage.height)-int(win.height))/2))

	host := noxflags.HasGame(noxflags.GameHost)
	child := nox_xxx_wndGetChildByID_46B0C0(win, 4102)
	optionsKey := sub_42E8E0(45, 1)
	path := internCStr(`C:\NoxPost\src\client\Gui\chathelp.c`)
	key, helpLine, chatLine := "Sanchlp.wnd:ClientHelp", int32(94), int32(96)
	if host {
		key, helpLine, chatLine = "Sanchlp.wnd:Help", 85, 87
	}
	help := unsafe.Slice((*wchar2_t)(memmap.PtrSizeOff(0x5D4594, 1304656, 1024)), 512)
	format := nox_strman_loadString_40F1D0(internCStr(key), nil, path, helpLine)
	formatWideStrings(help, format, optionsKey, nil)

	chatKey := sub_42E8E0(8, 1)
	chat := unsafe.Slice((*wchar2_t)(memmap.PtrSizeOff(0x5D4594, 1304400, 256)), 128)
	format = nox_strman_loadString_40F1D0(internCStr("cdecode.c:KeyToChat"), nil, path, chatLine)
	formatWideStrings(chat, format, chatKey, nil)
	if n := slices.Index(help, 0); n+1 < len(help) {
		help[n] = ' '
		end := n + 1 + copy(help[n+1:len(help)-1], chat[:slices.Index(chat, 0)])
		help[end] = 0
	}
	nox_window_call_field_94_fnc(child, 16385, uintptr(unsafe.Pointer(&help[0])), 0, nil, 0)
	if host {
		if sub_459DA0() == 0 {
			C.nox_xxx_guiServerOptsLoad_457500()
		}
		sub_459D80(1)
	}
	if noxflags.HasGame(noxflags.GameModeQuest) || nox_xxx_isQuest_4D6F50() != 0 {
		sub_49C7A0()
	}
}

//export nox_xxx_wnd_49C760
func nox_xxx_wnd_49C760(window *nox_window, event C.int, control *nox_window, argument C.uintptr_t) C.int {
	if event == 16391 {
		id := nox_xxx_wndGetID_46B0A0(control)
		nox_xxx_clientPlaySoundSpecial_452D80(766, 100)
		if id == 4103 {
			sub_49C7A0()
		}
	}
	return 1
}

//export sub_49C7A0
func sub_49C7A0() C.int {
	win := gameSetupWindows.help
	if win == nil {
		return 0
	}
	checkbox := nox_xxx_wndGetChildByID_46B0C0(win, 4104)
	Set_nox_server_sanctuaryHelp_54276(int(^uint32(checkbox.draw_data.field_0) >> 2 & 1))
	nox_xxx_wnd_46C6E0(win)
	nox_xxx_wndClearCaptureMain_46ADE0(win)
	nox_xxx_windowDestroyMB_46C4E0(win)
	gameSetupWindows.help = nil
	nox_xxx_windowFocus_46B500(nil)
	if noxflags.HasGame(noxflags.GameHost) {
		return sub_459D80(0)
	}
	return 0
}

const gameConnectionTypes = 4

func gameConnectionTypeText(index int, line int32) *wchar2_t {
	key := (*C.char)(*memmap.PtrPtr(0x587000, 164928+4*uintptr(index)))
	path := internCStr(`C:\NoxPost\src\client\Gui\conntype.c`)
	return nox_strman_loadString_40F1D0(key, nil, path, line)
}

//export sub_49C820
func sub_49C820() C.int {
	win := nox_new_window_from_file(internCStr("conntype.wnd"), C.sub_49CA60)
	storeWindowReference(&gameSetupWindows.connection, win)
	if win == nil {
		return 0
	}
	showGameSetupDialog(win)
	sub_49C910()
	nox_window_setPos_46A9B0(win, int32(clientDisplayStorage.width)/2-int32(win.width)/2, int32(clientDisplayStorage.height)/2-int32(win.height)/2)
	C.nox_xxx_guiServerOptsLoad_457500()
	sub_459D80(1)

	list := nox_xxx_wndGetChildByID_46B0C0(win, 10352)
	if list == nil {
		return 0
	}
	for i := range gameConnectionTypes {
		text := gameConnectionTypeText(i, 158)
		nox_window_call_field_94_fnc(list, 16397, uintptr(unsafe.Pointer(text)), ^uintptr(0), nil, 0)
	}
	return C.int(nox_window_call_field_94_fnc(list, 16403, 0, 0, nil, 0))
}

//export sub_49C910
func sub_49C910() *nox_window {
	win := gameSetupWindows.connection
	if win == nil {
		return nil
	}
	list := nox_xxx_wndGetChildByID_46B0C0(win, 10352)
	if list == nil {
		return nil
	}
	font := list.draw_data.font
	list.height = C.intptr_t(5*(nox_xxx_guiFontHeightMB_43F320(font)+1) + 2)
	list.end_y = list.off_y + list.height

	var width C.int
	for i := range gameConnectionTypes {
		var textWidth C.int
		nox_xxx_drawGetStringSize_43F840(font, gameConnectionTypeText(i, 53), &textWidth, nil, 0)
		width = max(width, textWidth)
	}
	var titleWidth C.int
	nox_xxx_drawGetStringSize_43F840(font, &list.draw_data.text[0], &titleWidth, nil, 0)
	list.width = C.intptr_t(max(width+7, titleWidth))
	list.end_x = list.off_x + list.width
	win.off_x, win.off_y = list.off_x-40, list.off_y-40
	win.end_x, win.end_y = list.end_x+40, list.end_y+40
	win.width, win.height = list.width+80, list.height+80

	button := nox_xxx_wndGetChildByID_46B0C0(win, 10353)
	if button != nil {
		button.off_x, button.off_y = list.end_x-button.width, list.end_y+10
		button.end_x, button.end_y = button.off_x+button.width, button.off_y+button.height
	}
	return button
}

//export sub_49CA60
func sub_49CA60(window *nox_window, event C.int, control *nox_window, argument C.uintptr_t) C.int {
	if event != 16391 || control == nil {
		return 1
	}
	id := nox_xxx_wndGetID_46B0A0(control)
	nox_xxx_clientPlaySoundSpecial_452D80(766, 100)
	win := gameSetupWindows.connection
	if id != 10353 || win == nil {
		return 1
	}
	nox_xxx_wnd_46C6E0(win)
	nox_xxx_wndClearCaptureMain_46ADE0(win)
	if noxflags.HasGame(noxflags.GameModeChat) && Get_nox_server_sanctuaryHelp_54276() != 0 {
		nox_xxx_cliShowHelpGui_49C560()
	} else {
		sub_459D80(0)
	}
	if list := nox_xxx_wndGetChildByID_46B0C0(win, 10352); list != nil {
		index := C.int(nox_window_call_field_94_fnc(list, 16404, 0, 0, nil, 0))
		Set_nox_server_connectionType_3596(int(index + 1))
		nox_xxx_rateUpdate_40A6D0(sub_40A710(index + 1))
	}
	nox_xxx_windowDestroyMB_46C4E0(win)
	gameSetupWindows.connection = nil
	nox_xxx_windowFocus_46B500(nil)
	return 1
}
