package legacy

/*
#include "GAME2_1.h"
extern nox_inventory_cell_t nox_client_inventory_grid_1050020[NOX_INVENTORY_CELLS_MAX];
extern uintptr_t dword_5d4594_1062480;
extern uint32_t dword_5d4594_1062484;
static void* secondary_grid_base_198(void) { return nox_client_inventory_grid_1050020; }
static void secondary_grid_layout_198(size_t* out) {
    out[0]=sizeof(nox_inventory_cell_t); out[1]=NOX_INVENTORY_CELLS_MAX;
    out[2]=offsetof(nox_inventory_cell_t,field_4);
    out[3]=offsetof(nox_inventory_cell_t,field_132);
    out[4]=offsetof(nox_inventory_cell_t,field_136);
    out[5]=offsetof(nox_inventory_cell_t,field_140);
    out[6]=sizeof(nox_drawable);
    out[7]=offsetof(nox_drawable,field_32);
    out[8]=offsetof(nox_drawable,flags28);
}
static uintptr_t secondary_grid_selected_198(void) { return dword_5d4594_1062480; }
static uint32_t secondary_grid_previous_198(void) { return dword_5d4594_1062484; }
static void secondary_grid_restore_198(uintptr_t selected, uint32_t code) {
    dword_5d4594_1062480=selected; dword_5d4594_1062484=code;
}
static void secondary_grid_prepare_previous_198(int slot) {
    dword_5d4594_1062480=slot<0 ? 0 : (uintptr_t)&nox_client_inventory_grid_1050020[slot];
    dword_5d4594_1062484=0x12345678u;
}
static int secondary_grid_select_198(int slot) {
    return nox_xxx_clientSetAltWeapon_461550(slot<0 ? NULL : &nox_client_inventory_grid_1050020[slot]);
}
*/
import "C"

import "unsafe"

func clientSecondaryGridLayout198() ([]byte, [9]int) {
	var raw [9]C.size_t
	C.secondary_grid_layout_198(&raw[0])
	var layout [9]int
	for i, value := range raw {
		layout[i] = int(value)
	}
	return unsafe.Slice((*byte)(C.secondary_grid_base_198()), layout[0]*layout[1]), layout
}

func clientSecondaryGridGlobals198() (uintptr, uint32) {
	return uintptr(C.secondary_grid_selected_198()), uint32(C.secondary_grid_previous_198())
}

func clientSecondaryGridRestore198(selected uintptr, code uint32) {
	C.secondary_grid_restore_198(C.uintptr_t(selected), C.uint32_t(code))
}

func clientSecondaryGridPrepare198(slot int) {
	C.secondary_grid_prepare_previous_198(C.int(slot))
}

func clientSecondaryGridSelect198(slot int) int {
	return int(C.secondary_grid_select_198(C.int(slot)))
}
