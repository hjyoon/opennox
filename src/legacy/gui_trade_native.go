package legacy

/*
#include <stddef.h>
#include <stdint.h>

#include "GAME3_1.h"

static int nox_test_gui_trade_pointer_round_trip(uintptr_t value) {
	nox_gui_trade_slot* slot = nox_gui_trade_slot_at(0, 0);
	nox_gui_trade_slot old_slot = *slot;
	nox_gui_trade_slot* old_local = dword_5d4594_1320932;
	nox_gui_trade_slot* old_remote = dword_5d4594_1320936;
	nox_drawable* old_dragged = dword_5d4594_1320968;
	nox_gui_trade_slot* old_source = dword_5d4594_1320972;

	slot->drawable = (nox_drawable*)value;
	dword_5d4594_1320932 = (nox_gui_trade_slot*)value;
	dword_5d4594_1320936 = (nox_gui_trade_slot*)value;
	dword_5d4594_1320968 = (nox_drawable*)value;
	dword_5d4594_1320972 = (nox_gui_trade_slot*)value;
	int result = (uintptr_t)slot->drawable == value &&
		(uintptr_t)dword_5d4594_1320932 == value &&
		(uintptr_t)dword_5d4594_1320936 == value &&
		(uintptr_t)dword_5d4594_1320968 == value &&
		(uintptr_t)dword_5d4594_1320972 == value;

	*slot = old_slot;
	dword_5d4594_1320932 = old_local;
	dword_5d4594_1320936 = old_remote;
	dword_5d4594_1320968 = old_dragged;
	dword_5d4594_1320972 = old_source;
	return result;
}

static int nox_test_gui_trade_slot_contract(void) {
	nox_drawable drawable = {0};
	drawable.field_27 = UINT32_C(0x12345678);
	nox_gui_trade_slot slot = {
		.drawable = &drawable,
		.count = 3,
		.item_ids = {UINT32_C(0x1111), UINT32_C(0x2222), UINT32_C(0x3333)},
		.total_cost = UINT32_C(900),
	};
	if (!sub_4C1760(&slot, UINT32_C(0x2222)) || sub_4C1760(&slot, UINT32_C(0x4444))) {
		return 0;
	}
	if (!sub_4C18E0(drawable.field_27, &slot) || sub_4C18E0(UINT32_C(0x87654321), &slot)) {
		return 0;
	}
	if (sub_4C1710(&slot, UINT32_C(0x2222)) != 1 ||
		slot.item_ids[0] != UINT32_C(0x1111) ||
		slot.item_ids[1] != UINT32_C(0x3333) ||
		slot.item_ids[2] != 0) {
		return 0;
	}
	drawable.flags28 = UINT32_C(0x1000);
	return !sub_4C18E0(drawable.field_27, &slot);
}

static int nox_test_gui_trade_slot_selection_order(void) {
	nox_gui_trade_slot saved[NOX_GUI_TRADE_SLOT_COUNT];
	nox_drawable drawables[NOX_GUI_TRADE_SLOT_COUNT] = {0};
	for (int index = 0; index < NOX_GUI_TRADE_SLOT_COUNT; index++) {
		nox_gui_trade_slot* slot = nox_gui_trade_slot_at(0, index);
		saved[index] = *slot;
		drawables[index].field_27 = UINT32_C(0x1000) + (uint32_t)index;
		*slot = (nox_gui_trade_slot){
			.drawable = &drawables[index],
			.count = 1,
		};
	}

	// The original nested row/column walk visits slots 0, 2, 1, 3.
	nox_gui_trade_slot_at(0, 2)->drawable = NULL;
	int empty_order_ok = sub_4C1910(UINT32_C(0x9999)) == nox_gui_trade_slot_at(0, 2);
	nox_gui_trade_slot_at(0, 2)->drawable = &drawables[2];
	drawables[0].field_27 = UINT32_C(0x7777);
	drawables[2].field_27 = UINT32_C(0x7777);
	int stack_order_ok = sub_4C1910(UINT32_C(0x7777)) == nox_gui_trade_slot_at(0, 0);

	for (int index = 0; index < NOX_GUI_TRADE_SLOT_COUNT; index++) {
		*nox_gui_trade_slot_at(0, index) = saved[index];
	}
	return empty_order_ok && stack_order_ok;
}

static size_t nox_test_gui_trade_slot_size(void) {
	return sizeof(nox_gui_trade_slot);
}

static size_t nox_test_gui_trade_slot_total_offset(void) {
	return offsetof(nox_gui_trade_slot, total_cost);
}
*/
import "C"

func guiTradePointerRoundTrip(value uintptr) bool {
	return C.nox_test_gui_trade_pointer_round_trip(C.uintptr_t(value)) != 0
}

func guiTradeSlotContract() bool {
	return C.nox_test_gui_trade_slot_contract() != 0
}

func guiTradeSlotSelectionOrder() bool {
	return C.nox_test_gui_trade_slot_selection_order() != 0
}

func guiTradeSlotNativeLayout() (size, totalOffset uintptr) {
	return uintptr(C.nox_test_gui_trade_slot_size()), uintptr(C.nox_test_gui_trade_slot_total_offset())
}
