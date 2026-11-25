package legacy

/*
#include "GAME3_1.h"
#include "noxstring.h"
#include "common__strman.h"
#include "client__gui__guimsg.h"
extern nox_gui_trade_slot nox_gui_trade_local_slots[NOX_GUI_TRADE_SLOT_COUNT];
extern nox_gui_trade_slot nox_gui_trade_remote_slots[NOX_GUI_TRADE_SLOT_COUNT];
extern uint32_t dword_5d4594_1320944, dword_5d4594_1320948, dword_5d4594_1320964;
*/
import "C"

import (
	"encoding/binary"
	"slices"
	"strconv"
	"unsafe"
)

type tradeClientSlot = C.nox_gui_trade_slot

const (
	tradeSlotCapacity = C.NOX_GUI_TRADE_SLOT_ITEM_CAPACITY
	tradeUnstackable  = 0x13001000
)

//export sub_4C1710
func sub_4C1710(slot *tradeClientSlot, id C.uint32_t) C.int {
	if slot == nil {
		return tradeSlotCapacity
	}
	items := slot.item_ids[:min(slot.count, tradeSlotCapacity)]
	index := slices.Index(items, id)
	if index < 0 {
		return tradeSlotCapacity
	}
	copy(items[index:], items[index+1:])
	items[len(items)-1] = 0
	return C.int(index)
}

//export sub_4C1760
func sub_4C1760(slot *C.nox_trade_const_slot_t, id C.uint32_t) C.int {
	if slot == nil {
		return 0
	}
	items := slot.item_ids[:min(slot.count, tradeSlotCapacity)]
	return C.int(bool2int(slices.Contains(items, id)))
}

//export nox_xxx_tradeClientAddItem_4C1790
func nox_xxx_tradeClientAddItem_4C1790(data *C.nox_trade_const_u8_t) *tradeClientSlot {
	if C.dword_5d4594_1320964 == 0 || data == nil {
		return nil
	}
	C.dword_5d4594_1320944, C.dword_5d4594_1320948 = 0, 0
	packet := unsafe.Slice((*byte)(unsafe.Pointer(data)), 11)
	thing := C.uint32_t(binary.LittleEndian.Uint16(packet[3:5]))
	slot, remote := C.dword_5d4594_1320936, C.int(1)
	if packet[2] == 1 {
		slot, remote = C.dword_5d4594_1320932, 0
	}
	if sub_4C18E0(thing, (*C.nox_trade_const_slot_t)(slot)) == 0 {
		slot = findTradeSlot(remote, thing)
	}
	if slot == nil || slot.count >= tradeSlotCapacity {
		return nil
	}
	if slot.drawable == nil {
		slot.drawable = nox_new_drawable_for_thing(int32(thing))
		if slot.drawable == nil {
			return nil
		}
		if slot.drawable.flags28&tradeUnstackable != 0 {
			var itemModifiers [4]unsafe.Pointer
			modifierIDs := unsafe.Slice((*byte)(unsafe.Add(unsafe.Pointer(data), 11)), 4)
			for i, id := range modifierIDs {
				if id != 0xff {
					itemModifiers[i] = nox_xxx_modifGetDescById_413330(int32(id))
				}
			}
			fields := drawableItemData(slot.drawable)
			fields.Field_108, fields.Field_109 = itemModifiers[0], itemModifiers[1]
			fields.Field_110, fields.Field_111 = itemModifiers[2], itemModifiers[3]
		}
		slot.count, slot.total_cost = 0, 0
	}
	slot.item_ids[slot.count] = C.uint32_t(binary.LittleEndian.Uint16(packet[5:7]))
	slot.count++
	slot.total_cost += C.uint32_t(binary.LittleEndian.Uint32(packet[7:11]))
	C.dword_5d4594_1320932, C.dword_5d4594_1320936 = nil, nil
	return slot
}

//export sub_4C18E0
func sub_4C18E0(thing C.uint32_t, slot *C.nox_trade_const_slot_t) C.int {
	if slot == nil || slot.count >= tradeSlotCapacity {
		return 0
	}
	if slot.count == 0 {
		return 1
	}
	dr := slot.drawable
	return C.int(bool2int(dr != nil && dr.field_27 == thing && dr.flags28&tradeUnstackable == 0))
}

