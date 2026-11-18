package legacy

/*
#include "server__system__trade.h"
*/
import "C"

import (
	"slices"

	noxflags "github.com/opennox/opennox/v1/common/flags"
	"github.com/opennox/opennox/v1/common/memmap"
	"github.com/opennox/opennox/v1/server"
)

const (
	shopTypeCacheBase = 0x5D4594
	questShopTypeOff  = 2386520
	gemShopTypeOff    = 2386536
)

func shopItemTypeMatches(item *server.Object, cache []uint32) bool {
	if cache[0] == 0 {
		types := &game4TradeRuntime().world.Types
		cache[0] = uint32(types.IndByID("Diamond"))
		cache[2] = uint32(types.IndByID("Emerald"))
		cache[1] = uint32(types.IndByID("Ruby"))
		if len(cache) == 4 {
			cache[3] = uint32(types.IndByID("AnkhTradable"))
		}
	}
	return slices.Contains(cache, uint32(item.TypeInd))
}

//export sub_510540
func sub_510540(item *nox_object_t) C.int {
	if !noxflags.HasGame(noxflags.GameModeQuest) {
		return 1
	}
	cache := memmap.PtrT[[4]uint32](shopTypeCacheBase, questShopTypeOff)
	if shopItemTypeMatches(asObjectS(item), cache[:]) {
		return 0
	}
	return 1
}

//export sub_5105D0
func sub_5105D0(item *nox_object_t) C.int {
	cache := memmap.PtrT[[3]uint32](shopTypeCacheBase, gemShopTypeOff)
	if shopItemTypeMatches(asObjectS(item), cache[:]) {
		return 1
	}
	return 0
}
