package legacy

/*
#include "client__gui__servopts__guiserv.h"
#include "client__gui__window.h"
#include "noxstring.h"

extern nox_window* dword_5d4594_1046492;
*/
import "C"

import (
	"sync"
	"unsafe"

	noxflags "github.com/opennox/opennox/v1/common/flags"
)

const serverOptionsStringsFile = "C:\\NoxPost\\src\\client\\Gui\\ServOpts\\guiserv.c"

type serverOptionMode struct {
	key    string
	flag   noxflags.GameFlag
	hidden bool
	title  *C.wchar2_t
}

var serverOptionModes = [...]serverOptionMode{
	{key: "CTF", flag: noxflags.GameModeCTF},
	{key: "Arena", flag: noxflags.GameModeArena},
	{key: "Highlander", flag: noxflags.GameModeElimination},
	{key: "KotR", flag: noxflags.GameModeKOTR},
	{key: "Flagball", flag: noxflags.GameModeFlagBall},
	{key: "Quest", flag: noxflags.GameModeQuest, hidden: true},
	{key: "Noxworld.c:Chat", flag: noxflags.GameModeChat},
}

var loadServerOptionModes sync.Once

//export nox_gui_gamemode_load_457410
func nox_gui_gamemode_load_457410() {
	loadServerOptionModes.Do(func() {
		for i := range serverOptionModes {
			mode := &serverOptionModes[i]
			mode.title = g2ServerLoad(mode.key, serverOptionsStringsFile, 212)
		}
	})
}

//export nox_xxx_guiServerOptionsGetGametypeName_4573C0
func nox_xxx_guiServerOptionsGetGametypeName_4573C0(mode C.short) *C.wchar2_t {
	nox_gui_gamemode_load_457410()
	flag := noxflags.GameFlag(uint16(mode)) & noxflags.GameModeMask
	for _, choice := range serverOptionModes {
		if choice.flag == flag {
			return choice.title
		}
	}
	return serverOptionModes[1].title
}

//export sub_457A10
func sub_457A10() {
	win := nox_xxx_wndGetChildByID_46B0C0(C.dword_5d4594_1046492, 10120)
	if win == nil {
		return
	}
	nox_gui_gamemode_load_457410()
	font := win.draw_data.font
	var widest C.int
	count := 0
	for _, mode := range serverOptionModes {
		if mode.hidden {
			continue
		}
		count++
		var width C.int
		nox_xxx_drawGetStringSize_43F840(font, mode.title, &width, nil, 0)
		widest = max(widest, width)
	}
	height := C.intptr_t(count*(int(nox_xxx_guiFontHeightMB_43F320(font))+1) + 2)
	width := C.intptr_t(widest) + 7
	win.end_y = win.off_y + height
	win.height = height
	win.width = width
	win.off_x = win.end_x - width
}

//export sub_459650
func sub_459650(title *C.wchar2_t) C.int {
	nox_gui_gamemode_load_457410()
	for _, mode := range serverOptionModes {
		if C.nox_wcscmp(mode.title, title) == 0 {
			return C.int(mode.flag)
		}
	}
	return 0
}

//export sub_459C10
func sub_459C10() C.int {
	win := nox_xxx_wndGetChildByID_46B0C0(C.dword_5d4594_1046492, 10119)
	if win == nil {
		return 0
	}
	return sub_459650(&win.draw_data.text[0])
}

//export nox_xxx_windowServerOptionsFillGametypeList_4596A0
func nox_xxx_windowServerOptionsFillGametypeList_4596A0() {
	win := nox_xxx_wndGetChildByID_46B0C0(C.dword_5d4594_1046492, 10120)
	if win == nil {
		return
	}
	nox_gui_gamemode_load_457410()
	g2ServerEvent(win, 16399, 0, 0)
	for _, mode := range serverOptionModes {
		if !mode.hidden {
			g2ServerText(win, 16397, unsafe.Pointer(mode.title), ^uintptr(0))
		}
	}
}

//export nox_xxx_options_457AA0
func nox_xxx_options_457AA0(_ C.int, flags *C.uint8_t) C.int {
	if *flags&4 != 0 {
		nox_xxx_cursorSetTooltip_4776B0(g2ServerLoad("AutoAssignOnTT", serverOptionsStringsFile, 446))
	} else {
		nox_xxx_cursorSetTooltip_4776B0(g2ServerLoad("AutoAssignOffTT", serverOptionsStringsFile, 450))
	}
	return 1
}

//export nox_xxx_options_457B00
func nox_xxx_options_457B00(_ C.int, flags *C.uint8_t) C.int {
	if *flags&4 != 0 {
		nox_xxx_cursorSetTooltip_4776B0(g2ServerLoad("TeamDamageOnTT", serverOptionsStringsFile, 465))
	} else {
		nox_xxx_cursorSetTooltip_4776B0(g2ServerLoad("TeamDamageOffTT", serverOptionsStringsFile, 469))
	}
	return 1
}
