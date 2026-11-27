package legacy

/*
#include "client__gui__guiinv.h"
#include "client__gui__guimsg.h"
#include "common__object__modifier.h"
*/
import "C"

import (
	"strconv"
	"unsafe"

	noxflags "github.com/opennox/opennox/v1/common/flags"
	"github.com/opennox/opennox/v1/common/memmap"
)

//export sub_462740
func sub_462740() C.int {
	if wndIsShown_nox_xxx_wndIsShown_46ACC0(inventoryWindows.identify) != 0 {
		return 0
	}
	nox_window_set_hidden(inventoryWindows.identify, 1)
	legacyPointers.inventoryTooltipItem, legacyPointers.inventoryPreviousTooltipItem = nil, nil
	text := (*wchar2_t)(memmap.PtrOff(0x5D4594, 1063124))
	copyWideString(text, clientInventoryString("thing.db:IdentifyDescription", 2361))
	child := nox_xxx_wndGetChildByID_46B0C0(inventoryWindows.identify, 9156)
	nox_window_call_field_94_fnc(child, 16399, 0, 0, internCStr("client_inventory_tooltip_native.go"), 0)
	nox_xxx_wndClearCaptureMain_46ADE0(inventoryWindows.panel)
	inventoryState.mode = 0
	nox_client_setCursorType_477610(0)
	return 1
}

//export sub_466160
func sub_466160() C.int {
	if nox_client_inventory_animation_state() == 2 {
		nox_xxx_cursorSetTooltip_4776B0(clientInventoryString("CloseInventoryTT", 410))
	} else {
		nox_xxx_cursorSetTooltip_4776B0(clientInventoryString("OpenInventoryTT", 414))
	}
	return 1
}

//export sub_4661D0
func sub_4661D0() C.int {
	if cell := clientInventoryAlternateCell(); cell != nil {
		nox_xxx_cursorSetTooltip_4776B0(nox_xxx_clientAskInfoMb_4BF050(cell.field_0))
	} else {
		nox_xxx_cursorSetTooltip_4776B0(clientInventoryString("ToolTipWeapon2Area", 3331))
	}
	return 1
}

//export nox_inventory_prepare_tooltip_drawable_466660
func nox_inventory_prepare_tooltip_drawable_466660(cell *C.nox_inventory_cell_t) *C.nox_drawable {
	if cell == nil || cell.field_140 == 0 || cell.field_0 == nil {
		return nil
	}
	drawable := cell.field_0
	drawable.field_32 = cell.field_4
	return drawable
}

func clientInventoryPointInRect(point *C.int2, segment, offset uintptr) bool {
	bounds := (*C.int4)(memmap.PtrOff(segment, offset))
	return nox_xxx_pointInRect_4281F0(point, bounds) != 0
}

//export sub_466660
func sub_466660(_ C.int, point *C.int2) *C.wchar2_t {
	if clientInventoryPointInRect(point, 0x587000, 136336) {
		if memmap.Uint8(0x5D4594, 1049870) == 1 {
			return nil
		}
		slot := sub_465990((*C.uint32_t)(unsafe.Pointer(point)))
		if slot == -1 {
			return clientInventoryString("DollRegionError", 3155)
		}
		if drawable := inventoryStateNativeHeads()[slot]; drawable != nil {
			return nox_xxx_clientAskInfoMb_4BF050((*nox_drawable)(drawable.C()))
		}
		return clientInventoryString("ToolTipDrag", 3159)
	}
	if memmap.Uint8(0x5D4594, 1049869) != 0 {
		return nil
	}
	var row, column C.int
	switch {
	case clientInventoryPointInRect(point, 0x587000, 136368):
		column = (point.field_4 - 13) / 50
		row = C.int(inventoryRows - 1)
	case clientInventoryPointInRect(point, 0x587000, 136352):
		column, row = clientInventoryStorageIndex(point)
	default:
		return nil
	}
	legacyWords.dword_5d4594_1049796_inventory_click_column_index = C.uint32_t(column)
	legacyWords.dword_5d4594_1049800_inventory_click_row_index = C.uint32_t(row)
	if sub_464B40(column, row) != 0 {
		drawable := nox_inventory_prepare_tooltip_drawable_466660(inventoryGridCell(int(row), int(column)))
		if drawable != nil {
			return nox_xxx_clientAskInfoMb_4BF050(drawable)
		}
	}
	return nil
}

