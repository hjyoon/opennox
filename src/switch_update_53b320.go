package opennox

import (
	"github.com/opennox/opennox/v1/legacy"
	"github.com/opennox/opennox/v1/server"
)

func (s *Server) updateSwitchNative53B320(unit *server.Object) {
	s.SwitchUpdate53B320(unit, server.SwitchUpdateRuntime53B320{
		QueueCollision: legacy.Nox_xxx_unitHasCollideOrUpdateFn_537610,
	})
}
