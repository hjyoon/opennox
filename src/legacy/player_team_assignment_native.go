package legacy

/*
#include "GAME3_2.h"
*/
import "C"

import (
	"unsafe"

	noxflags "github.com/opennox/opennox/v1/common/flags"
	"github.com/opennox/opennox/v1/server"
)

func Sub_4DF3C0(player *server.Player) {
	if player == nil || player.PlayerUnit == nil {
		return
	}
	unit := player.PlayerUnit
	world := GetServer().S()
	settings := sub416640Native()
	if sub40A740Native() == 0 && !noxflags.HasGame(noxflags.GameFlag16) {
		if world.Teams.Count() != 0 {
			if team := sub_4189D0(); team != nil && !unit.TeamVal.Has() {
				Nox_xxx_createAtImpl_4191D0(server.TeamID(team.field_57), &unit.TeamVal, 1, int(unit.NetCode), 1)
			}
		}
		return
	}
	if player.Field2068 == 0 {
		return
	}

	team := world.Teams.ByXxx(int(int32(player.Field2068)))
	if team == nil {
		maximum := *(*byte)(unsafe.Add(settings, 52))
		objective := noxflags.GameModeCTF | noxflags.GameModeFlagBall
		if noxflags.HasGame(objective) ||
			noxflags.HasGame(noxflags.GameModeKOTR) && noxflags.HasGamePlay(noxflags.GameplayFlag4) {
			maximum = min(maximum, 2)
		}
		if sub417DE0Native() >= maximum ||
			noxflags.HasGame(objective) && int32(sub417DE0Native()) >= sub417DC0Native() {
			return
		}
		record := sub_418A10()
		if record == nil {
			return
		}
		sub_418800(&record.name[0], (*wchar2_t)(unsafe.Pointer(&player.Field2072[0])), 0)
		sub_418830(record, C.int(player.Field2068))
		team = asTeam(record)
		sub4184D0Native(team)
	}
	if unit.TeamVal.Has() {
		Sub_4196D0(unit.TeamVal.C(), team.C(), int(unit.NetCode), 0)
	} else {
		Nox_xxx_createAtImpl_4191D0(team.ID(), &unit.TeamVal, 1, int(unit.NetCode), 0)
	}
}

//export sub_4DF3C0
func sub_4DF3C0(player *nox_playerInfo) {
	Sub_4DF3C0(asPlayerS(player))
}
