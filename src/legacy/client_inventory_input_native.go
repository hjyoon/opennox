package legacy

/*
#include "GAME2_1.h"
#include "client__gui__guiinv.h"
#include "client__gui__window.h"

typedef const nox_inventory_cell_t nox_const_inventory_cell;
extern nox_window* nox_inventory_window;
extern nox_window* nox_inventory_identify_window;
extern uint32_t dword_5d4594_1049796_inventory_click_column_index;
extern uint32_t dword_5d4594_1049800_inventory_click_row_index;
extern uint32_t dword_5d4594_1049856, dword_5d4594_1049864;
extern uint32_t dword_5d4594_1062512;
extern uintptr_t dword_5d4594_1062492;
*/
import "C"

import (
	"encoding/binary"
	"unsafe"

	"github.com/opennox/libs/noxnet/netmsg"
	"github.com/opennox/opennox/v1/common/memmap"
)

//export nox_xxx_XorEaxEaxSub_464BA0
func nox_xxx_XorEaxEaxSub_464BA0(win *C.nox_window, event C.int, a, b C.uintptr_t) C.int { return 0 }

//export nox_xxx_inventoryWndProc_464BB0
func nox_xxx_inventoryWndProc_464BB0(win *C.nox_window, event C.int, a, b C.uintptr_t) C.int {
	return C.int(bool2int(event != 8 && event != 12 && event != 16))
}

//export sub_464770
func sub_464770(win *C.nox_window, event C.int, packed, unused C.uintptr_t) C.int {
	if C.dword_5d4594_1049864 == 6 {
		return 1
	}
	switch event {
	case 5, 8:
		return 1
	case 6:
		if dragged := nox_client_inventory_get_dragged(); dragged != nil {
			inside := bool(C.nox_xxx_wndPointInWnd_46AAB0(C.nox_inventory_window, C.int(uint16(packed)), C.int(uint32(packed)>>16)))
			clearCapture := !inside || clientInventoryReleaseOnAlternate(dragged)
			clientInventoryFinishDrag(dragged)
			if !clearCapture {
				return 1
			}
		}
		nox_xxx_wndClearCaptureMain_46ADE0(C.nox_inventory_window)
		return 1
	case 7:
		if clientInventoryAlternateCell() != nil {
			nox_client_invAlterWeapon_4672C0()
		}
		return 1
	}
	return 0
}

func clientInventoryReleaseOnAlternate(dragged *nox_drawable) bool {
	weapon := dragged.flags28&0x1001000 != 0
	if C.dword_5d4594_1049856 != 0 {
		if weapon {
			if clientInventoryAlternateCell() != nil {
				nox_client_invAlterWeapon_4672C0()
			} else {
				C.dword_5d4594_1062492 = C.uintptr_t(uintptr(unsafe.Pointer(dragged)))
				nox_xxx_clientDequip_464B70(dragged)
			}
		}
		return true
	}
	column := C.int(C.dword_5d4594_1049796_inventory_click_column_index)
	row := C.int(C.dword_5d4594_1049800_inventory_click_row_index)
	if !weapon || sub_464B40(column, row) == 0 {
		sub_4649B0(dragged, column, row)
		return true
	}
	cell := inventoryGridCell(int(row), int(column))
	if cell.field_132 != 0 {
		sub_4649B0(dragged, column, row)
		return true
	}
	if nox_xxx_ammoCheck_415880(int32(dragged.field_27)) == 2 {
		ammo4 := sub_461600(C.int(sub_415840(4)))
		ammo8 := sub_461600(C.int(sub_415840(8)))
		if ammo4 == nil && ammo8 == nil {
			sub_4649B0(dragged, column, row)
			return false
		}
	}
	if alternate := clientInventoryAlternateCell(); alternate != nil {
		alternate.field_136 = 0
	}
	sub_4649B0(dragged, column, row)
	nox_xxx_clientSetAltWeapon_461550(cell)
	cell.field_136 = 1
	return true
}

func clientInventoryFinishDrag(dragged *nox_drawable) {
	nox_xxx_cursorResetDraggedItem_4776A0()
	if C.dword_5d4594_1049856 == 0 {
		nox_xxx_spriteDelete_45A4B0(dragged)
	}
	nox_client_inventory_set_dragged(nil)
	C.dword_5d4594_1049856 = 0
}

//export nox_xxx_trade_4657B0
func nox_xxx_trade_4657B0(code C.short) C.int {
	return shopItemPacket(30, uint16(code))
}

//export nox_xxx_clientTrade_465870
func nox_xxx_clientTrade_465870(code C.short) C.int {
	return shopItemPacket(28, uint16(code))
}

