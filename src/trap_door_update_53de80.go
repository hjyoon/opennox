package opennox

import (
	"github.com/opennox/opennox/v1/common/sound"
	"github.com/opennox/opennox/v1/server"
)

func (s *Server) updateTrapDoorNative53DE80(unit *server.Object) {
	s.TrapDoorUpdate53DE80(unit, server.TrapDoorUpdateRuntime53DE80{
		AudioEvent: func(id uint32, obj *server.Object) {
			s.Audio.EventObj(sound.ID(id), obj, 0, 0)
		},
	})
}
