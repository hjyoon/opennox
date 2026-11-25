package legacy

/*
#include "GAME2_1.h"
#include "GAME2_3.h"
#include "GAME3_3.h"
#include "client__gui__guiinv.h"
#include "client__gui__window.h"

extern uint32_t dword_5d4594_1049864, dword_5d4594_1062512, dword_5d4594_1062516;
extern uint32_t dword_5d4594_1062552;
extern nox_window* nox_win_unk5;
extern nox_window* nox_inventory_scroll_window;
*/
import "C"

import (
	"unsafe"

	"github.com/opennox/opennox/v1/common/flags"
	"github.com/opennox/opennox/v1/common/memmap"
)

func clientInventoryPlayableCell(column, row C.int) *C.nox_inventory_cell_t {
	if column < 0 || row < 0 || int(column) >= inventoryColumns || int(row) >= inventoryRows-1 {
		return nil
	}
	return inventoryGridCell(int(row), int(column))
}

//export sub_4673F0
func sub_4673F0(x, y C.int) C.int {
	*memmap.PtrInt32(0x5D4594, 1062580) = int32(x)
	*memmap.PtrInt32(0x5D4594, 1062584) = int32(y)
	return x
}

//export sub_467410
func sub_467410(value C.int) C.int {
	*memmap.PtrInt32(0x5D4594, 1062540) = int32(value)
	return value
}

//export sub_467420
func sub_467420(value C.char) C.char {
	*memmap.PtrUint8(0x5D4594, 1062536) = byte(value)
	return value
}

//export sub_467430
func sub_467430() C.uchar {
	return C.uchar(memmap.Uint8(0x5D4594, 1062536))
}

//export sub_467440
func sub_467440(value C.int) C.int {
	*memmap.PtrInt32(0x5D4594, 1062544) = int32(value)
	return value
}

//export sub_467450
func sub_467450(value C.int) C.int {
	*memmap.PtrInt32(0x5D4594, 1062548) = int32(value)
	return value
}

//export sub_467470
func sub_467470(index C.int, value C.float) C.int {
	index = C.int(byte(index))
	*memmap.PtrFloat32(0x5D4594, 1063100+4*uintptr(index)) = float32(value)
	return index
}

//export sub_467490
func sub_467490(value C.int) C.int {
	C.dword_5d4594_1062552 = C.uint32_t(value)
	return value
}

//export nox_window_set_visible_unk5
func nox_window_set_visible_unk5(visible C.int) {
	nox_window_set_hidden(C.nox_win_unk5, int32(bool2int(visible == 0)))
}

//export nox_xxx_cliUseCurePoison_4674E0
func nox_xxx_cliUseCurePoison_4674E0(thingType C.int) {
	if nox_xxx_guiCursor_477600() == 1 || nox_xxx_checkGameFlagPause_413A50() != 0 {
		return
	}
	if cell := nox_xxx_cliInventoryFirstItemByTT_467520(thingType); cell != nil && cell.field_0 != nil {
		cell.field_0.field_32 = cell.field_4
		nox_xxx_clientUse_465C70(cell.field_0)
	}
}

//export nox_xxx_cliInventoryFirstItemByTT_467520
func nox_xxx_cliInventoryFirstItemByTT_467520(thingType C.int) *C.nox_inventory_cell_t {
	for row := range inventoryRows {
		for column := range inventoryColumns {
			cell := inventoryGridCell(row, column)
			if cell.field_140 != 0 && cell.field_0 != nil && cell.field_0.field_27 == C.uint32_t(thingType) {
				return cell
			}
		}
	}
	return nil
}

func draggedInventoryItem(code C.int) *nox_drawable {
	if dragged := nox_client_inventory_get_dragged(); dragged != nil && dragged.field_32 == C.uint32_t(code) {
		return dragged
	}
	return nil
}

//export sub_4675E0
func sub_4675E0(code C.int, current, maximum C.short) C.short {
	var item *nox_drawable
	if cell, _ := inventoryLookupNativeSearch(uint32(code)); cell != nil {
		item = cell.field_0
	}
	if item == nil {
		item = draggedInventoryItem(code)
	}
	if item == nil {
		return 0
	}
	item.field_73_1, item.field_73_2 = C.uint16_t(current), C.uint16_t(maximum)
	return maximum
}

//export sub_467650
func sub_467650() C.int {
	sub_462740()
	C.dword_5d4594_1049864 = 6
	nox_client_setCursorType_477610(8)
	if result := sub_467C80(); result != 0 {
		return result
	}
	return sub_467BB0()
}

//export sub_4676D0
func sub_4676D0(code C.int) *C.nox_drawable {
	if cell, _ := inventoryLookupNativeSearch(uint32(code)); cell != nil {
		return cell.field_0
	}
	return draggedInventoryItem(code)
}

//export sub_467700
func sub_467700(code C.int) C.int {
	if cell, _ := inventoryLookupNativeSearch(uint32(code)); cell != nil {
		return C.int(cell.field_140)
	}
	return C.int(bool2int(draggedInventoryItem(code) != nil))
}

