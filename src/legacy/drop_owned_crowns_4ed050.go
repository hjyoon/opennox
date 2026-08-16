package legacy

/*
#include "drop_owned_crowns_4ed050.h"
*/
import "C"

import (
	"github.com/opennox/libs/types"

	"github.com/opennox/opennox/v1/common/memmap"
	"github.com/opennox/opennox/v1/server"
)

const dropOwnedCrownsTypeCacheOffset4ED050 = uintptr(1568248)

func dropOwnedCrownsRuntime4ED050(s *server.Server) server.DropOwnedCrownsRuntime4ED050 {
	return server.DropOwnedCrownsRuntime4ED050{
		LoadCrownTypeCache: func() uint32 {
			return *memmap.PtrUint32(0x5D4594, dropOwnedCrownsTypeCacheOffset4ED050)
		},
		LookupCrownType: func() uint32 {
			return uint32(s.Types.IndByID("Crown"))
		},
		StoreCrownTypeCache: func(value uint32) {
			*memmap.PtrUint32(0x5D4594, dropOwnedCrownsTypeCacheOffset4ED050) = value
		},
		DropCrown: func(owner, crown *server.Object, position *types.Pointf) uint32 {
			return uint32(objectDropDispatchCall4ED790(owner, crown, position))
		},
	}
}

//export sub_4ED050
func sub_4ED050(owner, target *C.nox_object_t) {
	s := GetServer().S()
	s.DropOwnedCrowns4ED050(
		asObjectS((*nox_object_t)(owner)),
		asObjectS((*nox_object_t)(target)),
		dropOwnedCrownsRuntime4ED050(s),
	)
}
