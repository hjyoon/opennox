package legacy

/*
#include "GAME2.h"
extern uint32_t dword_5d4594_1049484;
extern nox_window* dword_5d4594_1049500;
*/
import "C"

import (
	"encoding/binary"
	"unsafe"

	"github.com/opennox/opennox/v1/client/gui"
)

// ClientTrapQuickbarSnapshot only reads the packed numeric entries, including
// reserved slots/flag bytes. Native window pointers stay in the sidecar.
func ClientTrapQuickbarSnapshot() (entries [25][2]uint32, row int, open bool) {
	data := unsafe.Slice((*byte)(g2qTrap()), 201)
	for i := range entries {
		entries[i][0] = binary.LittleEndian.Uint32(data[8*i:])
		entries[i][1] = binary.LittleEndian.Uint32(data[8*i+4:])
	}
	return entries, int(data[200]), C.dword_5d4594_1049484 != 0
}

func ClientTrapQuickbarRoot() *gui.Window {
	return asWindow(nox_quickbar_root(g2qTrap()))
}

func ClientTrapQuickbarButton(slot int) *gui.Window {
	return asWindow(nox_quickbar_button(g2qTrap(), C.int(slot)))
}

func ClientTrapQuickbarControl(control int) *gui.Window {
	control = int(int32(control))
	if control == 0 {
		return asWindow(C.dword_5d4594_1049500)
	}
	root := nox_quickbar_root(g2qTrap())
	if root == nil || (control != 3 && control != 4) {
		return nil
	}
	for win := root.field_100; win != nil; win = win.prev {
		if win.field_92 == C.uint32_t(control) {
			return asWindow(win)
		}
	}
	return nil
}

func ClientQuickbarButton(slot int) *gui.Window {
	return asWindow(nox_quickbar_button(quickBarBase45DA50(), C.int(slot)))
}
