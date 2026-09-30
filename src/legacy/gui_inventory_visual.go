package legacy

/*
#include "GAME2_1.h"
#include "GAME2_3.h"
extern nox_window* nox_inventory_window;
extern nox_window* nox_inventory_identify_window;
extern nox_window* nox_inventory_scroll_down_button;
extern nox_window* nox_inventory_scroll_up_button;
extern nox_window* nox_inventory_current_weapon_window;
extern nox_window_yyy nox_windows_arr_1093036[7];
extern uintptr_t dword_5d4594_1062480;
extern uint32_t dword_5d4594_1062512;
extern nox_inventory_cell_t nox_client_inventory_grid_1050020[NOX_INVENTORY_CELLS_MAX];

// E2E raster pass: use the same save/setup/tray/restore contract as
// 00463430, with live stock images and live networked inventory cells.
static int nox_e2e_inventory_draw_tray(int x, int y) {
    nox_xxx_wndDraw_49F7F0();
    int clipped = nox_client_copyRect_49F6F0(x, y, 260, 150) != NULL;
    if (clipped) { nox_xxx_guiDrawInventoryTray_4643B0(x, y); }
    sub_49F860();
    return clipped;
}

static void nox_e2e_inventory_grid(int* filled, int* last_row) {
    *filled = 0;
    *last_row = -1;
    // The final native row contains equipped-item slots, not tray storage.
    for (int row = 0; row < NOX_INVENTORY_ROW_COUNT - 1; row++) {
        for (int column = 0; column < NOX_INVENTORY_COL_COUNT; column++) {
            const nox_inventory_cell_t* cell = &nox_client_inventory_grid_1050020[row + NOX_INVENTORY_ROW_COUNT * column];
            if (cell->field_0 && cell->field_140) {
                (*filled)++;
                *last_row = row;
            }
        }
    }
}

static nox_drawable* nox_e2e_inventory_cell_drawable(int column, int row) {
    if (column < 0 || column >= NOX_INVENTORY_COL_COUNT || row < 0 || row >= NOX_INVENTORY_ROW_COUNT - 1) {
        return NULL;
    }
    const nox_inventory_cell_t* cell = &nox_client_inventory_grid_1050020[row + NOX_INVENTORY_ROW_COUNT * column];
    return cell->field_140 ? cell->field_0 : NULL;
}

static int nox_e2e_inventory_cell_equipped(int column, int row) {
    if (!nox_e2e_inventory_cell_drawable(column, row)) { return 0; }
    return nox_client_inventory_grid_1050020[row + NOX_INVENTORY_ROW_COUNT * column].field_132 != 0;
}

static nox_window* nox_e2e_inventory_weapon_window(int alternate) {
    return alternate ? nox_inventory_current_weapon_window : nox_windows_arr_1093036[4].win;
}

static nox_drawable* nox_e2e_inventory_weapon_drawable(int alternate) {
    if (!alternate) { return sub_4615C0(); }
    const nox_inventory_cell_t* cell = (const nox_inventory_cell_t*)dword_5d4594_1062480;
    return cell ? cell->field_0 : NULL;
}

static int nox_e2e_inventory_draw_weapon(int alternate) {
    nox_window* win = nox_e2e_inventory_weapon_window(alternate);
    if (!win) { return 0; }
    return alternate ? sub_4625D0(win, &win->draw_data) : sub_465D50_draw(win);
}
*/
import "C"

import (
	"image"
	"unsafe"

	"github.com/opennox/opennox/v1/client"
	"github.com/opennox/opennox/v1/client/gui"
)

func InventoryWindow() *gui.Window {
	return AsWindowP(unsafe.Pointer(C.nox_inventory_window))
}

func InventoryIdentifyWindow() *gui.Window {
	return AsWindowP(unsafe.Pointer(C.nox_inventory_identify_window))
}

func InventoryScrollButton(down bool) *gui.Window {
	if down {
		return AsWindowP(unsafe.Pointer(C.nox_inventory_scroll_down_button))
	}
	return AsWindowP(unsafe.Pointer(C.nox_inventory_scroll_up_button))
}

func InventoryDrawTrayClipped(pos image.Point) bool {
	return C.nox_e2e_inventory_draw_tray(C.int(pos.X), C.int(pos.Y)) != 0
}

func InventoryGridMetrics() (columns, rows, filled, lastRow int) {
	var cfilled, clast C.int
	C.nox_e2e_inventory_grid(&cfilled, &clast)
	return C.NOX_INVENTORY_COL_COUNT, C.NOX_INVENTORY_ROW_COUNT, int(cfilled), int(clast)
}

func InventoryScrollOffset() int {
	return int(int32(C.dword_5d4594_1062512))
}

func InventoryCellDrawable(column, row int) *client.Drawable {
	return asDrawable(C.nox_e2e_inventory_cell_drawable(C.int(column), C.int(row)))
}

func InventoryCellEquipped(column, row int) bool {
	return C.nox_e2e_inventory_cell_equipped(C.int(column), C.int(row)) != 0
}

func InventoryWeaponWindow(alternate bool) *gui.Window {
	return AsWindowP(unsafe.Pointer(C.nox_e2e_inventory_weapon_window(C.int(bool2int(alternate)))))
}

func InventoryWeaponDrawable(alternate bool) *client.Drawable {
	return asDrawable(C.nox_e2e_inventory_weapon_drawable(C.int(bool2int(alternate))))
}

func InventoryDrawWeapon(alternate bool) bool {
	return C.nox_e2e_inventory_draw_weapon(C.int(bool2int(alternate))) != 0
}
