package legacy

/*
#include "GAME2_1.h"
#include "client__gui__guirank.h"

typedef const wchar2_t rank_const_text_t;
extern unsigned int dword_587000_145664;
extern uint32_t dword_587000_145672, dword_5d4594_1090120;
extern uint32_t nox_color_white_2523948, nox_color_yellow_2589772;
*/
import "C"

import (
	"math"
	"strconv"
	"unsafe"

	"github.com/opennox/opennox/v1/common/memmap"
	"github.com/opennox/opennox/v1/legacy/common/alloc"
)

func clientRankLocalized(id string, line int32) *wchar2_t {
	return nox_strman_loadString_40F1D0(internCStr(id), nil, internCStr(`C:\NoxPost\src\client\Gui\guirank.c`), line)
}

func clientRankStockText(offset uintptr) *wchar2_t {
	return (*wchar2_t)(memmap.PtrOff(0x587000, offset))
}

func clientRankWindow(offset int) *C.nox_window {
	return nox_rank_window_at(C.int(offset))
}

func clientRankList(offset int) *C.nox_scrollListBox_data {
	win := clientRankWindow(offset)
	if win == nil {
		return nil
	}
	return (*C.nox_scrollListBox_data)(win.widget_data)
}

//export sub_46DC60
func sub_46DC60(win *C.nox_window, color C.uchar, text *C.rank_const_text_t) C.int {
	if text == nil {
		text = (*C.rank_const_text_t)(clientRankLocalized("InternalError", 1050))
	}
	return sub_46DC00(win, color, text)
}

// Used by the fixed %s, %d and %d%% formats in the rank columns.
func clientRankWriteWideCell(offset int, color C.uchar, text *wchar2_t) C.int {
	buf := memmap.PtrT[[512]uint16](0x5D4594, 1089000)
	alloc.StrCopyZero16P(buf[:], (*uint16)(unsafe.Pointer(text)))
	return sub_46DC00(clientRankWindow(offset), color, (*C.rank_const_text_t)(&buf[0]))
}

func clientRankWriteNumberCell(offset int, color C.uchar, number int32, suffix string) C.int {
	buf := memmap.PtrT[[512]uint16](0x5D4594, 1089000)
	alloc.StrCopyZero16(buf[:], strconv.FormatInt(int64(number), 10)+suffix)
	return sub_46DC00(clientRankWindow(offset), color, (*C.rank_const_text_t)(&buf[0]))
}

//export sub_46F8F0
func sub_46F8F0(page, emptyRows C.int) C.int {
	start := 1090060 + 4*int(page)
	blank := [5]uintptr{147124, 147128, 147132, 147136, 147140}
	for row := C.int(0); row < emptyRows; row++ {
		for column, offset := range blank {
			sub_46DC60(clientRankWindow(start+8*column), 9, (*C.rank_const_text_t)(clientRankStockText(offset)))
		}
	}
	sub_46DC60(clientRankWindow(start), 9, (*C.rank_const_text_t)(clientRankLocalized("player", 189)))
	sub_46DC60(clientRankWindow(start+8), 9, (*C.rank_const_text_t)(clientRankStockText(147188)))
	mode := C.dword_5d4594_1090120
	classID, classLine := "class", int32(195)
	scoreID, scoreLine := "score", int32(203)
	lastID, lastLine := "ping", int32(209)
	if mode == 1 {
		classID, classLine = "LivesHeading", 193
		scoreID, scoreLine = "HealthHeading", 201
		lastID, lastLine = "class", 207
	} else if mode == 5 {
		scoreID, scoreLine = "rank", 199
	}
	sub_46DC60(clientRankWindow(start+16), 9, (*C.rank_const_text_t)(clientRankLocalized(classID, classLine)))
	sub_46DC60(clientRankWindow(start+24), 9, (*C.rank_const_text_t)(clientRankLocalized(scoreID, scoreLine)))
	return sub_46DC60(clientRankWindow(start+32), 9, (*C.rank_const_text_t)(clientRankLocalized(lastID, lastLine)))
}

