package legacy

/*
#include "GAME2_1.h"
#include "GAME2_2.h"
#include "GAME3_1.h"
#include "client__gui__window.h"
#include "noxstring.h"

static inline void gui_trade_format_initial_total(wchar2_t* out, const wchar2_t* format) {
	nox_swprintf(out, format, 0);
}
*/
import "C"

import (
	"encoding/binary"
	"unsafe"

	"github.com/opennox/opennox/v1/common/memmap"
)

func tradeActionPacket(command byte) C.int {
	packet := [2]byte{0xc9, command}
	return C.int(nox_netlist_addToMsgListCli_40EBC0(31, 0, (*C.uchar)(&packet[0]), int32(len(packet))))
}

//export nox_xxx_func_4C0610
func nox_xxx_func_4C0610() C.int { return tradeActionPacket(14) }

//export nox_xxx_clientTrade_4C0CE0
func nox_xxx_clientTrade_4C0CE0() C.int { return tradeActionPacket(17) }

//export nox_xxx_clientTrade_0_4C08E0
func nox_xxx_clientTrade_0_4C08E0(drawable *nox_drawable) C.int {
	packet := [4]byte{0xc9, 16}
	code := objectFieldDrawableCode578B00(asDrawable(drawable))
	binary.LittleEndian.PutUint16(packet[2:], uint16(code))
	return C.int(nox_netlist_addToMsgListCli_40EBC0(31, 0, (*C.uchar)(&packet[0]), int32(len(packet))))
}

//export sub_4C0630
func sub_4C0630(win *nox_window, event C.int, packed, unused C.uintptr_t) C.int {
	pos := C.int2{field_0: C.int(uint16(packed)), field_4: C.int(uint32(packed) >> 16)}
	switch event {
	case 5:
		if C.nox_xxx_wndPointInWnd_46AAB0(tradeChild(3704), pos.field_0, pos.field_4) {
			if slot := sub_4C0910(&pos); slot != nil && slot.count != 0 && slot.drawable != nil {
				startTradeDrag(slot, pos)
			}
		}
		return 1
	case 6, 7:
		if C.nox_xxx_wndGetCaptureMain_46AE00() == tradeRoot() {
			C.nox_xxx_wndClearCaptureMain_46ADE0(tradeRoot())
		}
		if clientItemStorage.trade_dragged != nil && clientItemStorage.trade_source != nil {
			finishTradeDrag(pos)
		}
		return 1
	default:
		return 0
	}
}

func startTradeDrag(slot *tradeClientSlot, pos C.int2) {
	C.nox_xxx_wndSetCaptureMain_46ADC0(tradeRoot())
	nox_xxx_setKeybTimeout_4160D0(2)
	*(*C.int2)(g2ServerMem(1319276)) = pos
	drawable := slot.drawable
	clientItemStorage.trade_dragged = drawable
	drawable.field_32 = slot.item_ids[slot.count-1]
	slot.item_ids[slot.count-1] = 0
	nox_xxx_cursorSetDraggedItem_477690(drawable)
	slot.count--
	if slot.count == 0 {
		slot.drawable = nil
	}
	clientItemStorage.trade_source = slot
	*g2ServerWord(1320304) = 0
	nox_xxx_clientPlaySoundSpecial_452D80(791, 100)
}

func tradeShortClick(pos C.int2, key C.uchar) bool {
	dx := memmap.Int32(0x5D4594, 1319276) - int32(pos.field_0)
	dy := memmap.Int32(0x5D4594, 1319280) - int32(pos.field_4)
	return !bool(nox_xxx_checkKeybTimeout_4160F0(key, C.uint(gameFPS()/3))) && dx*dx+dy*dy < 100
}

func finishTradeDrag(pos C.int2) {
	drawable := clientItemStorage.trade_dragged
	switch *g2ServerWord(1320304) {
	case 0:
		if !bool(C.nox_xxx_wndPointInWnd_46AAB0(tradeChild(3704), pos.field_0, pos.field_4)) || tradeShortClick(pos, 2) {
			nox_xxx_clientTrade_0_4C08E0(drawable)
		}
	case 1:
		if !bool(C.nox_xxx_wndPointInWnd_46AAB0(tradeChild(3705), pos.field_0, pos.field_4)) || tradeShortClick(pos, 3) {
			nox_xxx_clientTrade_0_4C08E0(drawable)
		}
	}
	slot := clientItemStorage.trade_source
	drawable = clientItemStorage.trade_dragged
	slot.item_ids[slot.count] = drawable.field_32
	slot.count++
	if slot.count == 1 {
		slot.drawable = drawable
	}
	nox_xxx_cursorResetDraggedItem_4776A0()
	clientItemStorage.trade_dragged, clientItemStorage.trade_source = nil, nil
}

