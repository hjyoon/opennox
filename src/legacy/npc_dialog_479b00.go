package legacy

/*
#include "GAME2_2.h"
extern uint32_t dword_5d4594_1047520;
extern nox_window* dword_5d4594_1123524;
*/
import "C"
import (
	"unsafe"

	"github.com/opennox/opennox/v1/client/gui"
	"github.com/opennox/opennox/v1/common/memmap"
	"github.com/opennox/opennox/v1/common/ntype"
	"github.com/opennox/opennox/v1/common/sound"
	"github.com/opennox/opennox/v1/internal/netlist"
	"github.com/opennox/opennox/v1/legacy/common/alloc"
)

//export nox_xxx_guiDialog_479B00
func nox_xxx_guiDialog_479B00(windowArg int32, event int32, buttonp *int32, reservedArg int32) int32 {
	return int32(gui.NPCDialogProc479B00(int(event), (*gui.Window)(unsafe.Pointer(buttonp)), npcDialogRuntime479B00()))
}

func npcDialogRuntime479B00() gui.NPCDialogRuntime479B00 {
	return gui.NPCDialogRuntime479B00{
		Paused: Sub_45D9B0,
		ClickSound: func() {
			Nox_xxx_clientPlaySoundSpecial_452D80(sound.ID(766), 100)
		},
		Done: func() { C.sub_479950() },
		Repeat: func() {
			// Use the same native pointer slot as guiOpenNPCDialogRaw. The
			// packed PE32 DWORD is neither this slot nor a host-width pointer.
			file := (*byte)(*memmap.PtrPtr(0x5D4594, 1115312))
			Dialogs.PlayFile(alloc.GoString(file), 100)
		},
		Answer: func(choice byte) {
			*memmap.PtrUint8(0x5D4594, 1123516) = choice
			packet := [3]byte{0xD0, 0x02, choice}
			GetServer().S().NetList.AddToMsgListCli(ntype.PlayerInd(31), netlist.Kind0, packet[:])
		},
	}
}

// NPCDialogProc479B00 is the callback installed by the stock Dialog.wnd
// constructor. It is also used by native CGo/window regression tests.
func NPCDialogProc479B00() unsafe.Pointer { return C.nox_xxx_guiDialog_479B00 }

func npcDialogPauseSlot479B00() *uint32 { return (*uint32)(&C.dword_5d4594_1047520) }

func npcDialogWindowSlot479B00() **gui.Window {
	return (**gui.Window)(unsafe.Pointer(&C.dword_5d4594_1123524))
}
