package legacy

/*
#include "GAME4_1.h"

extern void* nox_alloc_tradeSession_2386492;
*/
import "C"

import (
	"unsafe"

	"github.com/opennox/opennox/v1/common/memmap"
	"github.com/opennox/opennox/v1/legacy/common/alloc"
	"github.com/opennox/opennox/v1/server"
)

//export nox_xxx_shopCancelSession_510DC0
func nox_xxx_shopCancelSession_510DC0(ptr unsafe.Pointer) {
	if ptr == nil {
		return
	}
	session := (*server.TradeSession)(ptr)
	if GetServer != nil {
		if srv := GetServer(); srv != nil && srv.ShopCancelSessionNative510DC0(session) {
			return
		}
	}
	if cgoABIPointerSize != 4 || !alloc.AsClass(C.nox_alloc_tradeSession_2386492).HasActiveObject(ptr) {
		return
	}
	if session.Field16 != 0 {
		nox_xxx_shopExit_50F4C0(ptr)
	} else {
		sub_50F3A0(ptr)
	}
}

// questShopSessionSlotPE32 is used only when a native pointer fits the legacy slot.
func questShopSessionSlotPE32(index int) *unsafe.Pointer {
	return memmap.PtrPtr(0x5D4594, 2386364+4*uintptr(index))
}

//export sub_510E20
func sub_510E20(playerIndex C.int) {
	index := int(playerIndex)
	if index < 0 || index >= 32 {
		return
	}
	cleared := false
	if GetServer != nil {
		if srv := GetServer(); srv != nil {
			cleared = srv.ClearQuestShopSessionNative510E20(index)
		}
	}
	if cgoABIPointerSize != 4 {
		return
	}
	slot := questShopSessionSlotPE32(index)
	if !cleared {
		sub_510000(*slot)
	}
	*slot = nil
}