//export sub_46FB50
func sub_46FB50(objective C.int, color *C.uint8_t) *C.wchar2_t {
	id, line := "", int32(0)
	markerColor := C.uint8_t(4)
	switch objective {
	case 1:
		id, line = "King", 250
	case 2:
		id, line, markerColor = "Flag", 256, 7
	case 3:
		id, line, markerColor = "Flag", 262, 13
	case 4:
		id, line = "Ball", 244
	default:
		*color = 4
		return clientRankStockText(147724)
	}
	name := (*uint16)(unsafe.Pointer(clientRankLocalized(id, line)))
	output := (*uint16)(memmap.PtrOff(0x5D4594, 1090024))
	buf := unsafe.Slice(output, alloc.StrLen(name)+2)
	buf[0] = '<'
	alloc.StrCopyZero16P(buf[1:], name)
	*color = markerColor
	return (*wchar2_t)(output)
}

//export sub_46FFD0
func sub_46FFD0() C.uchar {
	sub_46DB80()
	sub_46F8F0(0, 0)
	count := memmap.Uint8(0x5D4594, 1090117)
	displayRow := int32(1)
	local := uint32(ClientPlayerNetCode())
	for i := range count {
		row := clientRankPlayerAt(i)
		player := nox_common_playerInfoGetByID_417040(int32(row.NetCode))
		if player == nil || player.field_4792 == 0 {
			continue
		}
		color := C.uchar(9)
		if row.Status&1 == 0 {
			color = 4
			if row.TeamID != -1 {
				color = sub_46FEB0(sub_46FE60(C.int(row.TeamID)))
			}
		}
		if row.NetCode == local {
			C.dword_587000_145672 = math.MaxUint32
			if data := clientRankList(1090060); data != nil {
				C.dword_587000_145672 = C.uint32_t(data.field_11_1)
			}
			*memmap.PtrUint32(0x5D4594, 1088996) = 0
		}
		clientRankWriteWideCell(1090060, color, (*wchar2_t)(&row.Name[0]))
		clientRankWriteNumberCell(1090076, color, int32(uint8(player.tail_padding[0])), "")
		health := int32(uint8(player.field_2282))
		if row.NetCode == local {
			health = 0
			if total := sub_470CD0(); total > 0 {
				health = int32(nox_float2int(C.float(float64(sub_470CC0()) / float64(total) * 100)))
			}
		}
		healthColor := C.uchar(4)
		switch {
		case health <= 25:
			healthColor = 6
		case health <= 50:
			healthColor = 15
		}
		clientRankWriteNumberCell(1090084, healthColor, health, "%")
		clientRankWriteWideCell(1090092, color, nox_rank_class_name(C.uint(row.Class)))

		var markerColor C.uint8_t
		marker := sub_46FB50(C.int(row.Objective), &markerColor)
		sub_46DC60(clientRankWindow(1090068), C.uchar(markerColor), (*C.rank_const_text_t)(marker))
		clientRankDrawQuestKeys(displayRow, Nox_client_playerQuestKeysReceived(asPlayerS(player)))
		displayRow++
	}
	C.dword_587000_145664 = 1
	return C.uchar(count)
}

func clientRankDrawQuestKeys(row int32, keys [2]byte) {
	list := clientRankWindow(1090068)
	if list == nil || list.widget_data == nil || keys == [2]byte{} {
		return
	}
	data := (*C.nox_scrollListBox_data)(list.widget_data)
	x, y := meterWindowPosition(list)
	height := int32(data.line_height)
	markerY := int32(y) + height/2 + height*row - 1
	for index, count := range keys {
		if count == 0 {
			continue
		}
		markerX := int32(x) + 5
		color := int32(C.nox_color_white_2523948)
		if index == 1 {
			color = int32(C.nox_color_yellow_2589772)
			if keys[0] == 1 {
				markerX += 15
			}
		}
		nox_client_drawSetColor_434460(color)
		nox_video_drawCircleColored_4C3270(markerX, markerY, 2, color)
		nox_client_drawAddPoint_49F500(markerX+2, markerY)
		nox_client_drawAddPoint_49F500(markerX+9, markerY)
		nox_client_drawLineFromPoints_49E4B0()
		nox_client_drawAddPoint_49F500(markerX+9, markerY)
		nox_client_drawAddPoint_49F500(markerX+9, markerY+3)
		nox_client_drawLineFromPoints_49E4B0()
		nox_client_drawAddPoint_49F500(markerX+7, markerY)
		nox_client_drawAddPoint_49F500(markerX+7, markerY+2)
		nox_client_drawLineFromPoints_49E4B0()
	}
}
