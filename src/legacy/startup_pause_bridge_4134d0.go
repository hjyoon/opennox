package legacy

/*
#include "GAME1.h"
#include "GAME1_3.h"
#include "GAME2.h"
#include "client__io__win95__focus.h"
extern uint32_t dword_5d4594_251708, dword_5d4594_251712;
extern uint32_t dword_5d4594_251716, dword_5d4594_251720, dword_5d4594_251744;
static uint32_t* startup_pause_cell_4134d0(int field) {
    switch(field) {
    case 0:return &dword_5d4594_251708;
    case 1:return &dword_5d4594_251712;
    case 2:return &dword_5d4594_251716;
    case 3:return &dword_5d4594_251720;
    case 5:return &dword_5d4594_251744;
    default:return 0;
    }
}
static void* startup_pause_callback_4134d0(uint32_t kind) {
    return kind==1 ? (void*)sub_413900 : (void*)sub_4138E0;
}
static const char startup_pause_name_one_4134d0[]="Pause";
static const char startup_pause_name_two_4134d0[]="Pause";
static const char* startup_pause_name_4134d0(uint32_t kind) {
    return kind==1 ? startup_pause_name_one_4134d0 : startup_pause_name_two_4134d0;
}
static void startup_pause_register_4134d0(uint32_t kind) {
    void (*callback)(int)=kind==1 ? sub_413900 : sub_4138E0;
    sub_42EBB0(kind,callback,0,startup_pause_name_4134d0(kind));
}
*/
import "C"
import (
	"github.com/opennox/opennox/v1/common/memmap"
	"unsafe"
)

func startupPauseDefaultCell4134D0(field int) unsafe.Pointer {
	if field == 4 {
		return memmap.PtrOff(0x5D4594, 251740)
	}
	return unsafe.Pointer(C.startup_pause_cell_4134d0(C.int(field)))
}
func startupPauseDefaultBuffer4134D0() unsafe.Pointer { return memmap.PtrOff(0x5D4594, 251636) }
func startupPauseCallback4134D0(kind uint32) unsafe.Pointer {
	return C.startup_pause_callback_4134d0(C.uint32_t(kind))
}
func startupPauseString4134D0(kind uint32) unsafe.Pointer {
	return unsafe.Pointer(C.startup_pause_name_4134d0(C.uint32_t(kind)))
}
func startupPauseDefaultService4134D0(site uint32, args []startupPauseArg4134D0) uint32 {
	switch site {
	case 0x4137f3:
		sub_43DBE0()
	case 0x4138e0:
		return uint32(C.nox_xxx_checkGameFlagPause_413A50())
	case 0x413900:
		return uint32(nox_video_inFadeTransition_44E0D0())
	case 0x4138ec, 0x413914:
		sub_413A00(int32(args[0].Word))
	case 0x41392e, 0x413941:
		C.startup_pause_register_4134d0(C.uint32_t(args[0].Word))
	default:
		panic("unknown startup/pause service")
	}
	return 0 // void services have no original physical EAX contract.
}
