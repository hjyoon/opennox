package legacy

/*
#include "GAME2_1.h"
#include "GAME2_2.h"
#include "client__gui__guiinv.h"
#include "client__gui__guimsg.h"

*/
import "C"

import (
	"unsafe"

	"github.com/opennox/opennox/v1/common/memmap"
)

//export sub_4649B0
func sub_4649B0(drawable *C.nox_drawable, column, row C.int) C.int {
	if drawable == nil || sub_464B40(column, row) == 0 {
		return 0
	}
	cell := inventoryGridCell(int(row), int(column))
	count := int(cell.field_140)
	if count != 0 && (drawable.flags28&0x4000000 != 0 || cell.field_0 == nil || cell.field_0.field_27 != drawable.field_27) {
		return 0
	}
	if count >= inventoryStackMax {
		return 0
	}
	if count == 0 {
		cell.field_0 = clientInventoryCloneDrawable(drawable)
		if cell.field_0 == nil {
			nox_xxx_printCentered_445490(clientInventoryString("DrawablesExhausted", 898))
			return 0
		}
	}
	inventoryCellCodes(cell)[count] = drawable.field_32
	cell.field_140++
	clientInventoryMarkEquipped(cell, uint32(drawable.field_32))
	return 1
}

func clientInventoryMarkEquipped(cell *C.nox_inventory_cell_t, code uint32) {
	cell.field_132 = 0
	for _, head := range inventoryStateNativeHeads() {
		for equipped := head; equipped != nil; equipped = equipped.NextPtr {
			if equipped.NetCode32 != code {
				continue
			}
			cell.field_132 = 1
			if cell.field_136 != 0 {
				nox_xxx_clientSetAltWeapon_461550(nil)
				cell.field_136 = 0
			}
			return
		}
	}
}

func clientInventoryStorageIndex(point *C.int2) (column, row C.int) {
	return (point.field_0 - 314) / 50, C.int((C.uint32_t(point.field_4) + legacyWords.dword_5d4594_1062512 - 13) / 50)
}

func clientInventoryCheckTabs(point *C.int2) bool {
	return clientInventoryPointInRect(point, 0x587000, 136384) || clientInventoryPointInRect(point, 0x587000, 136400)
}

//export sub_464BD0
func sub_464BD0(win *C.nox_window, event C.int, packed, extra C.uintptr_t) C.int {
	screen := C.int2{field_0: C.int(uint16(packed)), field_4: C.int(packed >> 16)}
	var local C.int2
	point := C.nox_point{x: screen.field_0, y: screen.field_4}
	sub_463370(inventoryWindows.panel, &point, (*C.uint32_t)(unsafe.Pointer(&local)))
	if sub_45D9B0() != 0 || nox_client_inventory_animation_state() != 2 {
		return 1
	}
	switch event {
	case 5:
		if nox_xxx_playerAnimCheck_4372B0() != 0 {
			return 1
		}
		return clientInventoryPress(&local)
	case 7:
		if nox_xxx_playerAnimCheck_4372B0() != 0 || inventoryState.mode == 6 {
			return 1
		}
		if memmap.Uint8(0x5D4594, 1049869) == 0 &&
			clientInventoryPointInRect(&local, 0x587000, 136368) && (local.field_4-13)/50 == 1 {
			if inventoryState.mode == 5 {
				sub_462740()
			} else {
				sub_465CA0()
			}
			return 1
		}
		fallthrough
	case 6:
		if nox_xxx_playerAnimCheck_4372B0() != 0 || inventoryState.mode == 6 {
			return 1
		}
		clientInventoryRelease(&screen, &local)
		return 1
	case 8:
		return 1
	case 9:
		if inventoryState.mode == 5 {
			sub_462740()
			return 1
		}
		return 0
	case 19, 20:
		if nox_xxx_playerAnimCheck_4372B0() != 0 {
			return 1
		}
		if clientInventoryCheckTabs(&local) {
			return C.int(bool2int(inventoryState.mode == 5))
		}
		button := inventoryWindows.scrollUp
		if event == 20 {
			button = inventoryWindows.scrollDown
		}
		nox_window_call_field_94_fnc(inventoryWindows.panel, 16391, uintptr(unsafe.Pointer(button)), 0,
			internCStr("client_inventory_drag_native.go"), 0)
		return 1
	default:
		return C.int(bool2int(inventoryState.mode == 5))
	}
}

