package legacy

/*
#include "GAME2_1.h"
#include "client__gui__guiinv.h"
#include "client__gui__window.h"

*/
import "C"

import (
	"unsafe"

	"github.com/opennox/opennox/v1/common/memmap"
)

//export nox_client_invAlterWeapon_4672C0
func nox_client_invAlterWeapon_4672C0() {
	local := (*nox_drawable)(*memmap.PtrPtr(0x852978, 8))
	if local == nil || nox_xxx_guiCursor_477600() != 0 || sub_461160(1) != 0 {
		return
	}
	player := Get_dword_8531A0_2576()
	if player == nil || local.field_69 == 34 {
		return
	}
	cursor := (*C.int2)(memmap.PtrOff(0x5D4594, 1062572))
	bounds := (*C.int4)(memmap.PtrOff(0x587000, 136336))
	if nox_xxx_pointInRect_4281F0(cursor, bounds) == 1 {
		nox_xxx_cursorSetDraggedItem_477690(nil)
	}

	cell := clientInventoryAlternateCell()
	var alternate *nox_drawable
	if cell != nil {
		alternate = cell.field_0
	}
	if alternate != nil && nox_xxx_ammoCheck_415880(int32(alternate.field_27)) == 2 {
		if equipped := sub_461600(C.int(sub_415840(2))); equipped != nil {
			inventoryDequipForAlternate(equipped)
		}
		return
	}
	if equipped := inventoryEquippedWeapon(player.WeaponEquip); equipped != nil {
		inventoryDequipForAlternate(equipped)
		return
	}
	if alternate != nil {
		alternate.field_32 = cell.field_4
		nox_xxx_clientEquip_4623B0(alternate)
		nox_xxx_clientPlaySoundSpecial_452D80(895, 100)
	}
}

func inventoryDequipForAlternate(equipped *nox_drawable) {
	legacyPointers.inventoryDragItem = equipped
	nox_xxx_clientDequip_464B70(equipped)
	nox_xxx_clientPlaySoundSpecial_452D80(895, 100)
}

//export sub_467980
func sub_467980() C.int {
	for row := range inventoryRows {
		for column := range inventoryColumns {
			cell := inventoryGridCell(row, column)
			if cell.field_0 != nil {
				nox_xxx_spriteDelete_45A4B0(cell.field_0)
				cell.field_0 = nil
			}
			cell.field_140 = 0
			cell.field_132 = 0
			cell.field_136 = 0
		}
	}
	nox_client_inventory_set_dragged(nil)
	legacyWords.dword_5d4594_1049856 = 0
	sub_462740()
	inventoryState.mode = 0
	nox_xxx_clientSetAltWeapon_461550(nil)
	clientItemStorage.equipped_item_code = 0
	clear(inventoryStateNativeHeads()[:])
	legacyPointers.inventoryDragItem = nil
	legacyWords.dword_5d4594_1062496 = 0
	*memmap.PtrUint8(0x5D4594, 1062536) = 0
	*memmap.PtrUint32(0x5D4594, 1062540) = 0
	*memmap.PtrUint32(0x5D4594, 1062544) = 0
	*memmap.PtrUint32(0x5D4594, 1062548) = 0
	clientUIWords.player_gold = 0
	sub_472310()
	legacyWords.dword_587000_136184 = ^C.uint32_t(224) // -225, the closed tray offset.
	*memmap.PtrUint8(0x5D4594, 1049868) = 0
	*memmap.PtrUint8(0x5D4594, 1049869) = 0
	legacyWords.dword_5d4594_1062516 = 0
	legacyWords.dword_5d4594_1062520 = 0
	legacyWords.dword_5d4594_1062512 = 0

	scroll := inventoryWindows.slider
	if slider := game3Slider(scroll); slider != nil {
		file := internCStr("client_inventory_actions_native.go")
		nox_window_call_field_94_fnc(scroll, 16395, 0, 850, file, 0)
		nox_window_call_field_94_fnc(scroll, 16394, uintptr(uint32(slider.Max)-uint32(legacyWords.dword_5d4594_1062512)), 0, file, 0)
	}
	journal := inventoryWindows.journal
	nox_xxx_wndSetIcon_46AE60(journal, nil)
	sub_46AEC0(journal, inventoryImages[C.NOX_INV_IMG_JOURNAL_LIT])
	nox_xxx_wndSetID_46B080(journal, 9105)
	*memmap.PtrUint8(0x5D4594, 1049870) = 0
	stats := inventoryWindows.stats
	nox_xxx_wndSetIcon_46AE60(stats, inventoryImages[C.NOX_INV_IMG_STATS])
	sub_46AEC0(stats, inventoryImages[C.NOX_INV_IMG_STATS_LIT])
	nox_xxx_wndSetID_46B080(stats, 9107)
	return C.int(nox_window_set_hidden(inventoryWindows.weapon, 0))
}

//export sub_467CD0
func sub_467CD0() C.int {
	dragged := nox_client_inventory_get_dragged()
	if dragged != nil {
		column := C.int(legacyWords.dword_5d4594_1049796_inventory_click_column_index)
		row := C.int(legacyWords.dword_5d4594_1049800_inventory_click_row_index)
		if legacyWords.dword_5d4594_1049856 == 0 && sub_4649B0(dragged, column, row) == 0 {
			inventoryReinsertDraggedItem(dragged)
		}
		nox_client_inventory_set_dragged(nil)
		legacyWords.dword_5d4594_1049856 = 0
		nox_xxx_cursorResetDraggedItem_4776A0()
	}
	capture := nox_xxx_wndGetCaptureMain_46AE00()
	if nox_window_is_child(nox_xxx_wndGetHandle_4676A0(), capture) == 1 {
		nox_xxx_wndClearCaptureMain_46ADE0(capture)
	}
	return C.int(bool2int(dragged != nil))
}

func inventoryReinsertDraggedItem(dragged *nox_drawable) {
	data := drawableItemData(dragged)
	attrs := C.nox_modifier_attrs_t{
		modifiers: [4]unsafe.Pointer{data.Field_108, data.Field_109, data.Field_110, data.Field_111},
		field_16:  C.uint32_t(uint32(uint16(data.Field_112_0)) | uint32(uint16(data.Field_112_2))<<16),
	}
	nox_xxx_spritePickup_461660(C.int(dragged.field_32), C.int(dragged.field_27), &attrs)
	netCode := uint32(dragged.field_32)
	cell, _ := inventoryLookupNativeSearch(netCode)
	if cell == nil {
		return
	}
	clientInventoryMarkEquipped(cell, netCode)
}
