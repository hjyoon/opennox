package opennox

import (
	"github.com/opennox/libs/types"

	"github.com/opennox/opennox/v1/legacy"
	"github.com/opennox/opennox/v1/server"
)

func (s *Server) elevatorRuntime53B5D0() server.ElevatorUpdateRuntime53B5D0 {
	return server.ElevatorUpdateRuntime53B5D0{
		Move: func(unit *server.Object, position types.Pointf) {
			asObjectS(unit).SetPos(position)
		},
		QueueCollision: legacy.Nox_xxx_unitHasCollideOrUpdateFn_537610,
	}
}

func (s *Server) updateElevatorNative53B5D0(unit *server.Object) {
	s.ElevatorUpdate53B5D0(unit, s.elevatorRuntime53B5D0())
}

func (s *Server) updateElevatorShaftNative53B380(unit *server.Object) {
	s.ElevatorShaftUpdate53B380(unit, s.elevatorRuntime53B5D0())
}
