package legacy

/*
#include "GAME2_1.h"
#include "GAME2_3.h"
extern nox_window_yyy nox_windows_arr_1093036[7];
extern uint32_t dword_5d4594_1062512;
*/
import "C"

import (
	"image"

	"github.com/opennox/opennox/v1/client"
	"github.com/opennox/opennox/v1/client/gui"
)

func InventoryWindow() *gui.Window {
	return asWindow(inventoryWindows.panel)
}

func InventoryIdentifyWindow() *gui.Window {
	return asWindow(inventoryWindows.identify)
}

func InventoryScrollButton(down bool) *gui.Window {
	if down {
		return asWindow(inventoryWindows.scrollDown)
	}
	return asWindow(inventoryWindows.scrollUp)
}

func InventoryDrawTrayClipped(pos image.Point) bool {
	x, y := C.int(pos.X), C.int(pos.Y)
	nox_xxx_wndDraw_49F7F0()
	clipped := currentDrawClipRect(x, y, 260, 150) != 0
	if clipped {
		nox_xxx_guiDrawInventoryTray_4643B0(x, y)
	}
	sub_49F860()
	return clipped
}

func InventoryGridMetrics() (columns, rows, filled, lastRow int) {
	lastRow = -1
	// The final native row holds equipped items outside the tray.
	for row := range inventoryRows - 1 {
		for column := range inventoryColumns {
			cell := inventoryGridCell(row, column)
			if cell.field_0 != nil && cell.field_140 != 0 {
				filled++
				lastRow = row
			}
		}
	}
	return inventoryColumns, inventoryRows, filled, lastRow
}

func InventoryScrollOffset() int {
	return int(int32(C.dword_5d4594_1062512))
}

func InventoryCellDrawable(column, row int) *client.Drawable {
	if cell := inventoryTrayCell(column, row); cell != nil {
		return asDrawable(cell.field_0)
	}
	return nil
}

func InventoryCellEquipped(column, row int) bool {
	cell := inventoryTrayCell(column, row)
	return cell != nil && cell.field_132 != 0
}

func inventoryTrayCell(column, row int) *C.nox_inventory_cell_t {
	column, row = int(int32(column)), int(int32(row))
	if column < 0 || column >= inventoryColumns || row < 0 || row >= inventoryRows-1 {
		return nil
	}
	cell := inventoryGridCell(row, column)
	if cell.field_140 == 0 || cell.field_0 == nil {
		return nil
	}
	return cell
}

func InventoryWeaponWindow(alternate bool) *gui.Window {
	return asWindow(inventoryWeaponWindow(alternate))
}

func inventoryWeaponWindow(alternate bool) *nox_window {
	if alternate {
		return inventoryWindows.weapon
	}
	return C.nox_windows_arr_1093036[4].win
}

func InventoryWeaponDrawable(alternate bool) *client.Drawable {
	if !alternate {
		return asDrawable(sub_4615C0())
	}
	if cell := (*C.nox_inventory_cell_t)(globalStateNativePointer(2)); cell != nil {
		return asDrawable(cell.field_0)
	}
	return nil
}

func InventoryDrawWeapon(alternate bool) bool {
	win := inventoryWeaponWindow(alternate)
	if win == nil {
		return false
	}
	if alternate {
		return sub_4625D0(win, &win.draw_data) != 0
	}
	return sub_465D50_draw(win) != 0
}
