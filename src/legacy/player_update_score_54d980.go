package legacy

/*
#include "player_update_score_54d980.h"
*/
import "C"

import "github.com/opennox/opennox/v1/server"

func playerUpdateScoreRuntime54D980(srv Server) server.PlayerUpdateScoreRuntime54D980 {
	s := srv.S()
	return server.PlayerUpdateScoreRuntime54D980{
		HasTeam:  func(team *server.ObjectTeam) bool { return team.Has() },
		TeamByID: func(id uint8) *server.Team { return s.Teams.ByID(server.TeamID(id)) },
		AddScore: func(unit *server.Object, value uint32) {
			server.PlayerScoreAdd4D8E90(unit, value)
		},
		SubtractScore: func(unit *server.Object, value uint32) {
			server.PlayerScoreSubtract4D8EC0(unit, value)
		},
		ReportLesson:       s.Nox_xxx_netReportLesson_4D8EF0,
		IncrementElimDeath: srv.PlayerIncrementElimDeath4D8D40,
		TeamChangeLessons: func(team *server.Team, value int32) {
			s.TeamChangeLessons(team, int(value))
		},
		ObserverMode:   func() uint32 { return uint32(Get_dword_5d4594_2650652()) },
		ObserverUpdate: flagPickupObserverUpdate425CA0,
	}
}

func Nox_xxx_playerUpdateScoreNative54D980(victim, killer, assist *server.Object, tracking uint32) {
	C.nox_player_update_score_native_call_54D980(asObjectC(victim), asObjectC(killer), asObjectC(assist), C.uint32_t(tracking))
}

//export nox_server_player_update_score_native_54D980
func nox_server_player_update_score_native_54D980(victim, killer, assist *nox_object_t, tracking C.uint32_t) {
	srv := GetServer()
	server.PlayerUpdateScore54D980(asObjectS(victim), asObjectS(killer), asObjectS(assist), uint32(tracking), playerUpdateScoreRuntime54D980(srv))
}
