package legacy

/*
#include "GAME4_1.h"
#include "server__system__trade.h"

*/
import "C"

import (
	"unsafe"

	"github.com/opennox/opennox/v1/legacy/common/alloc"
	"github.com/opennox/opennox/v1/server"
)

var legacyTradePools, _ = alloc.New(C.nox_legacy_trade_pools_t{})

//export nox_legacy_trade_pool_storage
func nox_legacy_trade_pool_storage() *C.nox_legacy_trade_pools_t { return legacyTradePools }

func releaseLegacyTradeItems(head *server.TradeItem, pool *alloc.Class, freeObject func(*server.Object) int) {
	for node := head; node != nil && pool.HasActiveObject(unsafe.Pointer(node)); {
		next := node.Field8
		if node.Item0 != nil && freeObject != nil {
			freeObject(node.Item0)
		}
		pool.FreeObjectFirst(unsafe.Pointer(node))
		node = next
	}
}

//export sub_50F6B0
func sub_50F6B0(ptr unsafe.Pointer) C.int {
	head := (*server.TradeItem)(ptr)
	if head == nil {
		return 0
	}
	world := game4TradeRuntime().world
	if !world.ReleaseTradeOfferNodesNative50F6B0(head) && unsafe.Sizeof(uintptr(0)) == 4 {
		releaseLegacyTradeItems(head, alloc.AsClass(legacyTradePools.items), nil)
	}
	return 0
}

//export sub_510000
func sub_510000(ptr unsafe.Pointer) {
	session := (*server.TradeSession)(ptr)
	if session == nil {
		return
	}
	world := game4TradeRuntime().world
	if world.ReleaseTradeSessionNative510000(session) || unsafe.Sizeof(uintptr(0)) != 4 {
		return
	}
	sessionPool := alloc.AsClass(legacyTradePools.sessions)
	if !sessionPool.HasActiveObject(ptr) {
		return
	}
	itemPool := alloc.AsClass(legacyTradePools.items)
	releaseLegacyTradeItems(session.Field20, itemPool, world.Objs.FreeObject)
	for _, gold := range [...]*server.Object{session.Field48, session.Field52} {
		if gold != nil {
			world.Objs.FreeObject(gold)
		}
	}
	releaseLegacyTradeItems(session.Field32, itemPool, nil)
	releaseLegacyTradeItems(session.Field36, itemPool, nil)
	if next := session.Field56; next != nil {
		next.Field60 = session.Field60
	}
	if previous := session.Field60; previous != nil {
		previous.Field56 = session.Field56
	}
	if legacyTradePools.head_word == C.uint32_t(uintptr(ptr)) {
		legacyTradePools.head_word = C.uint32_t(uintptr(unsafe.Pointer(session.Field56)))
	}
	sessionPool.FreeObjectFirst(ptr)
}
