package legacy

/*
#include "client__gui__guiggovr.h"
#include "client__gui__window.h"
#include "common__strman.h"
#include "GAME2_1.h"
#include "noxstring.h"
extern nox_window* dword_5d4594_1303452;

static void quest_game_over_format_49B6E0(wchar2_t* dst, const wchar2_t* format, const wchar2_t* title, int32_t seconds) {
	nox_swprintf(dst, format, title, seconds);
}
*/
import "C"

import (
	"unsafe"

	"github.com/opennox/opennox/v1/client/gui"
	"github.com/opennox/opennox/v1/common/memmap"
	"github.com/opennox/opennox/v1/server"
)

func questGameOverTimerRootSlot49B6E0() **gui.Window {
	return (**gui.Window)(unsafe.Pointer(&C.dword_5d4594_1303452))
}

func questGameOverTimerBuffer49B6E0() *C.wchar2_t {
	return (*C.wchar2_t)(memmap.Ptr(0x5D4594 + 1301852))
}

func questGameOverTimerNativeHooks49B6E0() questGameOverTimerHooks49B6E0[*gui.Window, *server.Player, unsafe.Pointer] {
	return questGameOverTimerHooks49B6E0[*gui.Window, *server.Player, unsafe.Pointer]{
		root: func() *gui.Window { return *questGameOverTimerRootSlot49B6E0() },
		hidden: func(win *gui.Window) int32 {
			return int32(C.wndIsShown_nox_xxx_wndIsShown_46ACC0((*C.nox_window)(win.C())))
		},
		fps:    gameFPS,
		frame:  gameFrame,
		start:  func() uint32 { return *memmap.PtrUint32(0x5D4594, 1303456) },
		player: Get_dword_8531A0_2576,
		index:  func(player *server.Player) uint8 { return player.PlayerInd },
		copyEmpty: func() {
			// The source is an inline UTF-16 string, not a PE32 pointer slot.
			C.nox_wcscpy(questGameOverTimerBuffer49B6E0(), (*C.wchar2_t)(memmap.Ptr(0x5D4594+1303464)))
		},
		loadText: func(key, source string, line int32) unsafe.Pointer {
			return unsafe.Pointer(C.nox_strman_loadString_40F1D0(internCStr(key), nil, internCStr(source), C.int(line)))
		},
		format: func(title unsafe.Pointer, seconds int32) {
			C.quest_game_over_format_49B6E0(questGameOverTimerBuffer49B6E0(), internWStr(questGameOverTimerFormat49B6E0), (*C.wchar2_t)(title), C.int32_t(seconds))
		},
		child: func(win *gui.Window, id int32) *gui.Window {
			return (*gui.Window)(unsafe.Pointer(C.nox_xxx_wndGetChildByID_46B0C0((*C.nox_window)(win.C()), C.int(id))))
		},
		setText: func(win *gui.Window) int32 {
			// The original helper discards the GUI procedure response and returns 0.
			return int32(C.sub_46AEE0((*C.nox_window)(win.C()), questGameOverTimerBuffer49B6E0()))
		},
	}
}

var questGameOverTimerCall49B6E0 = func() int32 {
	return questGameOverTimer49B6E0(questGameOverTimerNativeHooks49B6E0())
}

//export nox_client_questGameOverTimer_native_49B6E0
func nox_client_questGameOverTimer_native_49B6E0() C.int {
	return C.int(questGameOverTimerCall49B6E0())
}

func questGameOverTimerCEntry49B6E0() int32 {
	return int32(C.sub_49B6E0())
}
