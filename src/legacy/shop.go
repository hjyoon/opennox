package legacy

/*
#include <stddef.h>
#include <stdint.h>
#include <string.h>

#include "GAME2_2.h"

extern uint32_t dword_5d4594_1098624;

static void nox_test_shop_clear(void) {
	// Test drawables are intentionally synthetic high addresses. Never run the
	// production deletion path over them.
	memset(nox_client_shop_inventory, 0, sizeof(nox_client_shop_inventory));
}

static uintptr_t nox_test_shop_lookup_round_trip(uintptr_t drawable_value) {
	uint32_t old_active = dword_5d4594_1098624;
	nox_test_shop_clear();
	dword_5d4594_1098624 = 1;
	nox_shop_inventory_cell_t* cell = nox_client_shop_inventory_cell(7, 4);
	cell->drawable = (nox_drawable*)drawable_value;
	cell->net_codes[17] = 0x1234ABCDu;
	uintptr_t result = (uintptr_t)sub_478080(0x1234ABCDu);
	dword_5d4594_1098624 = old_active;
	nox_test_shop_clear();
	return result;
}

static uintptr_t nox_test_shop_lookup_order(void) {
	nox_test_shop_clear();
	// Physical index 10 precedes physical index 1 in GAME.EXE's row-outer,
	// column-inner traversal: (row 0, column 1) before (row 1, column 0).
	nox_shop_inventory_cell_t* first = nox_client_shop_inventory_cell(0, 1);
	nox_shop_inventory_cell_t* second = nox_client_shop_inventory_cell(1, 0);
	first->drawable = (nox_drawable*)(uintptr_t)0x1111u;
	second->drawable = (nox_drawable*)(uintptr_t)0x2222u;
	first->net_codes[31] = 77;
	second->net_codes[0] = 77;
	uintptr_t result = (uintptr_t)sub_4780A0(77)->drawable;
	nox_test_shop_clear();
	return result;
}

static uint64_t nox_test_shop_shift_contract(void) {
	nox_test_shop_clear();
	nox_shop_inventory_cell_t* cell = nox_client_shop_inventory_cell(0, 0);
	cell->net_codes[0] = 10;
	cell->net_codes[1] = 20;
	cell->net_codes[2] = 30;
	cell->net_codes[31] = 99;
	int index = sub_4794D0(cell, 20);
	uint64_t result = (uint64_t)(uint8_t)index |
		((uint64_t)(uint8_t)cell->net_codes[0] << 8) |
		((uint64_t)(uint8_t)cell->net_codes[1] << 16) |
		((uint64_t)(uint8_t)cell->net_codes[2] << 24) |
		((uint64_t)(uint8_t)cell->net_codes[30] << 32) |
		((uint64_t)(uint8_t)cell->net_codes[31] << 40);
	nox_test_shop_clear();
	return result;
}

static size_t nox_test_shop_cell_size(void) { return sizeof(nox_shop_inventory_cell_t); }
static size_t nox_test_shop_count_offset(void) { return offsetof(nox_shop_inventory_cell_t, count); }
static size_t nox_test_shop_codes_offset(void) { return offsetof(nox_shop_inventory_cell_t, net_codes); }
static size_t nox_test_shop_price_offset(void) { return offsetof(nox_shop_inventory_cell_t, price); }

static int nox_test_shop_viewport_contract(void) {
	nox_client_shop_viewport_init(1024, 768);
	const nox_draw_viewport_t* vp = nox_client_shop_viewport();
	return vp && vp->x1 == 0 && vp->y1 == 0 && vp->x2 == 1024 && vp->y2 == 768 &&
		vp->field_4 == 0 && vp->field_5 == 0 && vp->field_6 == 0 && vp->field_7 == 0 &&
		vp->width == 1024 && vp->height == 768 && vp->field_10 == 0 && vp->field_11 == 0 &&
		vp->field_12 == 0;
}
*/
import "C"

func shopCellNativeLayout() (size, countOffset, codesOffset, priceOffset uintptr) {
	return uintptr(C.nox_test_shop_cell_size()), uintptr(C.nox_test_shop_count_offset()),
		uintptr(C.nox_test_shop_codes_offset()), uintptr(C.nox_test_shop_price_offset())
}

func shopLookupRoundTrip(value uintptr) uintptr {
	return uintptr(C.nox_test_shop_lookup_round_trip(C.uintptr_t(value)))
}

func shopLookupOrder() uintptr {
	return uintptr(C.nox_test_shop_lookup_order())
}

func shopShiftContract() uint64 {
	return uint64(C.nox_test_shop_shift_contract())
}

func shopViewportNativeContract() bool {
	return C.nox_test_shop_viewport_contract() != 0
}
