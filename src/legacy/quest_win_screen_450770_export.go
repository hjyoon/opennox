package legacy

/*
#include "quest_win_screen_450770.h"
#include "client__gui__window.h"
#include "common__strman.h"
#include "GAME1.h"
#include "GAME2.h"
extern nox_window* nox_wnd_briefing_831232;
extern uint32_t dword_5d4594_832476;
*/
import "C"

import (
	"encoding/binary"
	"unsafe"

	"github.com/opennox/opennox/v1/client/gui"
	"github.com/opennox/opennox/v1/common/memmap"
	"github.com/opennox/opennox/v1/server"
)

func questWinWidthSlot450770() *uint32 {
	return (*uint32)(unsafe.Pointer(&C.dword_5d4594_832476))
}

func questWinNativeHooks450770() questWinHooks450770[*byte, *server.Player, *gui.Window, unsafe.Pointer, *uint16] {
	rows := questWinRows450770()
	return questWinHooks450770[*byte, *server.Player, *gui.Window, unsafe.Pointer, *uint16]{
		clearRows:  func() { *rows = [6]questWinRow450770{} },
		storeTotal: func(v uint32) { *memmap.PtrUint32(0x5D4594, 832356) = v },
		storeStage: func(v uint32) { *memmap.PtrUint32(0x5D4594, 831228) = v },
		read16: func(p *byte, off uintptr) uint16 {
			return binary.LittleEndian.Uint16(unsafe.Slice((*byte)(unsafe.Add(unsafe.Pointer(p), off)), 2))
		},
		read32: func(p *byte, off uintptr) uint32 {
			return binary.LittleEndian.Uint32(unsafe.Slice((*byte)(unsafe.Add(unsafe.Pointer(p), off)), 4))
		},
		lookup: func(id uint16) *server.Player {
			return (*server.Player)(unsafe.Pointer(C.nox_common_playerInfoGetByID_417040(C.int(id))))
		},
		storePlayer: func(i int, p *server.Player) { rows[i].Player = p },
		store16: func(i int, f questWinField450770, v uint16) {
			switch f {
			case questWinKills450770:
				rows[i].Kills = v
			case questWinGenerators450770:
				rows[i].Generators = v
			case questWinSecrets450770:
				rows[i].Secrets = v
			case questWinCoopSecrets450770:
				rows[i].CoopSecrets = v
			}
		},
		storeScore: func(i int, v uint32) { rows[i].Score = v },
		sort:       questWinSortRows450770,
		loadWidth:  func() int32 { return int32(*questWinWidthSlot450770()) },
		storeWidth: func(v int32) { *questWinWidthSlot450770() = uint32(v) },
		child: func(id int32) *gui.Window {
			return (*gui.Window)(unsafe.Pointer(C.nox_xxx_wndGetChildByID_46B0C0(C.nox_wnd_briefing_831232, C.int(id))))
		},
		loadText: func(key, source string, line int32) *uint16 {
			return (*uint16)(unsafe.Pointer(C.nox_strman_loadString_40F1D0(internCStr(key), nil, internCStr(source), C.int(line))))
		},
		font: func(w *gui.Window) unsafe.Pointer { return w.DrawData().FontC() },
		measure: func(font unsafe.Pointer, text *uint16) int32 {
			var width C.int
			nox_xxx_drawGetStringSize_43F840(font, (*wchar2_t)(unsafe.Pointer(text)), &width, nil, 0)
			return int32(width)
		},
		lock: func(screen, mode int32, flags int8) int32 {
			return int32(C.nox_client_lockScreenBriefing_450160(C.int(screen), C.int(mode), C.char(flags)))
		},
	}
}

var questWinCall450770 = func(packet *byte) int32 {
	return questWinScreen450770(packet, questWinNativeHooks450770())
}

//export nox_client_questWinScreen_native_450770
func nox_client_questWinScreen_native_450770(packet *C.uchar) C.int {
	return C.int(questWinCall450770((*byte)(unsafe.Pointer(packet))))
}

func questWinCEntry450770(packet *byte) int32 {
	return int32(C.nox_xxx_clientQuestWinScreen_450770((*C.uchar)(unsafe.Pointer(packet))))
}
