package legacy

/*
#include "defs.h"
#include "GAME5.h"
*/
import "C"

import (
	noxflags "github.com/opennox/opennox/v1/common/flags"
	"github.com/opennox/opennox/v1/server"
)

var monsterDieCallbackRawRuntime54A7D0 = func() server.MonsterDieCallbackRawRuntime54A7D0 {
	return server.MonsterDieCallbackRawRuntime54A7D0{
		RandomInt: GetServer().S().Rand.Logic.IntClamp,
		DropItem: func(unit *server.Object, drop server.MonsterDieDrop54A390) {
			if noxflags.HasGame(noxflags.GameModeCoop) {
				monsterDieDropNative54A390(unit, drop)
			}
		},
	}
}

//export nox_monster_die_callback_native
func nox_monster_die_callback_native(kind C.int, unit *nox_object_t) C.int {
	return C.int(server.MonsterDieCallbackRaw54A7D0(asObjectS(unit), server.MonsterDieCallbackKind54A7D0(kind), monsterDieCallbackRawRuntime54A7D0()))
}
