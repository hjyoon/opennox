package legacy

/*
#include "GAME4_3.h"
*/
import "C"

import "github.com/opennox/opennox/v1/server"

func monsterFightActionRuntime531EC0() server.MonsterActionFightRuntime531EC0 {
	return server.MonsterActionFightRuntime531EC0{
		Distance:  objectDistance_4E6C00,
		CanSummon: monsterFightSchedulerRuntime531B40().CanSummon,
	}
}

//export nox_xxx_mobActionFight_531EC0
func nox_xxx_mobActionFight_531EC0(unit *nox_object_t) C.char {
	return C.char(GetServer().S().MonsterActionFightResult531EC0(asObjectS(unit), monsterFightActionRuntime531EC0()))
}
