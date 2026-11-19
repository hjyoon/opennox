package legacy

/*
#include "GAME3_2.h"
*/
import "C"

import (
	"github.com/opennox/opennox/v1/common/memmap"
	"github.com/opennox/opennox/v1/server"
)

const gameBallPlayerDamageTypeCacheOffset4E1230 = uintptr(1563316)

func gameBallPlayerDamageRuntime4E1230(s *server.Server) server.GameBallPlayerDamageRuntime4E1230 {
	teamServices := ballCollideRuntime4EBA00(s)
	return server.GameBallPlayerDamageRuntime4E1230{
		LoadTypeCache: func() uint32 {
			return *memmap.PtrUint32(0x5D4594, gameBallPlayerDamageTypeCacheOffset4E1230)
		},
		StoreTypeCache: func(value uint32) {
			*memmap.PtrUint32(0x5D4594, gameBallPlayerDamageTypeCacheOffset4E1230) = value
		},
		ApplyForce: GetServer().ApplyForce,
		ChangeTeam: teamServices.ChangeTeam,
		CreateTeam: teamServices.CreateTeam,
	}
}

//export nox_xxx_gameballOnPlayerDamage_4E1230
func nox_xxx_gameballOnPlayerDamage_4E1230(source, target *nox_object_t, damage C.int) {
	s := GetServer().S()
	s.GameBallOnPlayerDamage4E1230(asObjectS(source), asObjectS(target), int32(damage), gameBallPlayerDamageRuntime4E1230(s))
}
