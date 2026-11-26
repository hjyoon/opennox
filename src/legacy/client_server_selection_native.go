package legacy

/*
#ifdef _WIN32
#include <winsock.h>
#else
#include <arpa/inet.h>
#endif
#include "GAME1_2.h"
#include "GAME2_2.h"

extern uint32_t dword_5d4594_815056;
extern nox_gui_animation* nox_wnd_xxx_815040;
*/
import "C"

import (
	"unsafe"

	"github.com/opennox/opennox/v1/common/memmap"
)

//export nox_xxx_copyServerIPAndPort_431790
func nox_xxx_copyServerIPAndPort_431790(addr *C.char) *C.char {
	if addr == nil {
		return nil
	}
	dst := memmap.PtrT[[23]C.char](0x5D4594, 806060)
	for i := range dst {
		value := *(*C.char)(unsafe.Add(unsafe.Pointer(addr), i))
		if value == 0 {
			clear(dst[i:])
			break
		}
		dst[i] = value
	}
	return &dst[0]
}

//export nox_client_setConnError_43AFA0
func nox_client_setConnError_43AFA0(code C.int) {
	Set_nox_client_connError_814552(int(code))
	sub_43AF90(2)
}

func selectedLegacyServer() *Nox_gui_server_ent_t {
	if C.dword_5d4594_815056 == 0 {
		return nil
	}
	return (*Nox_gui_server_ent_t)(Get_dword_5d4594_814624())
}

//export nox_client_getServerAddr_43B300
func nox_client_getServerAddr_43B300() C.uint {
	if server := selectedLegacyServer(); server != nil {
		return C.uint(C.inet_addr(&server.C().addr[0]))
	}
	return 0
}

//export nox_client_getServerPort_43B320
func nox_client_getServerPort_43B320() C.int {
	if C.dword_5d4594_815056 == 0 {
		return 0
	}
	return C.int(memmap.Uint32(0x5D4594, 814604))
}

//export sub_43B340
func sub_43B340() C.int {
	if server := selectedLegacyServer(); server != nil {
		return C.int(server.Flags())
	}
	return 0
}

//export sub_43B460
func sub_43B460() C.int {
	C.sub_438370()
	C.nox_wnd_xxx_815040.fnc_done_out = (*[0]byte)(C.sub_43B490)
	nox_xxx_wnd_46C6E0(wolWindows.serverInfo)
	return C.int(nox_xxx_wnd_46ABB0_go(wolWindows.mapPanel, 0))
}

//export sub_43B490
func sub_43B490() C.int {
	if nox_game_getStateCode_43BE10() == 1700 {
		return C.sub_438330()
	}
	nox_window_set_hidden(wolWindows.root, 1)
	nox_window_set_hidden(wolWindows.serverInfo, 1)
	C.nox_client_setCursorType_477610(0)
	return 1
}
