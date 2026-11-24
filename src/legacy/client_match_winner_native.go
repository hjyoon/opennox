package legacy

/*
#include "GAME2_3.h"
*/
import "C"

import (
	noxflags "github.com/opennox/opennox/v1/common/flags"
	"github.com/opennox/opennox/v1/server"
)

func matchScoreLimit(host bool) int32 {
	if host {
		flags := int16(noxflags.GetGame())
		return int32(uint16(headSettingsGamedata409470(flags)))
	}
	settings := (*server.Settings2)(cliGamedataNative416590(0))
	return int32(settings.Field54)
}

func applyMatchScore(player *server.Player, winner, elimination bool, limit int32) {
	if player == nil || player.Field3680&1 != 0 {
		return
	}
	if elimination {
		if winner {
			if player.Field2140 >= uint32(limit) {
				player.Field2140 = uint32(limit - 1)
			}
		} else if player.Field2140 < uint32(limit) {
			player.Field2140 = uint32(limit)
		}
		return
	}
	if winner {
		player.Lessons = limit
	} else if player.Lessons >= limit {
		player.Lessons = limit - 1
	}
}

//export sub_4947E0
func sub_4947E0(winner *nox_playerInfo) {
	limit := matchScoreLimit(noxflags.HasGame(noxflags.GameHost))
	elimination := noxflags.HasGame(noxflags.GameModeElimination)
	players := &GetServer().S().Players
	playerWinner := asPlayerS(winner)
	for player := players.First(); player != nil; player = players.Next(player) {
		applyMatchScore(player, player == playerWinner, elimination, limit)
	}
}

//export sub_4948B0
func sub_4948B0(winner *nox_team_t) {
	if winner == nil {
		return
	}
	host := noxflags.HasGame(noxflags.GameHost)
	limit := matchScoreLimit(host)
	elimination := noxflags.HasGame(noxflags.GameModeElimination)
	world := GetServer().S()
	winningTeam := asTeam(winner)
	if !elimination {
		for team := world.Teams.First(); team != nil; team = world.Teams.Next(team) {
			if team == winningTeam {
				team.Lessons = limit
			} else if team.Lessons >= limit {
				team.Lessons = limit - 1
			}
		}
	}
	winnerID := winningTeam.ID()
	report := func(team *server.ObjectTeam, code uint32) {
		if !world.Teams.ContainsObject(team, winnerID) {
			player := asPlayerS(nox_common_playerInfoGetByID_417040(int32(code)))
			applyMatchScore(player, false, elimination, limit)
		}
	}
	if host {
		for unit := world.Players.FirstUnit(); unit != nil; unit = world.NextPlayerUnit4DA7F0(unit) {
			report(&unit.TeamVal, unit.NetCode)
		}
	} else {
		for dr := GetClient().Cli().Objs.FirstPlayerList(); dr != nil; dr = dr.Field104() {
			report(&dr.TeamVal, dr.NetCode32)
		}
	}
}
