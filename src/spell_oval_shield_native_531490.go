package opennox

import "github.com/opennox/opennox/v1/server"

func (s *Server) SpellOvalShieldCreate531490(record *server.DurSpell) int32 {
	return s.Server.SpellOvalShieldCreate531490(record, ovalShieldRuntime531490())
}

func (s *Server) SpellOvalShieldUpdate5314F0(record *server.DurSpell) int32 {
	return s.Server.SpellOvalShieldUpdate5314F0(record)
}

func (s *Server) SpellOvalShieldDestroy531560(record *server.DurSpell) {
	s.Server.SpellOvalShieldDestroy531560(record, ovalShieldRuntime531490())
}