//export nox_xxx_inventroryOnHovewerSub_4667E0
func nox_xxx_inventroryOnHovewerSub_4667E0(control, event C.int, packed C.uint) C.int {
	x := int(uint16(packed))
	icon := 0
	for boundary := 40; boundary <= x; boundary += 35 {
		icon++
	}
	for spell := range 31 {
		if memmap.Uint32(0x5D4594, 1062540)&(uint32(1)<<spell) == 0 {
			continue
		}
		if icon == 0 {
			nox_xxx_cursorSetTooltip_4776B0(nox_xxx_spellTitle_424930(nox_xxx_getEnchantSpell_424920(int32(spell))))
			return 1
		}
		icon--
	}
	for effect := range 6 {
		mask := byte(1) << effect
		if memmap.Uint8(0x5D4594, 1062536)&mask == 0 {
			continue
		}
		if icon == 0 {
			nox_xxx_cursorSetTooltip_4776B0(sub_413480(C.char(mask)))
			return 1
		}
		icon--
	}
	var text *wchar2_t
	if noxflags.HasGame(noxflags.GameModeQuest) {
		point := C.int2{field_0: C.int(x), field_4: C.int(packed >> 16)}
		switch {
		case clientInventoryPointInRect(&point, 0x5D4594, 1049812):
			text = clientInventoryString("thing.db:AnkhGUI", 4385)
		case clientInventoryPointInRect(&point, 0x5D4594, 1049828) && sub_4BFD30() == 1:
			text = clientInventoryString("GeneralPrint:TooltipKeyIcon", 4388)
		}
	}
	nox_xxx_cursorSetTooltip_4776B0(text)
	return 1
}

//export sub_466E20
func sub_466E20(id *C.uint32_t) C.int {
	var text *wchar2_t
	switch *id {
	case 9105:
		text = clientInventoryString("JournalModeTT", 424)
	case 9106:
		text = clientInventoryString("InventoryModeTT", 428)
	case 9107:
		text = clientInventoryString("StatsModeTT", 432)
	case 9108:
		text = clientInventoryString("PaperDollModeTT", 436)
	case 9111:
		text = clientInventoryString("CloseInventoryTT", 440)
	default:
		return 0
	}
	nox_xxx_cursorSetTooltip_4776B0(text)
	return 1
}

//export nox_xxx_inventoryNameSignInit_4671E0
func nox_xxx_inventoryNameSignInit_4671E0() C.int {
	title := (*wchar2_t)(memmap.PtrOff(0x5D4594, 1062588))
	copyWideString(title, (*wchar2_t)(memmap.PtrOff(0x5D4594, 1063676)))
	player := Get_dword_8531A0_2576()
	var level int32
	if noxflags.HasGame(noxflags.GameModeQuest) || nox_xxx_isQuest_4D6F50() != 0 || sub_4D6F70() != 0 {
		level = min(int32(legacyWords.dword_5d4594_1049844), C.NOX_PLAYER_MAX_LEVEL)
	} else {
		if player == nil {
			return 0
		}
		level = int32(player.Level)
	}
	if player != nil {
		names := [...]string{"Warrior", "Wizard", "Conjurer"}
		class := int(player.PlayerClass())
		if class < len(names) {
			key := "experience:" + names[class] + strconv.Itoa(int(level))
			levelName := clientInventoryString(key, 4763)
			format := clientInventoryString("ElaborateNameFormat", 4761)
			name := (*wchar2_t)(unsafe.Pointer(&player.NameFinal[0]))
			return formatFixedWide(title, format, formatValue{text: unsafe.Pointer(name)}, formatValue{text: unsafe.Pointer(levelName)})
		}
	}
	return C.int(level)
}

//export sub_467750
func sub_467750(code C.int, rejected C.char) C.int {
	previous := clientInventoryAlternateCell()
	if code != 0 {
		if cell, _ := inventoryLookupNativeSearch(uint32(code)); cell != nil {
			if previous != nil {
				previous.field_136 = 0
			}
			storeForeignPointer(globalStateNativeSlot(2), unsafe.Pointer(cell))
			cell.field_136 = 1
			return 1
		}
	} else if previous != nil {
		previous.field_136 = 0
		storeForeignPointer(globalStateNativeSlot(2), nil)
	}
	if rejected != 0 {
		if rejected != 1 {
			return 0
		}
		nox_xxx_printCentered_445490(clientInventoryString("Weapon2CantUse", 5379))
		if code := clientItemStorage.secondary_previous; code != 0 {
			if cell, _ := inventoryLookupNativeSearch(uint32(code)); cell != nil {
				nox_xxx_clientSetAltWeapon_461550(cell)
				return 0
			}
		}
	}
	clientItemStorage.secondary_previous = 0
	return 0
}