func clientInventoryPress(local *C.int2) C.int {
	switch inventoryState.mode {
	case 5:
		if clientInventoryPointInRect(local, 0x587000, 136352) {
			column, row := clientInventoryStorageIndex(local)
			if sub_464B40(column, row) != 0 {
				legacyPointers.inventoryTooltipItem = nox_inventory_prepare_tooltip_drawable_466660(inventoryGridCell(int(row), int(column)))
			}
		} else if sub_478030() != 0 && sub_479870() != 0 && bool(sub_479880((*C.uint32_t)(unsafe.Pointer(local)))) {
			legacyPointers.inventoryTooltipItem = sub_4798A0((*C.uint32_t)(unsafe.Pointer(local)))
		}
	case 6:
		if clientInventoryPointInRect(local, 0x587000, 136352) {
			column, row := clientInventoryStorageIndex(local)
			if sub_464B40(column, row) != 0 {
				cell := inventoryGridCell(int(row), int(column))
				if drawable := nox_inventory_prepare_tooltip_drawable_466660(cell); drawable != nil {
					nox_xxx_trade_4657B0(C.short(drawable.field_32))
				}
			}
		}
	default:
		if memmap.Uint8(0x5D4594, 1049870) == 1 && clientInventoryPointInRect(local, 0x587000, 136336) {
			return 1
		}
		if clientInventoryCheckTabs(local) {
			return 0
		}
		if memmap.Uint8(0x5D4594, 1049869) != 0 {
			return 1
		}
		nox_xxx_wndSetCaptureMain_46ADC0(inventoryWindows.panel)
		if sub_479590() == 3 {
			nox_xxx_clientTradeMB_4657E0((*C.uint32_t)(unsafe.Pointer(local)))
		} else {
			sub_4658A0(inventoryWindows.panel, local)
		}
		if dragged := nox_client_inventory_get_dragged(); dragged != nil {
			nox_xxx_cursorSetDraggedItem_477690(dragged)
			nox_xxx_setKeybTimeout_4160D0(0)
			*memmap.PtrT[C.int2](0x5D4594, 1062572) = *local
			nox_xxx_clientPlaySoundSpecial_452D80(791, 100)
		}
	}
	return 1
}

func clientInventoryRelease(screen, local *C.int2) {
	if inventoryState.mode == 5 {
		if nox_xxx_cursorGetTypePrev_477630() == 7 {
			sub_462740()
			return
		}
	} else {
		nox_xxx_wndClearCaptureMain_46ADE0(inventoryWindows.panel)
	}
	if inventoryState.mode == 4 {
		clientInventoryCheckWorldTarget(screen)
		return
	}
	dragged := nox_client_inventory_get_dragged()
	if dragged == nil {
		return
	}
	defer clientInventoryFinishDrag(dragged)

	fromColumn := C.int(legacyWords.dword_5d4594_1049796_inventory_click_column_index)
	fromRow := C.int(legacyWords.dword_5d4594_1049800_inventory_click_row_index)
	if !bool(nox_xxx_wndPointInWnd_46AAB0(inventoryWindows.panel, screen.field_0, screen.field_4)) || clientInventoryCheckTabs(local) {
		clientInventoryDropOutside(dragged, screen, fromColumn, fromRow)
		return
	}
	initial := memmap.PtrT[C.int2](0x5D4594, 1062572)
	dx, dy := initial.field_0-local.field_0, initial.field_4-local.field_4
	inStorage := clientInventoryPointInRect(local, 0x587000, 136352)
	if !bool(nox_xxx_checkKeybTimeout_4160F0(0, C.uint(gameFPS()/3))) && dx*dx+dy*dy < 100 {
		if inStorage {
			clientInventoryUseFromCell(dragged, fromColumn, fromRow)
			sub_4649B0(dragged, fromColumn, fromRow)
		}
		return
	}
	if clientInventoryPointInRect(local, 0x587000, 136336) && memmap.Uint8(0x5D4594, 1049870) == 0 {
		if legacyWords.dword_5d4594_1049856 == 0 {
			nox_xxx_clientEquip_4623B0(dragged)
			sub_4649B0(dragged, fromColumn, fromRow)
		}
		return
	}
	if !inStorage || clientInventoryIsSpecialThing(dragged.field_27) {
		sub_4649B0(dragged, fromColumn, fromRow)
		return
	}
	column, row := clientInventoryStorageIndex(local)
	legacyWords.dword_5d4594_1049804, legacyWords.dword_5d4594_1049808 = C.uint32_t(column), C.uint32_t(row)
	if sub_464B40(column, row) == 0 {
		return
	}
	destination := inventoryGridCell(int(row), int(column))
	if legacyWords.dword_5d4594_1049856 != 0 {
		*memmap.PtrUint32(0x5D4594, 1049860) = 1
		if item := destination.field_0; destination.field_140 != 0 && item != nil && clientInventoryEquipmentMatches(item, dragged) {
			item.field_32 = destination.field_4
			nox_xxx_clientEquip_4623B0(item)
		} else {
			nox_xxx_clientDequip_464B70(dragged)
		}
		return
	}
	source := inventoryGridCell(int(fromRow), int(fromColumn))
	if source.field_140 != 0 || sub_4649B0(dragged, column, row) == 0 {
		sub_4649B0(dragged, fromColumn, fromRow)
		return
	}
	nox_xxx_clientPlaySoundSpecial_452D80(792, 100)
	if source.field_136 != 0 {
		destination.field_136, source.field_136 = source.field_136, 0
		globalStateNativeStorePointer(globalStateNativeSlot(2), unsafe.Pointer(destination))
	}
	sub_461B50()
}

