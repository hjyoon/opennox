package legacy

/*
#include "GAME4_1.h"

extern void* nox_alloc_tradeSession_2386492;
*/
import "C"

import (
	"unsafe"

	noxflags "github.com/opennox/opennox/v1/common/flags"
	"github.com/opennox/opennox/v1/legacy/common/alloc"
	"github.com/opennox/opennox/v1/server"
)

//export nox_xxx_shopExit_50F4C0
func nox_xxx_shopExit_50F4C0(ptr unsafe.Pointer) {
	if ptr == nil {
		return
	}
	world := game4TradeRuntime().world
	session := (*server.TradeSession)(ptr)
	if world.CloseShopNative50F4C0(session, Nox_xxx_unitUnFreeze_4E7A60) || unsafe.Sizeof(uintptr(0)) != 4 {
		return
	}
	pool := alloc.AsClass(C.nox_alloc_tradeSession_2386492)
	if !pool.HasActiveObject(ptr) {
		return
	}
	index, notified := world.NotifyShopExit50F4C0(session, Nox_xxx_unitUnFreeze_4E7A60)
	if !pool.HasActiveObject(ptr) {
		return
	}
	if notified && index >= 0 && index < 32 && noxflags.HasGame(noxflags.GameModeQuest) {
		*questShopSessionSlotPE32(index) = ptr
		return
	}
	sub_510000(ptr)
}
