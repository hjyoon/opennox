package legacy

/*
#include "GAME4_3.h"
*/
import "C"

import (
	"github.com/opennox/libs/types"

	"github.com/opennox/opennox/v1/common/memmap"
	"github.com/opennox/opennox/v1/server"
)

const (
	playerRespawnSettingsBase4F7EF0        = uintptr(0x5d4594)
	playerRespawnSettingsOffset4F7EF0      = uintptr(371516)
	playerRespawnCorpseInitOffset53FBC0    = uintptr(2488736)
	playerRespawnCorpseTypesOffset53FBC0   = 2488740
	playerRespawnCorpsePointsBase53FBC0    = uintptr(0x587000)
	playerRespawnCorpsePointsOff53FBC0     = 280376
	playerRespawnCorpseDirectionSize53FBC0 = 88
	playerRespawnCorpseTypeSize53FBC0      = 44
)

func respawnPlayerTypeIndex53FBC0(direction int32, part int) uint32 {
	offset := playerRespawnCorpseTypesOffset53FBC0 +
		playerRespawnCorpseTypeSize53FBC0*int(direction) + 4*part
	return memmap.Uint32(playerRespawnSettingsBase4F7EF0, uintptr(offset))
}

func respawnPlayerOffset53FBC0(direction int32, part int) types.Pointf {
	offset := playerRespawnCorpsePointsOff53FBC0 +
		playerRespawnCorpseDirectionSize53FBC0*int(direction) + 8*part
	return types.Pointf{
		X: memmap.Float32(playerRespawnCorpsePointsBase53FBC0, uintptr(offset)),
		Y: memmap.Float32(playerRespawnCorpsePointsBase53FBC0, uintptr(offset+4)),
	}
}

func respawnPlayerImplRuntime53FBC0() server.RespawnPlayerImplRuntime53FBC0 {
	outer := GetServer()
	return server.RespawnPlayerImplRuntime53FBC0{
		Initialized: func() uint32 {
			return memmap.Uint32(playerRespawnSettingsBase4F7EF0, playerRespawnCorpseInitOffset53FBC0)
		},
		Initialize: func() {
			C.nox_xxx_createCorpse_53FCA0()
		},
		Direction: func(direction int16) int32 {
			return int32(Nox_xxx_math_509EA0(int(direction)))
		},
		TypeIndex: respawnPlayerTypeIndex53FBC0,
		Offset:    respawnPlayerOffset53FBC0,
		NetworkMode: func() uint32 {
			return uint32(Get_dword_5d4594_2650652())
		},
		CreateAt: func(object *server.Object, position types.Pointf) {
			outer.CreateObjectAt(object, nil, position)
		},
	}
}

func respawnPlayerImplCall53FBC0(center *types.Pointf, direction int16) {
	GetServer().S().RespawnPlayerImpl53FBC0(
		center,
		direction,
		respawnPlayerImplRuntime53FBC0(),
	)
}

func playerRespawnRuntime4F7EF0() server.PlayerRespawnRuntime4F7EF0 {
	outer := GetServer()
	s := outer.S()
	return server.PlayerRespawnRuntime4F7EF0{
		LoadSettings: func() *server.Settings {
			return memmap.PtrT[server.Settings](
				playerRespawnSettingsBase4F7EF0,
				playerRespawnSettingsOffset4F7EF0,
			)
		},
		MakeDefaultItems: func(unit *server.Object, restoreStats, keepItems int32) {
			_ = playerMakeDefItemsCall4EF7D0(unit, restoreStats, keepItems)
		},
		NetworkMode: func() uint32 {
			return uint32(Get_dword_5d4594_2650652())
		},
		RespawnCorpse:        respawnPlayerImplCall53FBC0,
		MapTileAllowTeleport: mapTileAllowTeleport411A90,
		Move:                 Nox_xxx_unitMove_4E7010,
		CrownPickup: func(who, crown *server.Object, first, second int32) {
			_ = crownPickupCall4F3400(s, who, crown, first, second)
		},
		ApplyBuff: func(unit *server.Object, enchant server.EnchantID, duration uint16, power uint8) {
			Nox_xxx_buffApplyTo_4FF380(unit, enchant, int(duration), int(power))
		},
	}
}

func playerRespawnCall4F7EF0(unit *server.Object) int16 {
	return GetServer().S().PlayerRespawn4F7EF0(unit, playerRespawnRuntime4F7EF0())
}
