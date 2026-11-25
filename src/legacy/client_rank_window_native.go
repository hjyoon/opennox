package legacy

/*
#include "GAME2_1.h"
#include "client__gui__guirank.h"

typedef const wchar2_t rank_const_text_t;
extern nox_window* dword_5d4594_1090048;
extern nox_window* dword_5d4594_1090100;
extern nox_window* dword_5d4594_1090108;
extern nox_window* dword_5d4594_1090112;
extern uint32_t dword_5d4594_1090040, dword_5d4594_1090044;
extern unsigned int dword_587000_145664, dword_587000_145668;
extern uint32_t dword_8531A0_2572;
extern uint32_t nox_color_white_2523948, nox_color_yellow_2589772;
extern int nox_win_width;
*/
import "C"

import (
	"fmt"
	"unsafe"

	noxflags "github.com/opennox/opennox/v1/common/flags"
	"github.com/opennox/opennox/v1/common/memmap"
	"github.com/opennox/opennox/v1/legacy/common/alloc"
)

const (
	clientRankFirstWindow = 1090052
	clientRankLastWindow  = 1090112
	clientRankTransparent = 0x80000000
)

var clientRankWindows [1 + (clientRankLastWindow-clientRankFirstWindow)/4]*C.nox_window
var clientRankClassNames [3]*wchar2_t

//export nox_rank_window_at
func nox_rank_window_at(offset C.int) *C.nox_window {
	if offset < clientRankFirstWindow || offset > clientRankLastWindow || (offset-clientRankFirstWindow)%4 != 0 {
		return nil
	}
	return clientRankWindows[(offset-clientRankFirstWindow)/4]
}

func clientRankStoreWindow(offset int, window *C.nox_window) {
	clientRankWindows[(offset-clientRankFirstWindow)/4] = window
}

//export nox_rank_class_name
func nox_rank_class_name(class C.uint) *C.rank_const_text_t {
	if class < C.uint(len(clientRankClassNames)) && clientRankClassNames[class] != nil {
		return (*C.rank_const_text_t)(clientRankClassNames[class])
	}
	return (*C.rank_const_text_t)(clientRankLocalized("InternalError", 1050))
}

//export sub_46F030
func sub_46F030() *C.wchar2_t {
	var name *wchar2_t
	file := internCStr(`C:\NoxPost\src\client\Gui\guirank.c`)
	for i := range clientRankClassNames {
		id := (*C.char)(*memmap.PtrPtr(0x587000, 145676+4*uintptr(i)))
		name = nox_strman_loadString_40F1D0(id, nil, file, 167)
		clientRankClassNames[i] = name
	}
	return name
}

//export nox_xxx_guiDrawRank_46E870
func nox_xxx_guiDrawRank_46E870() *C.nox_window {
	C.dword_587000_145668 = 6
	sub_46F030()
	height := C.int(nox_xxx_guiFontHeightMB_43F320(nil))
	columns := clientRankMeasureColumns()
	var width C.int
	for _, value := range columns {
		width += C.int(value)
	}
	C.dword_5d4594_1090040 = C.uint32_t(width)
	C.dword_5d4594_1090044 = C.uint32_t(439 - height)
	root := nox_window_new_go(nil, 1560, 0, height+40, 1, 1, nil)
	C.dword_5d4594_1090048 = root
	if root == nil {
		return nil
	}
	nox_window_set_all_funcs_go(root, unsafe.Pointer(C.sub_46F060), unsafe.Pointer(C.sub_46F080), nil)
	clientRankSetTransparentColors(&root.draw_data)
	draw := C.nox_window_data{style: 32, text_color: C.uint32_t(memmap.Uint32(0x85B3FC, 940))}
	clientRankSetTransparentColors(&draw)
	text := (*uint16)(memmap.PtrOff(0x5D4594, 1090136))
	drawText := (*[64]uint16)(unsafe.Pointer(&draw.text[0]))
	alloc.StrCopyZero16P(drawText[:], text)
	options := C.nox_scrollListBox_data{count: 64, line_height: C.uint16_t(height + 1), field_2: 1}
	listHeight := C.int(C.dword_5d4594_1090044) - 2*(height+1)
	for page := 0; page < 2; page++ {
		container := nox_gui_newScrollListBox_4A4310(root, 1088, C.int(page)*width, 3*height+1, width, listHeight, &draw, &options)
		clientRankStoreWindow(1090052+4*page, container)
		var x C.int
		for column, span := range columns {
			win := nox_gui_newScrollListBox_4A4310(container, 1088, x, 2*height, C.int(span), listHeight, &draw, &options)
			clientRankStoreWindow(1090060+8*column+4*page, win)
			x += C.int(span)
		}
		nox_xxx_wndSetProc_46B2C0_go(container, unsafe.Pointer(C.nox_xxx_Proc_46F070))
		for column := range columns {
			sub_46B120(nox_rank_window_at(C.int(1090060+8*column+4*page)), container)
		}
	}
	draw.style = 2048
	textData := C.nox_staticText_data{}
	labels := []struct {
		offset int
		y      C.int
		color  C.uint32_t
		id     string
		line   int32
	}{
		{1090100, height, C.nox_color_yellow_2589772, "yourrank", 1772},
		{1090112, 2 * height, C.nox_color_white_2523948, "WindowDir:Empty", 1782},
		{1090108, 3 * height, C.uint32_t(memmap.Uint32(0x85B3FC, 940)), "WindowDir:Empty", 1790},
		{1090104, 0, C.dword_8531A0_2572, "TeamPlayerRank", 1798},
	}
	for _, label := range labels {
		draw.text_color = label.color
		textData.text = clientRankLocalized(label.id, label.line)
		win := nox_gui_newStaticText_489300(root, 1088, 0, label.y, width, height+1, &draw, &textData)
		clientRankStoreWindow(label.offset, win)
		switch label.offset {
		case 1090100:
			C.dword_5d4594_1090100 = win
		case 1090108:
			C.dword_5d4594_1090108 = win
		case 1090112:
			C.dword_5d4594_1090112 = win
		}
	}
	C.dword_587000_145664 = 1
	return root
}

