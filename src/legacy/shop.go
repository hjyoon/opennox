package legacy

/*
#include <stddef.h>
#include <stdint.h>
#include <string.h>

#include "GAME2_1.h"
#include "GAME2_2.h"
#include "client__gui__guiinv.h"

extern uint32_t dword_5d4594_1098624;
extern nox_inventory_cell_t nox_client_inventory_grid_1050020[NOX_INVENTORY_CELLS_MAX];

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

static uint64_t nox_test_inventory_capacity_contract(uint32_t requested_type,
	uint32_t drawable_type, uint32_t flags, uint8_t count, int amount) {
	nox_inventory_cell_t backup[NOX_INVENTORY_CELLS_MAX];
	memcpy(backup, nox_client_inventory_grid_1050020, sizeof(backup));
	memset(nox_client_inventory_grid_1050020, 0, sizeof(backup));

	nox_drawable drawable = {0};
	drawable.field_27 = drawable_type;
	drawable.flags28 = flags;
	nox_inventory_cell_t* cell = &nox_client_inventory_grid_1050020[0];
	cell->field_0 = &drawable;
	cell->field_140 = count;

	uint64_t result = (uint32_t)sub_467B00((int)requested_type, amount);
	if (sizeof(void*) > 4 && (uintptr_t)cell->field_0 > UINT32_MAX) {
		result |= (uint64_t)1 << 32;
	}
	memcpy(nox_client_inventory_grid_1050020, backup, sizeof(backup));
	return result;
}

static int nox_client_inventory_item_state(uint32_t thing_type, uint32_t* count,
	uint16_t* current_health, uint16_t* maximum_health) {
	if (count) {
		*count = 0;
	}
	if (current_health) {
		*current_health = 0;
	}
	if (maximum_health) {
		*maximum_health = 0;
	}
	int found = 0;
	for (int row = 0; row < NOX_INVENTORY_ROW_COUNT; row++) {
		for (int column = 0; column < NOX_INVENTORY_COL_COUNT; column++) {
			nox_inventory_cell_t* cell = &nox_client_inventory_grid_1050020[
				row + NOX_INVENTORY_ROW_COUNT * column];
			if (!cell->field_140 || !cell->field_0 || cell->field_0->field_27 != thing_type) {
				continue;
			}
			if (count) {
				*count += cell->field_140;
			}
			if (!found) {
				if (current_health) {
					*current_health = cell->field_0->field_73_1;
				}
				if (maximum_health) {
					*maximum_health = cell->field_0->field_73_2;
				}
				found = 1;
			}
		}
	}
	return found;
}

static int nox_client_inventory_item_location(uint32_t thing_type, int* column,
	int* row, uint32_t* net_code) {
	if (column) {
		*column = -1;
	}
	if (row) {
		*row = -1;
	}
	if (net_code) {
		*net_code = 0;
	}
	for (int current_row = 0; current_row < NOX_INVENTORY_ROW_COUNT; current_row++) {
		for (int current_column = 0; current_column < NOX_INVENTORY_COL_COUNT; current_column++) {
			nox_inventory_cell_t* cell = &nox_client_inventory_grid_1050020[
				current_row + NOX_INVENTORY_ROW_COUNT * current_column];
			if (!cell->field_140 || !cell->field_0 || cell->field_0->field_27 != thing_type) {
				continue;
			}
			if (column) {
				*column = current_column;
			}
			if (row) {
				*row = current_row;
			}
			if (net_code) {
				*net_code = cell->field_4;
			}
			return 1;
		}
	}
	return 0;
}

static uint64_t nox_test_inventory_item_state_contract(void) {
	nox_inventory_cell_t backup[NOX_INVENTORY_CELLS_MAX];
	memcpy(backup, nox_client_inventory_grid_1050020, sizeof(backup));
	memset(nox_client_inventory_grid_1050020, 0, sizeof(backup));

	nox_drawable drawable = {0};
	drawable.field_27 = UINT32_C(0x12345678);
	drawable.field_73_1 = UINT16_C(0x2345);
	drawable.field_73_2 = UINT16_C(0x6789);
	nox_inventory_cell_t* cell = &nox_client_inventory_grid_1050020[0];
	cell->field_0 = &drawable;
	cell->field_140 = 3;

	uint32_t count = 0;
	uint16_t current_health = 0;
	uint16_t maximum_health = 0;
	int found = nox_client_inventory_item_state(drawable.field_27, &count, &current_health, &maximum_health);
	uint64_t result = (uint64_t)(uint16_t)count |
		((uint64_t)current_health << 16) |
		((uint64_t)maximum_health << 32);
	if (found) {
		result |= UINT64_C(1) << 63;
	}
	memcpy(nox_client_inventory_grid_1050020, backup, sizeof(backup));
	return result;
}

static int nox_test_inventory_tooltip_pointer_contract(void) {
	nox_drawable drawable = {0};
	nox_inventory_cell_t cell = {0};
	cell.field_0 = &drawable;
	cell.field_4 = UINT32_C(0x89ABCDEF);
	cell.field_140 = 1;
	nox_drawable* result = nox_inventory_prepare_tooltip_drawable_466660(&cell);
	return result == &drawable && drawable.field_32 == UINT32_C(0x89ABCDEF) &&
		(sizeof(void*) <= 4 || (uintptr_t)result > UINT32_MAX);
}

static uint64_t nox_test_inventory_item_location_contract(void) {
	nox_inventory_cell_t backup[NOX_INVENTORY_CELLS_MAX];
	memcpy(backup, nox_client_inventory_grid_1050020, sizeof(backup));
	memset(nox_client_inventory_grid_1050020, 0, sizeof(backup));

	nox_drawable drawable = {0};
	drawable.field_27 = UINT32_C(0x12345678);
	nox_inventory_cell_t* cell = &nox_client_inventory_grid_1050020[
		19 + NOX_INVENTORY_ROW_COUNT * 2];
	cell->field_0 = &drawable;
	cell->field_4 = UINT32_C(0x89ABCDEF);
	cell->field_140 = 1;

	int column = -1;
	int row = -1;
	uint32_t net_code = 0;
	int found = nox_client_inventory_item_location(drawable.field_27, &column, &row, &net_code);
	uint64_t result = (uint64_t)(uint8_t)row |
		((uint64_t)(uint8_t)column << 8) |
		((uint64_t)net_code << 16);
	if (found) {
		result |= UINT64_C(1) << 63;
	}
	memcpy(nox_client_inventory_grid_1050020, backup, sizeof(backup));
	return result;
}

static uint64_t nox_test_inventory_trade_net_code_contract(void) {
	nox_inventory_cell_t cell = {0};
	cell.field_0 = (nox_drawable*)(uintptr_t)(sizeof(void*) > 4 ?
		UINT64_C(0xAABBCCDD55667788) : UINT32_C(0x55667788));
	cell.field_4 = UINT32_C(0x11112222);
	cell.data_4[0] = UINT32_C(0x33334444);
	cell.field_140 = 1;
	uint32_t first = nox_inventory_cell_last_net_code_4657E0(&cell);
	cell.field_140 = 2;
	uint32_t second = nox_inventory_cell_last_net_code_4657E0(&cell);
	return (uint64_t)first | ((uint64_t)second << 32);
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

func inventoryCapacityContract(requestedType, drawableType, flags uint32, count uint8, amount int) (available int, highPointer bool) {
	result := uint64(C.nox_test_inventory_capacity_contract(
		C.uint32_t(requestedType), C.uint32_t(drawableType), C.uint32_t(flags), C.uint8_t(count), C.int(amount),
	))
	return int(uint32(result)), result>>32 != 0
}

func Nox_client_inventoryItemState(thingType uint32) (found bool, count uint32, currentHealth, maximumHealth uint16) {
	var ccount C.uint32_t
	var ccurrent, cmaximum C.uint16_t
	found = C.nox_client_inventory_item_state(
		C.uint32_t(thingType), &ccount, &ccurrent, &cmaximum,
	) != 0
	return found, uint32(ccount), uint16(ccurrent), uint16(cmaximum)
}

func Nox_client_inventoryItemLocation(thingType uint32) (found bool, column, row int, netCode uint32) {
	var ccolumn, crow C.int
	var cnetCode C.uint32_t
	found = C.nox_client_inventory_item_location(
		C.uint32_t(thingType), &ccolumn, &crow, &cnetCode,
	) != 0
	return found, int(ccolumn), int(crow), uint32(cnetCode)
}

func inventoryItemStateContract() (found bool, count uint16, currentHealth, maximumHealth uint16) {
	result := uint64(C.nox_test_inventory_item_state_contract())
	return result>>63 != 0, uint16(result), uint16(result >> 16), uint16(result >> 32)
}

func inventoryTooltipPointerContract() bool {
	return C.nox_test_inventory_tooltip_pointer_contract() != 0
}

func inventoryItemLocationContract() (found bool, column, row int, netCode uint32) {
	result := uint64(C.nox_test_inventory_item_location_contract())
	return result>>63 != 0, int(uint8(result >> 8)), int(uint8(result)), uint32(result >> 16)
}

func inventoryTradeNetCodeContract() (first, second uint32) {
	result := uint64(C.nox_test_inventory_trade_net_code_contract())
	return uint32(result), uint32(result >> 32)
}

func Nox_client_gold_4674A0() uint32 {
	return uint32(C.sub_4674A0())
}
