package legacy

/*
#include <stdint.h>
#include "defs.h"
*/
import "C"

import (
	"unsafe"

	"github.com/opennox/libs/types"

	noxflags "github.com/opennox/opennox/v1/common/flags"
	"github.com/opennox/opennox/v1/common/memmap"
	"github.com/opennox/opennox/v1/common/sound"
	"github.com/opennox/opennox/v1/server"
)

func playerBotRuntime4FAB20() server.PlayerBotRuntime4FAB20 {
	if rt := playerBotRuntimeOverride4FAB20; rt != nil {
		return *rt
	}
	return server.PlayerBotRuntime4FAB20{
		Create: playerBotCreateRuntime4FA700(),
		Frame: func() uint32 {
			return GetServer().S().Frame()
		},
		LoadSettings: func() *server.Settings {
			return memmap.PtrT[server.Settings](playerRespawnSettingsBase4F7EF0, playerRespawnSettingsOffset4F7EF0)
		},
		NetworkMode: func() uint32 {
			return uint32(Get_dword_5d4594_2650652())
		},
		DefaultItems: func(unit *server.Object, first, second int32) {
			_ = playerMakeDefItemsCall4EF7D0(unit, first, second)
		},
		RespawnCorpse: respawnPlayerImplCall53FBC0,
		FindStart: func(out *types.Pointf, unit *server.Object) {
			GetServer().S().MapFindPlayerStartInto4F7AB0(out, unit)
		},
		Move: func(unit *server.Object, pos *types.Pointf) {
			Nox_xxx_unitMove_4E7010(unit, *pos)
		},
		Audio: func(id int32, unit *server.Object, first, second int32) {
			GetServer().S().Audio.EventObj(sound.ID(id), unit, int(first), uint32(second))
		},
		GameFlag: func(mask uint32) bool {
			return noxflags.HasGame(noxflags.GameFlag(mask))
		},
		Buff: func(unit *server.Object, id int32, duration int16, power int8) {
			Nox_xxx_buffApplyTo_4FF380(unit, server.EnchantID(id), int(duration), int(power))
		},
		UpdateMonster: func(unit *server.Object) {
			Nox_xxx_unitUpdateMonster_50A5C0(unit)
		},
		RegularPlayerUpdate: playerBotRegularUpdate4FAB20(),
	}
}

//export nox_xxx_updatePlayerMonsterBot_4FAB20
func nox_xxx_updatePlayerMonsterBot_4FAB20(unit *C.uint32_t) C.int {
	return C.int((*server.Object)(unsafe.Pointer(unit)).PlayerBotUpdate4FAB20(playerBotRuntime4FAB20()))
}

//export nox_xxx_respawnPlayerBot_4FAC70
func nox_xxx_respawnPlayerBot_4FAC70(unit *nox_object_t) C.int {
	return C.int(asObjectS(unit).PlayerBotRespawn4FAC70(playerBotRuntime4FAB20()))
}
