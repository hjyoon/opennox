package opennox

import (
	"github.com/opennox/libs/spell"
	"github.com/opennox/opennox/v1/legacy"
	"github.com/opennox/opennox/v1/server"
)

func (s *Server) castPixies540440(
	spellID spell.ID,
	_ *server.Object,
	owner, caster *server.Object,
	_ *server.SpellAcceptArg,
	level int,
) int {
	return legacy.CastPixies540440(s, spellID, owner, caster, level)
}
