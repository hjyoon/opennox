package legacy

/*
#include "defs.h"
*/
import "C"

import (
	"github.com/opennox/libs/types"
	"github.com/opennox/opennox/v1/common/memmap"
	"github.com/opennox/opennox/v1/server"
)

var arachnaphobiaRawRuntime53DA60 = func() server.ArachnaphobiaRawRuntime53DA60 {
	return server.ArachnaphobiaRawRuntime53DA60{
		Frame:              func() uint32 { return GetServer().S().Frame() },
		FPS:                func() uint32 { return GetServer().S().TickRate() },
		LoadTypeIndex:      func() uint32 { return *memmap.PtrUint32(0x5D4594, 2488708) },
		StoreTypeIndex:     func(v uint32) { *memmap.PtrUint32(0x5D4594, 2488708) = v },
		TypeIndexByID:      func(id string) uint32 { return uint32(GetServer().S().Types.IndByID(id)) },
		NewObjectByTypeInd: func(index int32) *server.Object { return GetServer().S().NewObjectByTypeInd(int(index)) },
		CreateAt:           func(o, owner *server.Object, p types.Pointf) { GetServer().CreateObjectAt(o, owner, p) },
		RandomInt:          func(a, b int) int { return GetServer().S().Rand.Logic.IntClamp(a, b) },
		DelayedDelete:      func(o *server.Object) { GetServer().DelayedDelete(o) },
	}
}

var stateUpdateRawRuntime53DB30 = func() server.StateUpdateRawRuntime53DB30 {
	return server.StateUpdateRawRuntime53DB30{
		Frame:        func() uint32 { return GetServer().S().Frame() },
		FPS:          func() uint32 { return GetServer().S().TickRate() },
		UnsetXStatus: (*server.Object).UnsetXStatus, SetXStatus: (*server.Object).SetXStatus,
		RemoveFromUpdatable: func(o *server.Object) { GetServer().S().Objs.RemoveFromUpdatable(o) },
		DelayedDelete:       func(o *server.Object) { GetServer().DelayedDelete(o) },
	}
}

//export nox_state_update_native
func nox_state_update_native(kind C.int, source *nox_object_t) {
	switch kind {
	case 0:
		server.ArachnaphobiaUpdateRaw53DA60(asObjectS(source), arachnaphobiaRawRuntime53DA60())
	case 1:
		server.BreakUpdateRaw53DB30(asObjectS(source), stateUpdateRawRuntime53DB30())
	case 2:
		server.OpenUpdateRaw53DBB0(asObjectS(source), stateUpdateRawRuntime53DB30())
	case 3:
		server.BreakAndRemoveUpdateRaw53DC30(asObjectS(source), stateUpdateRawRuntime53DB30())
	}
}