//export sub_467810
func sub_467810(column, row C.int) C.int {
	if cell := clientInventoryPlayableCell(column, row); cell != nil {
		return C.int(cell.field_140)
	}
	return 0
}

//export sub_467850
func sub_467850(thingType C.int) C.int {
	if cell := nox_xxx_cliInventoryFirstItemByTT_467520(thingType); cell != nil {
		return C.int(cell.field_140)
	}
	return 0
}

//export sub_467870
func sub_467870(column, row C.int) *C.char {
	if cell := clientInventoryPlayableCell(column, row); cell != nil {
		return (*C.char)(unsafe.Pointer(&cell.field_4))
	}
	return nil
}

//export sub_4678D0
func sub_4678D0() *C.nox_drawable {
	player := Get_dword_8531A0_2576()
	if player == nil {
		return nil
	}
	if item := inventoryEquippedWeapon(player.WeaponEquip); item != nil {
		if cell, _ := inventoryLookupNativeSearch(uint32(item.field_32)); cell != nil {
			return cell.field_0
		}
	}
	return nil
}

func inventoryEquippedWeapon(weapons uint32) *nox_drawable {
	for bit := 2; bit < 27; bit++ {
		mask := uint32(1) << bit
		if weapons&mask == 0 {
			continue
		}
		if item := sub_461600(C.int(sub_415840(int32(mask)))); item != nil {
			return item
		}
	}
	return nil
}

//export sub_467930
func sub_467930(code, current, maximum C.int) *C.nox_drawable {
	if code == 0 {
		return nil
	}
	cell, _ := inventoryLookupNativeSearch(uint32(code))
	if cell == nil || cell.field_0 == nil {
		return nil
	}
	item := cell.field_0
	data := drawableItemData(item)
	data.Field_112_0, data.Field_112_2 = int16(current), int16(maximum)
	if cell.field_132 == 1 {
		sub_470D90(current, maximum)
	}
	return item
}

//export sub_467B00
func sub_467B00(thingType, incoming C.int) C.int {
	var available C.int
	for row := range inventoryRows - 1 {
		for column := range inventoryColumns {
			cell := inventoryGridCell(row, column)
			count := C.int(cell.field_140)
			if count == 0 {
				available++
				continue
			}
			item := cell.field_0
			if item == nil || item.field_27 != C.uint32_t(thingType) || item.flags28&0x4000000 != 0 {
				continue
			}
			limit := C.int(31)
			if item.flags28&0x10 != 0 {
				limit = 3
				if noxflags.HasGame(noxflags.GameModeCoop | noxflags.GameModeQuest) {
					limit = 9
				}
			}
			if incoming+count <= limit {
				available++
			}
		}
	}
	return available
}

//export sub_467BB0
func sub_467BB0() C.int {
	if result := nox_gui_xxx_check_446360(); result != 0 {
		return C.int(result)
	}
	if result := sub_4AE3D0(); result != 0 {
		return result
	}
	if result := nox_xxx_guiCursor_477600(); result != 0 {
		return result
	}
	if result := nox_xxx_playerAnimCheck_4372B0(); result != 0 {
		return result
	}
	if result := nox_xxx_get_57AF20(); result != 0 {
		return result
	}
	state := memmap.PtrUint8(0x5D4594, 1049868)
	if *state == 3 || *state == 0 {
		*state = 1
		nox_xxx_clientPlaySoundSpecial_452D80(789, 100)
	}
	position := C.dword_5d4594_1062516
	C.dword_5d4594_1062512 = position
	return C.int(position)
}

//export sub_467C10
func sub_467C10() C.int {
	if C.dword_5d4594_1049864 == 6 {
		return 1
	}
	if sub_467C80() == 0 {
		return 0
	}
	*memmap.PtrUint8(0x5D4594, 1049868) = 3
	nox_xxx_clientPlaySoundSpecial_452D80(790, 100)
	if C.dword_5d4594_1049864 == 5 {
		sub_462740()
	}
	sub_467CD0()
	return 1
}

//export nox_client_toggleInventory_467C60
func nox_client_toggleInventory_467C60() C.int {
	if sub_467C80() != 0 {
		return sub_467C10()
	}
	return sub_467BB0()
}

//export sub_467C80
func sub_467C80() C.int {
	state := nox_client_inventory_animation_state()
	return C.int(bool2int(state == 1 || state == 2))
}

//export sub_467CA0
func sub_467CA0() C.int {
	if result := sub_467C80(); result != 0 {
		return result
	}
	C.dword_5d4594_1062516 = 0
	win := C.nox_inventory_scroll_window
	if win == nil {
		return 0
	}
	slider := game3Slider(win)
	return C.int(nox_window_call_field_94_fnc(win, 16394, uintptr(slider.Max), 0, internCStr("client_inventory_entries_native.go"), 0))
}