//export sub_4C0C90
func sub_4C0C90(win *nox_window, event C.int, arg, unused C.uintptr_t) C.int {
	if event == 16391 {
		id := nox_xxx_wndGetID_46B0A0((*nox_window)(unsafe.Pointer(uintptr(arg))))
		nox_xxx_clientPlaySoundSpecial_452D80(766, 100)
		switch id {
		case 3708:
			nox_xxx_clientTrade_4C0CE0()
		case 3710:
			nox_xxx_func_4C0610()
		}
	}
	return 0
}

//export nox_xxx_closeP2PTradeWnd_4C12A0
func nox_xxx_closeP2PTradeWnd_4C12A0() C.int {
	nox_xxx_windowDestroyMB_46C4E0(tradeRoot())
	tradeWindow, tradeActive = nil, 0
	return 0
}

//export nox_xxx_showP2PTradeWnd_4C12D0
func nox_xxx_showP2PTradeWnd_4C12D0() C.int {
	if tradeActive == 0 {
		return 0
	}
	return C.int(nox_xxx_wndShowModalMB_46A8C0(tradeRoot()))
}

func tradeSetNames(player *nox_playerInfo) C.int {
	g2ServerText(tradeChild(3702), 16385, unsafe.Pointer(&player.name_final[0]), 0)
	return C.int(g2ServerText(tradeChild(3703), 16385, g2ServerMem(1319844), 0))
}

//export nox_xxx_netP2PStartTrade_4C1320
func nox_xxx_netP2PStartTrade_4C1320(data *C.nox_trade_const_u8_t) C.int {
	player := (*nox_playerInfo)(unsafe.Pointer(Get_dword_8531A0_2576()))
	if player == nil || data == nil || tradeActive == 1 {
		return 0
	}
	tradeActive = 1
	sub_4C1410()
	copyWideString((*wchar2_t)(g2ServerMem(1319844)), (*wchar2_t)(unsafe.Add(unsafe.Pointer(data), 2)))
	g2ServerEnabled(tradeRoot(), true)
	nox_xxx_showP2PTradeWnd_4C12D0()
	var width, height C.int
	C.nox_window_get_size(tradeRoot(), &width, &height)
	nox_window_setPos_46A9B0(tradeRoot(), 198, 193)
	tradeSetNames(player)
	C.sub_467BB0()
	return 1
}

//export sub_4C1410
func sub_4C1410() C.int {
	for _, offset := range [...]uintptr{1320240, 1320868, 1320100} {
		*memmap.PtrUint16(0x5D4594, offset) = 0
	}
	g2ServerText(tradeChild(3702), 16385, g2ServerMem(1320976), 0)
	g2ServerText(tradeChild(3703), 16385, g2ServerMem(1320980), 0)
	nox_gui_trade_slots_reset()
	var text [64]uint16
	ptr := unsafe.Pointer(&text[0])
	C.gui_trade_format_initial_total((*wchar2_t)(ptr), (*wchar2_t)(g2ServerMem(1319972)))
	g2ServerText(tradeChild(3711), 16385, ptr, 0)
	g2ServerText(tradeChild(3712), 16385, ptr, 0)
	result := g2ServerText(tradeChild(3713), 16385, ptr, 0)
	tradeLocalAccepted, tradeRemoteAccepted = 0, 0
	clientItemStorage.trade_dragged, clientItemStorage.trade_source = nil, nil
	clientItemStorage.trade_local, clientItemStorage.trade_remote = nil, nil
	return C.int(result)
}

//export sub_4C1590
func sub_4C1590() C.int {
	if tradeActive == 0 {
		return 0
	}
	sub_4C1410()
	g2ServerHidden(tradeRoot(), true)
	g2ServerEnabled(tradeRoot(), false)
	tradeActive = 0
	return C.sub_467C10()
}

//export nox_xxx_prepareP2PTrade_4C1BF0
func nox_xxx_prepareP2PTrade_4C1BF0() C.int {
	result := tradeActive
	player := (*nox_playerInfo)(unsafe.Pointer(Get_dword_8531A0_2576()))
	if result != 0 && player != nil {
		sub_4C1410()
		return tradeSetNames(player)
	}
	return C.int(result)
}
