package legacy

/*
#include "GAME2.h"
#include "GAME3_1.h"
*/
import "C"

import (
	"encoding/binary"
	"unsafe"

	"github.com/opennox/opennox/v1/client"
	"github.com/opennox/opennox/v1/client/gui"
)

// ClientItemAmountDrawable observes the live quantity dialog without replacing
// its item, callback, viewport, or rendered pixels.
func ClientItemAmountDrawable() *client.Drawable {
	return asDrawable(itemAmountDialog.drawable)
}

// ClientQuickbarSnapshot observes the PE32 numeric entries through the live
// native-width base. Flags includes reserved bytes so E2E can check preservation.
func ClientQuickbarSnapshot() (entries [25][2]uint32, row int, ok bool) {
	base := quickBarBase45DA50()
	if base == nil {
		return entries, 0, false
	}
	data := unsafe.Slice((*byte)(base), quickBarSelected45DA50+1)
	for i := range entries {
		entries[i][0] = binary.LittleEndian.Uint32(data[8*i:])
		entries[i][1] = binary.LittleEndian.Uint32(data[8*i+4:])
	}
	row = int(data[quickBarSelected45DA50])
	return entries, row, row < quickBarRows45DA50
}

// ClientQuickbarNugget observes the real self/other target toggle window.
func ClientQuickbarNugget(slot int) *gui.Window {
	return AsWindowP(unsafe.Pointer(C.nox_quickbar_nugget(quickBarBase45DA50(), C.int(slot))))
}

// ClientP2PTradeWindow observes the trade popup and its active state.
func ClientP2PTradeWindow() (*gui.Window, bool) {
	return asWindow(tradeRoot()), sub_4C12C0() != 0
}

// ClientP2PTradeOffer observes an item drawn by the live trade window.
func ClientP2PTradeOffer(remote bool, index int) (*client.Drawable, uint32, uint32) {
	slot := nox_gui_trade_slot_at(C.int(bool2int(remote)), C.int(index))
	if slot == nil {
		return nil, 0, 0
	}
	return asDrawable(slot.drawable), uint32(slot.count), uint32(slot.total_cost)
}
