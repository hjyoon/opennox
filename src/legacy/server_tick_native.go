package legacy

/*
#include "server__system__server.h"
#include "common__system__team.h"
#include "cgo_const.h"
#include "noxstring.h"
*/
import "C"

import (
	"unsafe"

	"github.com/opennox/libs/types"

	noxflags "github.com/opennox/opennox/v1/common/flags"
	"github.com/opennox/opennox/v1/common/memmap"
	"github.com/opennox/opennox/v1/common/sound"
	"github.com/opennox/opennox/v1/legacy/common/alloc"
	"github.com/opennox/opennox/v1/server"
)

//export nox_xxx_getSomeMapName_4D0CF0
func nox_xxx_getSomeMapName_4D0CF0() *C.char {
	mode := int32(sub_4D0D50(C.int(noxflags.GetGame())))
	count := *mapCycleModeCount(mode)
	if count == 0 {
		return nil
	}
	index := memmap.PtrUint32(0x5D4594, mapCycleIndexOff+4*uintptr(mode))
	if int32(*index) > int32(count) {
		*index = 0
	}
	offset := mapCycleEntriesOff + mapCycleRecordSize*uintptr(*index+uint32(mode)*mapCycleMapLimit)
	*index = (*index + 1) % count
	return (*C.char)(memmap.PtrOff(0x5D4594, offset))
}

//export sub_4D10F0
func sub_4D10F0(name *C.char) {
	if name == nil {
		return
	}
	count, clock := questMapFileGlobals4D0F60()
	for i := int32(0); i < int32(*count); i++ {
		entry := questMapEntryNative4D0F60(i)
		entryName := (*C.char)(unsafe.Pointer(&entry.Name[0]))
		if C.nox_strcmpi(entryName, name) != 0 {
			continue
		}
		stamp := *clock
		*memmap.PtrUint32(0x587000, 191880) = uint32(i)
		entry.Uses++
		entry.LastUsed = stamp
		for j := int32(0); j < int32(*count); j++ {
			related := questMapEntryNative4D0F60(j)
			if j != i && related.Group == entry.Group {
				related.Uses++
				related.LastUsed = stamp
			}
		}
		*clock = stamp + 1
		return
	}
}

//export sub_4D2160
func sub_4D2160() {
	playerCount := nox_common_playerInfoCount_416F40()
	for team := nox_server_teamFirst_418B10(); team != nil; team = nox_server_teamNext_418B60(team) {
		if asObjectS(nox_server_teamGetFlagObject_4180D0(team)).InvHolder != nil || sub_418BC0(team) == 0 {
			continue
		}
		for {
			index := nox_common_randomInt_415FA0(0, playerCount-1)
			player := nox_common_playerInfoGetFirst_416EA0()
			for player != nil && index != 0 {
				index--
				player = nox_common_playerInfoGetNext_416EE0(player)
			}
			unit := asObjectS(player.playerUnit)
			if unit == nil || nox_xxx_teamCompare2_419180((*C.nox_object_team_t)(unit.TeamVal.C()), team.field_57) == 0 {
				continue
			}
			crown := nox_server_teamGetFlagObject_4180D0(team)
			sub_4F3400(player.playerUnit, crown, 1, 1)
			break
		}
	}
}

