package legacy

/*
#include "GAME1_1.h"
extern uint32_t dword_5d4594_528252, dword_5d4594_528256, dword_5d4594_528260, dword_5d4594_528264;
*/
import "C"

import (
	"fmt"

	"github.com/opennox/opennox/v1/common/memmap"
)

const (
	reconnectWindowSeconds = 3600
	reconnectDelaySeconds  = 120
)

//export sub_41E370
func sub_41E370() C.int {
	C.dword_5d4594_528252 = 0
	C.dword_5d4594_528256 = 0
	C.dword_5d4594_528260 = 0
	C.dword_5d4594_528264 = 0
	return 0
}

//export nox_xxx_reconAttempt_41E390
func nox_xxx_reconAttempt_41E390() C.int {
	if gameFrame()-uint32(C.dword_5d4594_528264) > reconnectWindowSeconds*gameFPS() {
		sub_41E370()
		return sub_41E4B0(1)
	}
	if C.dword_5d4594_528252 == 0 {
		return 0
	}
	if C.dword_5d4594_528256 != 0 {
		return C.int(C.dword_5d4594_528256)
	}
	NetworkLogPrint("RECON: Attempting to re-login")
	sub_40E090()
	if C.nox_xxx_officialStringCmp_41FDE0() != 1 {
		return sub_41E470()
	}
	C.dword_5d4594_528256 = 1
	return 1
}

//export nox_xxx_reconStart_41E400
func nox_xxx_reconStart_41E400() {
	if C.dword_5d4594_528252 == 1 || C.dword_5d4594_528256 == 1 ||
		C.dword_5d4594_528260 != 0 || C.dword_5d4594_528264 != 0 {
		return
	}
	NetworkLogPrint(fmt.Sprintf("RECON: Starting reconnection process frame (%d)", int32(gameFrame())))
	C.dword_5d4594_528252 = 1
	C.dword_5d4594_528256 = 0
	C.dword_5d4594_528264 = C.uint32_t(gameFrame())
	C.dword_5d4594_528260 = C.uint32_t(gameFrame() + reconnectDelaySeconds*gameFPS())
}

//export sub_41E470
func sub_41E470() C.int {
	NetworkLogPrint(fmt.Sprintf("RECON: TryReconnectAgain called on frame (%d)", int32(gameFrame())))
	C.dword_5d4594_528256 = 0
	result := gameFrame() + reconnectDelaySeconds*gameFPS()
	C.dword_5d4594_528260 = C.uint32_t(gameFrame() + reconnectDelaySeconds*gameFPS())
	return C.int(result)
}

//export sub_41E4B0
func sub_41E4B0(value C.int) C.int {
	*memmap.PtrInt32(0x5D4594, 528268) = int32(value)
	return value
}
