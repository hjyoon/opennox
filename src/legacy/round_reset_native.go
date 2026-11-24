package legacy

/*
#include "GAME3_2.h"
*/
import "C"

import (
	"encoding/binary"

	noxflags "github.com/opennox/opennox/v1/common/flags"
	"github.com/opennox/opennox/v1/common/memmap"
	"github.com/opennox/opennox/v1/legacy/common/alloc"
	"github.com/opennox/opennox/v1/server"
)

//export nox_xxx_mapFindCrown_4CFC30
func nox_xxx_mapFindCrown_4CFC30() {
	outer := GetServer()
	world := outer.S()
	crownType := memmap.PtrUint32(0x5d4594, 1523076)
	if *crownType == 0 {
		*crownType = uint32(world.Types.IndByID("Crown"))
	}
	for obj := world.Objs.First(); obj != nil; {
		next := obj.Next()
		if uint32(obj.TypeInd) == *crownType {
			outer.DelayedDelete(obj)
			Nox_xxx_respawnRemove_4EC6A0(obj)
		}
		obj = next
	}
}

//export sub_4CFDF0
func sub_4CFDF0(value C.int) C.int {
	*memmap.PtrUint32(0x5d4594, 1523072) = uint32(value)
	return value
}

//export sub_4CFE00
func sub_4CFE00() C.int { return C.int(memmap.Uint32(0x5d4594, 1523072)) }

//export sub_4D15C0
func sub_4D15C0() { *memmap.PtrUint32(0x5d4594, 1548508) = 0 }

//export nox_xxx_scavengerTreasureMax_4D1600
func nox_xxx_scavengerTreasureMax_4D1600() C.int {
	return C.int(memmap.Uint32(0x5d4594, 1548528))
}

//export sub_4D1610
func sub_4D1610() { *memmap.PtrUint32(0x5d4594, 1548528) = 0 }

//export nox_xxx_servResetPlayers_4D23C0
func nox_xxx_servResetPlayers_4D23C0() C.int {
	world := GetServer().S()
	for player := world.Players.First(); player != nil; player = world.Players.Next(player) {
		if player.PlayerUnit == nil {
			continue
		}
		setImportantRecipientMaskC(importantRecipientMaskC() &^ (1 << (uint8(player.PlayerInd) & 31)))
		player.Field3676 = 2
		playerMakeDefItemsCall4EF7D0(player.PlayerUnit, 1, 0)
		player.Field2140 = 0
		player.Lessons = 0
	}
	sub_51A100()
	noxflags.UnsetGame(noxflags.GameFlag18)
	nox_xxx_netGameSettings_4DEF00()
	headSettingsDirty409470(0)
	return 1
}

//export nox_xxx_netGameSettings_4DEF00
func nox_xxx_netGameSettings_4DEF00() C.int {
	settings := (*server.Settings2)(cliGamedataNative416590(0))
	frame := gameFrame()
	flags := uint32(noxflags.GetGame())
	subflags := uint32(headSettingsSubflags409470())
	limit := byte(headSettingsGetLimit409470())
	game := server.GameSettingsPacket(frame, uint32(C.NOX_CLIENT_VERS_CODE), flags, subflags, limit,
		byte(headSettingsGamedata409470(int16(settings.Field52))), sub40A180Native(int16(settings.Field52)))
	name := alloc.GoStringS(memmap.Slice(0x5d4594, 1324)[:16])
	details := server.GameSettingsDetailsPacket(name, settings)
	if sub40A220Native() != 0 && (sub_40A300() != 0 || sub40A180Native(int16(settings.Field52)) != 0) {
		binary.LittleEndian.PutUint32(details[45:49], uint32(sub40A230Native()))
	}
	const clients = C.int(0x80 | server.HostPlayerIndex)
	sendImportantPacketBytes(clients, game[:], false, true)
	return sendImportantPacketBytes(clients, details[:], false, true)
}