func findTradeSlot(remote C.int, thing C.uint32_t) *tradeClientSlot {
	slots := [...]C.int{0, 2, 1, 3}
	for _, index := range slots {
		slot := nox_gui_trade_slot_at(remote, index)
		dr := slot.drawable
		if dr != nil && slot.count < tradeSlotCapacity && dr.field_27 == thing && dr.flags28&tradeUnstackable == 0 {
			return slot
		}
	}
	for _, index := range slots {
		if slot := nox_gui_trade_slot_at(remote, index); slot.drawable == nil {
			return slot
		}
	}
	return nil
}

//export sub_4C1910
func sub_4C1910(thing C.uint32_t) *tradeClientSlot {
	return findTradeSlot(0, thing)
}

//export sub_4C19C0
func sub_4C19C0(thing C.uint32_t) *tradeClientSlot {
	return findTradeSlot(1, thing)
}

//export sub_4C1B50
func sub_4C1B50(data *C.nox_trade_const_u8_t) C.int {
	if C.dword_5d4594_1320964 == 0 || data == nil {
		return C.int(C.dword_5d4594_1320964)
	}
	packet := unsafe.Slice((*byte)(unsafe.Pointer(data)), 14)
	playerTotal := strconv.FormatInt(int64(int32(binary.LittleEndian.Uint32(packet[2:6]))), 10)
	itemCount := int32(binary.LittleEndian.Uint32(packet[6:10]))
	remoteTotal := strconv.FormatInt(int64(int32(binary.LittleEndian.Uint32(packet[10:14]))), 10)
	WStrCopy((*wchar2_t)(g2ServerMem(1320240)), 16, playerTotal)
	countText := (*wchar2_t)(g2ServerMem(1320868))
	if itemCount != 0 {
		WStrCopy(countText, 16, "("+strconv.FormatInt(int64(itemCount), 10)+")")
	} else {
		copyWideString(countText, (*wchar2_t)(g2ServerMem(1320984)))
	}
	return C.int(WStrCopy((*wchar2_t)(g2ServerMem(1320100)), 16, remoteTotal))
}

//export sub_4C1BC0
func sub_4C1BC0(data *C.nox_trade_const_u8_t) C.int {
	if C.dword_5d4594_1320964 != 0 && data != nil {
		flags := *(*byte)(unsafe.Add(unsafe.Pointer(data), 2))
		C.dword_5d4594_1320944 = C.uint32_t(flags & 1)
		C.dword_5d4594_1320948 = C.uint32_t((flags >> 1) & 1)
	}
	return C.int(C.dword_5d4594_1320964)
}

//export nox_gui_trade_slot_at
func nox_gui_trade_slot_at(remote, index C.int) *tradeClientSlot {
	if index < 0 || index >= C.NOX_GUI_TRADE_SLOT_COUNT {
		return nil
	}
	if remote != 0 {
		return &C.nox_gui_trade_remote_slots[index]
	}
	return &C.nox_gui_trade_local_slots[index]
}

//export nox_gui_trade_slots_reset
func nox_gui_trade_slots_reset() {
	for remote := C.int(0); remote < 2; remote++ {
		for index := C.int(0); index < C.NOX_GUI_TRADE_SLOT_COUNT; index++ {
			slot := nox_gui_trade_slot_at(remote, index)
			if slot.drawable != nil {
				nox_xxx_spriteDelete_45A4B0(slot.drawable)
			}
			*slot = tradeClientSlot{}
		}
	}
}

func tradeString(name string, line C.int) *wchar2_t {
	return C.nox_strman_loadString_40F1D0(internCStr(name), nil, internCStr("C:\\NoxPost\\src\\client\\Gui\\GUITrade.c"), line)
}

//export sub_4C15D0
func sub_4C15D0(data *C.nox_trade_const_u8_t) C.int {
	if C.dword_5d4594_1320964 == 0 || data == nil {
		return 0
	}
	packet := unsafe.Slice((*byte)(unsafe.Pointer(data)), 4)
	id := C.uint32_t(binary.LittleEndian.Uint16(packet[2:4]))
	for remote := C.int(0); remote < 2; remote++ {
		for index := C.int(0); index < C.NOX_GUI_TRADE_SLOT_COUNT; index++ {
			slot := nox_gui_trade_slot_at(remote, index)
			if sub_4C1760((*C.nox_trade_const_slot_t)(slot), id) == 0 {
				continue
			}
			slot.total_cost -= slot.total_cost / slot.count
			sub_4C1710(slot, id)
			slot.count--
			if slot.count == 0 {
				nox_xxx_spriteDelete_45A4B0(slot.drawable)
				slot.drawable, slot.total_cost = nil, 0
			}
			return C.int(slot.count)
		}
	}
	C.nox_xxx_printCentered_445490(tradeString("TradeGUIItemNotFound", 1141))
	return 0
}