//export sub_4D22B0
func sub_4D22B0() {
	for player := nox_common_playerInfoGetFirst_416EA0(); player != nil; player = nox_common_playerInfoGetNext_416EE0(player) {
		unit := player.playerUnit
		if unit == nil {
			continue
		}
		update := (*C.nox_player_update_data_t)(unit.data_update)
		if update.trade_70 != nil {
			nox_xxx_shopCancelSession_510DC0(update.trade_70)
		}
		update.trade_70 = nil
		nox_xxx_playerResetControlBuffer_51AC30(int32(player.playerInd))
		if noxflags.HasGame(noxflags.GameModeQuest) {
			value := player.field_4676
			player.field_4676 = 0
			player.field_4680 = value
		}
		if nox_xxx_playerSetState_4FA020(player.playerUnit, int32(server.PlayerState13)) == 0 ||
			!noxflags.HasGame(noxflags.GameModeCoopTeam) {
			player.field_4700 = 0
			keepItems := (uint32(sub_4CFE00()) >> 1) & 1
			restoreStats := nox_xxx_gameIsSwitchToSolo_4DB240() == 0
			nox_xxx_playerMakeDefItems_4EF7D0(player.playerUnit, C.int32_t(bool2int(restoreStats)), C.int32_t(keepItems))
		}
		if uint32(player.field_3680)&0x20 != 0 {
			nox_xxx_playerLeaveObserver_0_4E6AA0(player)
			nox_xxx_playerCameraUnlock_4E6040(player.playerUnit)
		}
		unit = player.playerUnit
		unit.field_34 = C.uint32_t(gameFrame())
	}
}

//export sub_4E4170
func sub_4E4170() {
	if gameFrame()%(5*gameFPS()) != 0 {
		return
	}
	before := float32(sub_4E3CA0())
	sub_4E3D50()
	if float64(sub_4E3CA0()) != float64(before) {
		sub_4E3DD0()
	}
}

//export sub_500510
func sub_500510(name *C.nox_cgo_const_char) {
	if name == nil {
		return
	}
	value := GoString((*C.char)(name))
	slot := memmap.Slice(questJournalMemoryBase5009B0, questJournalMapNameOffset5009B0)
	slot = slot[:questJournalScratchOffset5009B0-questJournalMapNameOffset5009B0]
	alloc.StrCopy(slot, value)
}

//export sub_51A920
func sub_51A920(value C.int) C.int {
	*memmap.PtrUint32(0x5D4594, 2388660) = uint32(value)
	return value
}

func playSecretWallSound(wall *server.SecretWall, name *C.char) {
	pos := types.Pointf{X: float32(23*wall.X + 11), Y: float32(23*wall.Y + 11)}
	id := sound.ID(nox_xxx_utilFindSound_40AF50(name))
	GetServer().S().Audio.EventPos(id, pos, 0, 0)
}

//export nox_xxx_updateUnits_51B100_D
func nox_xxx_updateUnits_51B100_D() {
	moving := false
	for wall := secretWallFirstNative410780(); wall != nil; wall = secretWallNextNative410790(wall) {
		switch wall.State {
		case 1:
			moving = false
			if wall.Flags&0xc == 0xc {
				wall.LastOpen--
				if wall.LastOpen == 0 {
					wall.State = 4
					playSecretWallSound(wall, nox_xxx_wallFindOpenSound_410EE0(int32(wall.Wall.Tile1)))
				}
			}
		case 2:
			wall.OpenDelay--
			if wall.OpenDelay == 0 {
				wall.State = 1
				wall.LastOpen = gameFPS() * wall.OpenWait
			}
			moving = true
		case 3:
			moving = false
			if wall.Flags&0xc == 4 {
				wall.LastOpen--
				if wall.LastOpen == 0 {
					wall.State = 2
					playSecretWallSound(wall, nox_xxx_wallFindCloseSound_410F20(int32(wall.Wall.Tile1)))
				}
			}
		case 4:
			wall.OpenDelay++
			if wall.OpenDelay == 23 {
				wall.State = 3
				wall.LastOpen = gameFPS() * wall.OpenWait
			}
			moving = true
		}
		if moving {
			x, y := float64(23*wall.X)+11.5, float64(23*wall.Y)+11.5
			rect := types.Rectf{
				Min: types.Pointf{X: float32(x - 42.5), Y: float32(y - 42.5)},
				Max: types.Pointf{X: float32(x + 42.5), Y: float32(y + 42.5)},
			}
			GetServer().S().Map.EachObjInRect(rect, func(obj *server.Object) bool {
				sub_51B860(asObjectC(obj))
				return true
			})
		}
	}
}
