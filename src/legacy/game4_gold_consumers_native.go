package legacy

/*
#include <stdint.h>
#include "GAME4.h"
static uintptr_t game4_consumer_gold_add(nox_object_t* owner, int32_t delta) {
	return (uintptr_t)nox_xxx_playerAddGold_4FA590(owner, delta);
}
static uintptr_t game4_consumer_gold_sub(nox_object_t* owner, uint32_t amount) {
	return (uintptr_t)nox_xxx_playerSubGold_4FA5D0(owner, amount);
}
*/
import "C"

import (
	"runtime"
	"unsafe"

	"github.com/opennox/opennox/v1/server"
)

type game4TradeRuntimeProvider interface {
	Game4TradeP2PRuntime() server.TradeP2PRuntime50F3A0
	Game4ShopBuyRuntime() server.ShopBuyRuntime5100C0
	Game4ShopSellRuntime() server.ShopSellRuntime5109C0
	Game4ShopRepairRuntime() server.ShopRepairRuntime5108D0
	Game4TradePacketNative(*server.Object, []byte) int
}

// A scoped test override supplies actual server transaction services, never
// substitute gold/transaction algorithms. Tests restore it before returning.
type game4TradeServices struct {
	world  *server.Server
	p2p    func() server.TradeP2PRuntime50F3A0
	buy    func() server.ShopBuyRuntime5100C0
	sell   func() server.ShopSellRuntime5109C0
	repair func() server.ShopRepairRuntime5108D0
	packet func(*server.Object, []byte) int
}

var game4TradeServicesForTest *game4TradeServices

func game4TradeRuntime() game4TradeServices {
	if override := game4TradeServicesForTest; override != nil {
		return *override
	}
	root := GetServer()
	provider, ok := root.(game4TradeRuntimeProvider)
	if !ok {
		panic("GAME4 native trade runtime is not connected")
	}
	return game4TradeServices{
		world: root.S(), p2p: provider.Game4TradeP2PRuntime,
		buy: provider.Game4ShopBuyRuntime, sell: provider.Game4ShopSellRuntime,
		repair: provider.Game4ShopRepairRuntime, packet: provider.Game4TradePacketNative,
	}
}

func game4GoldAdd(owner *server.Object, amount uint32) uintptr {
	result := C.game4_consumer_gold_add(asObjectC(owner), C.int32_t(amount))
	runtime.KeepAlive(owner)
	return uintptr(result)
}

func game4GoldSub(owner *server.Object, amount uint32) uintptr {
	result := C.game4_consumer_gold_sub(asObjectC(owner), C.uint32_t(amount))
	runtime.KeepAlive(owner)
	return uintptr(result)
}

func game4GoldGet(owner *server.Object) uint32 {
	result := uint32(C.nox_xxx_playerGetGold_4FA6B0(asObjectC(owner)))
	runtime.KeepAlive(owner)
	return result
}

func game4GoldReturn50F7F0(owner, gold *server.Object) uintptr {
	if uint8(owner.ObjClass)&4 == 0 {
		return uintptr(unsafe.Pointer(owner))
	}
	return game4GoldAdd(owner, gold.InitDataGold().Amount)
}

func game4TradeOfferValue50FD20(session *server.TradeSession, owner *server.Object) uint32 {
	var gold *server.Object
	var head *server.TradeItem
	if session.Field8 == owner {
		gold, head = session.Field48, session.Field32
	} else {
		gold, head = session.Field52, session.Field36
	}
	result := gold.InitDataGold().Amount
	for node := head; node != nil; node = node.Field8 {
		result += node.Cost4
	}
	return result
}

// All pointer fields come from the native TradeSession layout. The return
// word may be a scalar EAX residue or a genuine native CRC/Gold field address;
// it stays uintptr until the tiny C adapter, never a manufactured Go pointer.
func game4TradeReprice50FB90(session *server.TradeSession) uintptr {
	first := session.Field8
	result := uintptr(uint32(first.ObjClass) & 4)
	if result == 4 && uint8(session.Field12.ObjClass)&4 == 4 {
		return result
	}
	goldA := session.Field48.InitDataGold()
	if result != 0 {
		game4GoldAdd(first, goldA.Amount)
	}
	goldA.Amount = 0
	second := session.Field12
	goldB := session.Field52.InitDataGold()
	if uint8(second.ObjClass)&4 != 0 {
		game4GoldAdd(second, goldB.Amount)
	}
	goldB.Amount = 0
	valueA := game4TradeOfferValue50FD20(session, session.Field8)
	second = session.Field12
	session.Field40 = valueA
	valueB := game4TradeOfferValue50FD20(session, second)
	result = uintptr(valueB)
	valueA = session.Field40
	session.Field44 = valueB
	if valueB < valueA {
		delta := valueA - valueB
		owner := session.Field12
		if uint8(owner.ObjClass)&4 != 0 {
			available := game4GoldGet(owner)
			if available < delta {
				delta = available
			}
			session.Field44 += delta
			session.Field52.InitDataGold().Amount += delta
			result = game4GoldSub(session.Field12, delta)
		} else {
			session.Field44 = valueA
			amount := &session.Field52.InitDataGold().Amount
			*amount += delta
			result = uintptr(unsafe.Pointer(amount))
		}
	} else if valueB > valueA {
		delta := valueB - valueA
		owner := session.Field8
		if uint8(owner.ObjClass)&4 != 0 {
			available := game4GoldGet(owner)
			if available < delta {
				delta = available
			}
			session.Field40 += delta
			session.Field48.InitDataGold().Amount += delta
			result = game4GoldSub(session.Field8, delta)
		} else {
			session.Field40 += delta
			amount := &session.Field48.InitDataGold().Amount
			*amount += delta
			result = uintptr(unsafe.Pointer(amount))
		}
	}
	runtime.KeepAlive(session)
	return result
}
