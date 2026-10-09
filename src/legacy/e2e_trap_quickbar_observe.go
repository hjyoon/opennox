package legacy

/*
#include "GAME2.h"
#include "memmap.h"
extern uint32_t dword_5d4594_1049484;
extern nox_window* dword_5d4594_1049500;

static nox_window* native_e2e_trap_control(int control) {
	if (control == 0) return dword_5d4594_1049500;
	nox_window* root = nox_quickbar_root(getMemAt(0x5D4594, 1047940));
	if (!root || (control != 3 && control != 4)) return NULL;
	for (nox_window* win = root->field_100; win; win = win->prev) {
		if (win->field_92 == (uint32_t)control) return win;
	}
	return NULL;
}
*/
import "C"

import (
	"encoding/binary"
	"unsafe"

	"github.com/opennox/opennox/v1/client/gui"
	"github.com/opennox/opennox/v1/common/memmap"
)

// ClientTrapQuickbarSnapshot only reads the packed numeric entries, including
// reserved slots/flag bytes. Native window pointers stay in the sidecar.
func ClientTrapQuickbarSnapshot() (entries [25][2]uint32, row int, open bool) {
	data := unsafe.Slice((*byte)(memmap.PtrOff(0x5D4594, 1047940)), 201)
	for i := range entries {
		entries[i][0] = binary.LittleEndian.Uint32(data[8*i:])
		entries[i][1] = binary.LittleEndian.Uint32(data[8*i+4:])
	}
	return entries, int(data[200]), C.dword_5d4594_1049484 != 0
}

func ClientTrapQuickbarRoot() *gui.Window {
	return AsWindowP(unsafe.Pointer(C.nox_quickbar_root(memmap.PtrOff(0x5D4594, 1047940))))
}

func ClientTrapQuickbarButton(slot int) *gui.Window {
	return AsWindowP(unsafe.Pointer(C.nox_quickbar_button(memmap.PtrOff(0x5D4594, 1047940), C.int(slot))))
}

func ClientTrapQuickbarControl(control int) *gui.Window {
	return AsWindowP(unsafe.Pointer(C.native_e2e_trap_control(C.int(control))))
}

func ClientQuickbarButton(slot int) *gui.Window {
	return AsWindowP(unsafe.Pointer(C.nox_quickbar_button(quickBarBase45DA50(), C.int(slot))))
}