//export nox_xxx_clientTradeMB_4657E0
func nox_xxx_clientTradeMB_4657E0(point *C.uint32_t) C.char {
	pos := (*C.int2)(unsafe.Pointer(point))
	if nox_xxx_pointInRect_4281F0(pos, (*C.int4)(memmap.PtrOff(0x587000, 136352))) == 0 {
		return 0
	}
	column := (uint32(pos.field_0) - 314) / 50
	row := (uint32(pos.field_4) + uint32(C.dword_5d4594_1062512) - 13) / 50
	if column >= uint32(inventoryColumns) || row >= uint32(inventoryRows) {
		return 0
	}
	cell := inventoryGridCell(int(row), int(column))
	if cell.field_140 == 0 {
		return 0
	}
	code := nox_inventory_cell_last_net_code_4657E0((*C.nox_const_inventory_cell)(cell))
	return C.char(nox_xxx_clientTrade_465870(C.short(code)))
}

//export sub_4658A0
func sub_4658A0(win *C.nox_window, point *C.int2) {
	if nox_client_inventory_animation_state() != 2 {
		return
	}
	switch {
	case nox_xxx_pointInRect_4281F0(point, (*C.int4)(memmap.PtrOff(0x587000, 136336))) != 0:
		var dragged *nox_drawable
		if slot := sub_465990((*C.uint32_t)(unsafe.Pointer(point))); slot >= 0 {
			dragged = (*nox_drawable)(inventoryStateNativeHeads()[slot].C())
		}
		nox_client_inventory_set_dragged(dragged)
		C.dword_5d4594_1049856 = C.uint32_t(bool2int(dragged != nil))
	case nox_xxx_pointInRect_4281F0(point, (*C.int4)(memmap.PtrOff(0x587000, 136368))) != 0:
		C.dword_5d4594_1049856 = 0
		if (point.field_4-13)/50 == 2 {
			nox_client_toggleMap_473610()
		}
	default:
		C.dword_5d4594_1049856 = 0
		if nox_xxx_pointInRect_4281F0(point, (*C.int4)(memmap.PtrOff(0x587000, 136352))) != 0 {
			column := (point.field_0 - 314) / 50
			row := (point.field_4 + C.int(C.dword_5d4594_1062512) - 13) / 50
			clientInventorySetClickCell(column, row)
			if sub_464B40(column, row) != 0 {
				nox_xxx_cliInventorySpriteUpd_465A30()
			}
		}
	}
}

func clientInventorySetClickCell(column, row C.int) {
	C.dword_5d4594_1049796_inventory_click_column_index = C.uint32_t(column)
	C.dword_5d4594_1049800_inventory_click_row_index = C.uint32_t(row)
}

//export nox_xxx_clientDrop_465BE0
func nox_xxx_clientDrop_465BE0(position *C.int2) C.int {
	dragged := nox_client_inventory_get_dragged()
	if dragged == nil {
		return 0
	}
	packet := [7]byte{byte(netmsg.MSG_TRY_DROP)}
	binary.LittleEndian.PutUint16(packet[1:], uint16(nox_xxx_netGetUnitCodeCli_578B00(dragged)))
	binary.LittleEndian.PutUint16(packet[3:], uint16(position.field_0))
	binary.LittleEndian.PutUint16(packet[5:], uint16(position.field_4))
	return clientInventorySendPacket(packet[:])
}

//export nox_xxx_clientKeyEquip_465C30
func nox_xxx_clientKeyEquip_465C30(column, row C.int) C.int {
	clientInventorySetClickCell(column, row)
	nox_xxx_cliInventorySpriteUpd_465A30()
	dragged := nox_client_inventory_get_dragged()
	if dragged == nil {
		return 0
	}
	nox_xxx_clientEquip_4623B0(dragged)
	return sub_4649B0(dragged, column, row)
}

//export nox_xxx_clientUse_465C70
func nox_xxx_clientUse_465C70(drawable *C.nox_drawable) {
	if drawable != nil {
		clientInventorySendItem(netmsg.MSG_TRY_USE, drawable)
	}
}

//export sub_465CA0
func sub_465CA0() C.int {
	nox_window_set_hidden(C.nox_inventory_identify_window, 0)
	C.dword_5d4594_1049864 = 5
	nox_client_setCursorType_477610(6)
	return C.int(nox_xxx_wndSetCaptureMain_46ADC0(C.nox_inventory_window))
}

//export sub_465CD0
func sub_465CD0(position *C.int2, code, thingType, amount, extra C.uint32_t) {
	if amount == 0 {
		return
	}
	var projected C.int2
	sub_473970(position, &projected)
	cell, _ := inventoryLookupNativeSearch(uint32(code))
	if cell == nil || cell.field_0 == nil {
		return
	}
	codes := inventoryCellCodes(cell)
	for i := range min(uint32(amount), uint32(len(codes))) {
		dragged := cell.field_0
		nox_client_inventory_set_dragged(dragged)
		dragged.field_32 = codes[i]
		if sub_4C12C0() == 0 {
			nox_xxx_clientDrop_465BE0(&projected)
		}
		nox_client_inventory_set_dragged(nil)
	}
}
