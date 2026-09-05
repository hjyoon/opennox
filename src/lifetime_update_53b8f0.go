package opennox

import "github.com/opennox/opennox/v1/server"

func (s *Server) updateLifetimeNative53B8F0(source *server.Object) {
	s.LifetimeUpdate53B8F0(source, server.LifetimeUpdateRuntime53B8F0{
		DelayedDelete: s.DelayedDelete,
	})
}
