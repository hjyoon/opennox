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

//export nox_xxx_updateArachnaphobia_53DA60
func nox_xxx_updateArachnaphobia_53DA60(source *nox_object_t) {
	server.ArachnaphobiaUpdateRaw53DA60(asObjectS(source), arachnaphobiaRawRuntime53DA60())
}

//export nox_xxx_updateBreak_53DB30
func nox_xxx_updateBreak_53DB30(source *nox_object_t) {
	server.BreakUpdateRaw53DB30(asObjectS(source), stateUpdateRawRuntime53DB30())
}

//export nox_xxx_updateOpen_53DBB0
func nox_xxx_updateOpen_53DBB0(source *nox_object_t) {
	server.OpenUpdateRaw53DBB0(asObjectS(source), stateUpdateRawRuntime53DB30())
}

//export nox_xxx_updateBreakAndRemove_53DC30
func nox_xxx_updateBreakAndRemove_53DC30(source *nox_object_t) {
	server.BreakAndRemoveUpdateRaw53DC30(asObjectS(source), stateUpdateRawRuntime53DB30())
}
