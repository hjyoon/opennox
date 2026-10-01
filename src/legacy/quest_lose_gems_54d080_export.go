package legacy

/*
#include "GAME5.h"
extern uint32_t dword_5d4594_2491676;
*/
import "C"

import (
	"unsafe"

	"github.com/opennox/opennox/v1/common/memmap"
	"github.com/opennox/opennox/v1/server"
)

func questGemTypeCache54D080() [3]*uint32 {
	return [3]*uint32{
		(*uint32)(unsafe.Pointer(&C.dword_5d4594_2491676)),
		memmap.PtrUint32(0x5D4594, 2491680),
		memmap.PtrUint32(0x5D4594, 2491684),
	}
}

func questLoseGemsCEntry54D080(unit *server.Object) { C.sub_54D080(asObjectC(unit)) }

//export nox_server_questLoseGems_native_54D080
func nox_server_questLoseGems_native_54D080(unit *nox_object_t) {
	outer := GetServer()
	outer.S().QuestLoseGems54D080(asObjectS(unit), questGemTypeCache54D080(), outer.DelayedDelete, func(unit *server.Object, amount int32) {
		Nox_xxx_playerAddGold_4FA590(unit, int(amount))
	})
}