func clientInventoryEquipmentMatches(a, b *nox_drawable) bool {
	return a.flags28&0x2000000 != 0 && b.flags28&0x2000000 != 0 && a.flags29 == b.flags29 ||
		a.flags28&0x1001000 != 0 && b.flags28&0x1001000 != 0
}

func clientInventoryUseFromCell(dragged *nox_drawable, column, row C.int) {
	if sub_4C12C0() != 0 {
		return
	}
	if dragged.flags28&0x3001000 == 0 {
		nox_xxx_clientUse_465C70(dragged)
		return
	}
	cell := inventoryGridCell(int(row), int(column))
	switch {
	case cell.field_136 != 0:
		nox_xxx_clientSetAltWeapon_461550(nil)
		cell.field_136 = 0
	case cell.field_132 != 0:
		nox_xxx_clientDequip_464B70(dragged)
	default:
		nox_xxx_clientKeyEquip_465C30(column, row)
	}
}

func clientInventoryCheckWorldTarget(screen *C.int2) {
	defer func() { inventoryState.mode = 0 }()
	var world C.int2
	sub_473970(screen, &world)
	point := C.nox_point{x: world.field_0, y: world.field_4}
	target := nox_drawable_find_49ABF0(&point, 20)
	if target == nil {
		nox_xxx_printCentered_445490(clientInventoryString("NoObject", 3869))
		return
	}
	center := C.int2{field_0: clientDisplayStorage.width / 2, field_4: clientDisplayStorage.height / 2}
	sub_473970(&center, &center)
	dx, dy := center.field_0-C.int(target.pos.x), center.field_4-C.int(target.pos.y)
	if dx*dx+dy*dy > 5625 {
		nox_xxx_printCentered_445490(clientInventoryString("ObjectTooFar", 3858))
	}
}

func clientInventoryDropOutside(dragged *nox_drawable, screen *C.int2, column, row C.int) {
	var world C.int2
	sub_473970(screen, &world)
	if legacyWords.dword_5d4594_1049856 == 1 {
		if sub_4C12C0() == 0 {
			nox_xxx_clientDrop_465BE0(&world)
		}
		return
	}
	count := inventoryGridCell(int(row), int(column)).field_140
	if count != 0 {
		nox_xxx_wndClearCaptureMain_46ADE0(inventoryWindows.panel)
		var attrs *C.nox_modifier_attrs_t
		if dragged.flags28&0x13001000 != 0 {
			attrs = (*C.nox_modifier_attrs_t)(unsafe.Pointer(drawableItemData(dragged)))
		}
		sub_4C05F0(0, 0)
		nox_gui_itemAmountDialog_4C0430(clientInventoryString("DropLabel", 4148), screen.field_0, screen.field_4,
			dragged.field_32, dragged.field_27, attrs, C.uint32_t(count)+1, 0, C.nox_gui_item_amount_callback_t(C.sub_465CD0), nil)
	} else if sub_4C12C0() == 0 {
		nox_xxx_clientDrop_465BE0(&world)
	}
	if legacyWords.dword_5d4594_1049856 == 0 {
		sub_4649B0(dragged, column, row)
	}
}
