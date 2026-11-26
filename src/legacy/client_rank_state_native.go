package legacy

/*
#include "GAME2_1.h"
#include "client__gui__guirank.h"

extern nox_window* dword_5d4594_1090048;
extern uint32_t dword_5d4594_1090120, dword_587000_145672;
*/
import "C"

import (
	"math"

	noxflags "github.com/opennox/opennox/v1/common/flags"
	"github.com/opennox/opennox/v1/common/memmap"
)

//export sub_46F060
func sub_46F060(win *C.nox_window, event C.int, arg1, arg2 C.uintptr_t) C.int { return 0 }

//export nox_xxx_Proc_46F070
func nox_xxx_Proc_46F070(win *C.nox_window, event C.int, arg1, arg2 C.uintptr_t) C.int { return 0 }

//export sub_46FAE0
func sub_46FAE0() {
	page := memmap.Int32(0x5D4594, 1088996)
	win := nox_rank_window_at(C.int(1090060 + 4*page))
	if win == nil || win.widget_data == nil || C.dword_587000_145672 == math.MaxUint32 {
		return
	}
	data := (*C.nox_scrollListBox_data)(win.widget_data)
	x, y := meterWindowPosition(win)
	h := int32(data.line_height)
	markerY := int32(y) + int32(C.dword_587000_145672)*h + h/2
	nox_client_drawSetColor_434460(int32(renderPalette.yellow))
	nox_xxx_drawPointMB_499B70(int32(x)+1, markerY, 3)
}

//export sub_46FE60
func sub_46FE60(id C.int) C.uchar {
	for i := range memmap.Uint8(0x5D4594, 1090116) {
		if clientRankTeamAt(i).ID == uint32(id) {
			return C.uchar(i)
		}
	}
	return 0
}

//export sub_46FEB0
func sub_46FEB0(index C.uchar) C.uchar {
	color := clientRankTeamAt(byte(index)).Color % 10
	return C.uchar(memmap.Uint8(0x587000, 145584+8*uintptr(color)))
}

//export sub_46FEE0
func sub_46FEE0() C.char {
	count := memmap.Uint8(0x5D4594, 1090117)
	localCode := uint32(ClientPlayerNetCode())
	for i := range count {
		if row := clientRankPlayerAt(i); row.NetCode == localCode {
			place := C.char(1)
			ascending := noxflags.HasGame(noxflags.GameModeElimination)
			for index := range count {
				if clientRankAhead(clientRankPlayerAt(index).Score, row.Score, ascending) {
					place++
				}
			}
			return place
		}
	}
	return 0
}

//export sub_46FF70
func sub_46FF70(score C.int) C.char {
	place := C.char(1)
	ascending := noxflags.HasGame(noxflags.GameModeElimination)
	for i := range memmap.Uint8(0x5D4594, 1090116) {
		if clientRankAhead(uint32(clientRankTeamAt(i).Score), uint32(score), ascending) {
			place++
		}
	}
	return place
}

func clientRankAhead(a, b uint32, ascending bool) bool {
	if ascending {
		return a < b
	}
	return a > b
}

//export sub_470580
func sub_470580() C.int {
	return C.int(bool2int(C.dword_5d4594_1090120 != 0 &&
		wndIsShown_nox_xxx_wndIsShown_46ACC0(C.dword_5d4594_1090048) != 1))
}

//export sub_4705B0
func sub_4705B0() {
	win := C.dword_5d4594_1090048
	if win == nil {
		return
	}
	if wndIsShown_nox_xxx_wndIsShown_46ACC0(win) != 0 {
		nox_window_set_hidden(win, 0)
	}
	C.dword_5d4594_1090120 = 0
	sub_4703F0()
}

//export sub_4705F0
func sub_4705F0(action, team C.char, code C.short) C.char {
	if team != 1 && team != 2 {
		return team
	}
	carrier := memmap.PtrUint16(0x5D4594, 1090128+2*uintptr(team-1))
	switch action {
	case 0, 2:
		*carrier = 0
	case 1:
		*carrier = uint16(code)
		if team == 1 {
			return C.char(code)
		}
	}
	return action
}

//export sub_470650
func sub_470650(action C.char, code C.short) C.char {
	carrier := memmap.PtrUint16(0x5D4594, 1090132)
	switch action {
	case 0, 1:
		*carrier = 0
	case 2, 4:
		*carrier = uint16(code)
		return C.char(code)
	}
	return action
}

//export sub_470680
func sub_470680() C.int {
	carriers := memmap.PtrT[[3]uint16](0x5D4594, 1090128)
	clear(carriers[:])
	return 0
}

//export sub_4706A0
func sub_4706A0() C.int {
	return C.int(bool2int(C.dword_5d4594_1090048 != nil && C.dword_5d4594_1090120 != 0))
}
