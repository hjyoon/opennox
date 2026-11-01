package legacy

/*
#include <stdlib.h>
#include "defs.h"
*/
import "C"

import (
	"unsafe"

	"github.com/opennox/opennox/v1/legacy/common/alloc"
	"github.com/opennox/opennox/v1/server"
)

func playerBotCreateAllocate4FA700() *server.MonsterUpdateData {
	size := unsafe.Sizeof(server.MonsterUpdateData{})
	p := C.calloc(1, C.size_t(size))
	if p == nil {
		return nil
	}
	alloc.AdoptCAllocation(p, size)
	return (*server.MonsterUpdateData)(p)
}

func playerBotCreateRuntime4FA700() server.PlayerBotCreateRuntime4FA700 {
	if rt := playerBotCreateRuntimeOverride4FA700; rt != nil {
		return *rt
	}
	return server.PlayerBotCreateRuntime4FA700{
		Allocate: playerBotCreateAllocate4FA700,
		TypeID: func(name string) uint32 {
			return uint32(GetServer().S().Types.IndByID(name))
		},
		Definition: func(id uint32) *server.MonsterDef {
			return Nox_xxx_monsterDefByTT_517560(int(id))
		},
		FPS: func() uint32 {
			return GetServer().S().TickRate()
		},
		Random: func(min, max int32) int32 {
			return int32(GetServer().S().Rand.Other.Int(int(min), int(max)))
		},
	}
}

//export nox_xxx_playerBotCreate_4FA700
func nox_xxx_playerBotCreate_4FA700(unit *nox_object_t) C.int {
	return C.int(asObjectS(unit).PlayerBotCreate4FA700(playerBotCreateRuntime4FA700()))
}
