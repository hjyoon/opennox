package legacy

/*
#include "GAME1.h"
#include "GAME1_3.h"
#include "GAME2.h"
#include "client__io__win95__focus.h"
#include "vardefs.h"
extern uint32_t dword_5d4594_251708, dword_5d4594_251712;
extern uint32_t dword_5d4594_251716, dword_5d4594_251720;
*/
import "C"

import (
	"unsafe"

	"github.com/opennox/opennox/v1/common/memmap"
)

func startupPauseDefaultCell4134D0(field int) unsafe.Pointer {
	if field == 4 {
		return memmap.PtrOff(0x5D4594, 251740)
	}
	switch int32(field) {
	case 0:
		return unsafe.Pointer(&C.dword_5d4594_251708)
	case 1:
		return unsafe.Pointer(&C.dword_5d4594_251712)
	case 2:
		return unsafe.Pointer(&C.dword_5d4594_251716)
	case 3:
		return unsafe.Pointer(&C.dword_5d4594_251720)
	case 5:
		return unsafe.Pointer(&C.dword_5d4594_251744)
	default:
		return nil
	}
}
func startupPauseDefaultBuffer4134D0() unsafe.Pointer { return memmap.PtrOff(0x5D4594, 251636) }
func startupPauseCallback4134D0(kind uint32) unsafe.Pointer {
	if kind == 1 {
		return unsafe.Pointer(C.sub_413900)
	}
	return unsafe.Pointer(C.sub_4138E0)
}
func startupPauseString4134D0(kind uint32) unsafe.Pointer {
	if kind == 1 {
		return unsafe.Pointer(&C.nox_startup_pause_names[0][0])
	}
	return unsafe.Pointer(&C.nox_startup_pause_names[1][0])
}
func startupPauseDefaultService4134D0(site uint32, args []startupPauseArg4134D0) uint32 {
	switch site {
	case 0x4137f3:
		sub_43DBE0()
	case 0x4138e0:
		return uint32(nox_xxx_checkGameFlagPause_413A50())
	case 0x413900:
		return uint32(nox_video_inFadeTransition_44E0D0())
	case 0x4138ec, 0x413914:
		sub_413A00(int32(args[0].Word))
	case 0x41392e, 0x413941:
		kind := args[0].Word
		callback := C.nox_focus_callback(startupPauseCallback4134D0(kind))
		name := (*C.nox_focus_name)(startupPauseString4134D0(kind))
		sub_42EBB0(C.uint(kind), callback, 0, name)
	default:
		panic("unknown startup/pause service")
	}
	return 0 // void services have no original physical EAX contract.
}
