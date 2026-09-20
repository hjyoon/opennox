package legacy

import (
	"github.com/opennox/libs/types"

	"github.com/opennox/opennox/v1/server"
)

type gameBallResetServer417F50 interface {
	GameBallReset417F50(*server.Object) int
}

func gameBallResetCall417F50(outer Server, old *server.Object) int {
	return outer.(gameBallResetServer417F50).GameBallReset417F50(old)
}

func gameBallUpdateRuntime53DF40(outer Server) server.GameBallUpdateRuntime53DF40 {
	return server.GameBallUpdateRuntime53DF40{
		Ticks: PlatformTicks,
		ResetBall: func(old *server.Object) int {
			return gameBallResetCall417F50(outer, old)
		},
		ChangeTeam: func(team *server.ObjectTeam, netCode uint32) {
			Nox_xxx_netChangeTeamMb_419570(team, netCode)
		},
		BallStatus: Sub_4E8290,
		Move: func(ball *server.Object, destination types.Pointf) {
			Nox_xxx_unitMove_4E7010(ball, destination)
		},
		ApplyForce: outer.ApplyForce,
	}
}

var gameBallUpdateCall53DF40 = func(ball *server.Object) {
	outer := GetServer()
	outer.S().GameBallUpdate53DF40(ball, gameBallUpdateRuntime53DF40(outer))
}