func clientRankSetTransparentColors(draw *C.nox_window_data) {
	draw.bg_color = clientRankTransparent
	draw.en_color = clientRankTransparent
	draw.hl_color = clientRankTransparent
	draw.dis_color = clientRankTransparent
	draw.sel_color = clientRankTransparent
}

func clientRankMeasureColumns() *[5]int32 {
	columns := memmap.PtrT[[5]int32](0x5D4594, 1084036)
	measure := func(text *wchar2_t) int32 {
		var width C.int
		nox_xxx_drawGetStringSize_43F840(nil, text, &width, nil, int32(C.nox_win_width))
		return int32(width)
	}
	columns[0] = 80
	columns[1] = max(18, measure(clientRankLocalized("Flag", 1641))) + 14
	columns[3] = max(measure(clientRankLocalized("Score", 1649)), measure(clientRankLocalized("HealthHeading", 1650))) + 7
	columns[4] = max(measure(clientRankLocalized("Ping", 1657)), measure(clientRankStockText(145972))) + 7
	columns[2] = max(
		measure(clientRankLocalized("Class", 1667)),
		measure(clientRankLocalized("Warrior", 1668)),
		measure(clientRankLocalized("Wizard", 1671)),
		measure(clientRankLocalized("Conjurer", 1674)),
		measure(clientRankLocalized("LivesHeading", 1677)),
	) + 7
	return columns
}

//export sub_46FC50
func sub_46FC50() C.char {
	win := C.dword_5d4594_1090108
	if win == nil {
		return 0
	}
	if sub_40A220() == 0 || !clientRankTimerMode() {
		return C.char(clientRankHideLabel(win))
	}
	if !clientRankTimerMode() {
		return 0
	}
	if wndIsShown_nox_xxx_wndIsShown_46ACC0(win) != 0 {
		nox_window_set_hidden(win, 0)
	}
	remaining := int32(sub_40A230())
	format := GoWString(clientRankLocalized("TimeRemaining", 352))
	buffer := memmap.PtrT[[32]uint16](0x5D4594, 1084068)
	alloc.StrCopyZero16(buffer[:], fmt.Sprintf(format, remaining/60000, remaining%60000/1000))
	return C.char(clientRankSetText(win, (*wchar2_t)(&buffer[0])))
}

func clientRankTimerMode() bool {
	return !noxflags.HasGame(noxflags.GameHost) || sub_40A300() != 0 || sub_40A180(C.short(noxflags.GetGame())) != 0
}

//export sub_46FD80
func sub_46FD80() C.int {
	win := C.dword_5d4594_1090112
	if win == nil {
		return 0
	}
	if noxflags.HasGame(noxflags.GameModeChat | noxflags.GameModeQuest) {
		return C.int(clientRankHideLabel(win))
	}
	if wndIsShown_nox_xxx_wndIsShown_46ACC0(win) != 0 {
		nox_window_set_hidden(win, 0)
	}
	format := GoWString(clientRankLocalized("LessonLimit", 390))
	buffer := memmap.PtrT[[32]uint16](0x5D4594, 1083972)
	alloc.StrCopyZero16(buffer[:], fmt.Sprintf(format, clientRankGameLimit()))
	return C.int(clientRankSetText(win, (*wchar2_t)(&buffer[0])))
}

func clientRankHideLabel(win *C.nox_window) int32 {
	if hidden := wndIsShown_nox_xxx_wndIsShown_46ACC0(win); hidden != 0 {
		return hidden
	}
	return nox_window_set_hidden(win, 1)
}
