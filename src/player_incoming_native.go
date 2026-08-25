package opennox

import (
	noxflags "github.com/opennox/opennox/v1/common/flags"
	"github.com/opennox/opennox/v1/common/ntype"
	"github.com/opennox/opennox/v1/legacy"
	"github.com/opennox/opennox/v1/server"
)

// netPlayerIncomingServNative4DDF60 preserves the named Go player fields that
// GAME.EXE mutates in nox_xxx_netPlayerIncomingServ_4DDF60. The legacy routine
// still performs the remaining protocol work, but its Win32 player offsets do
// not address these fields in the native 64-bit layout.
func (s *Server) netPlayerIncomingServNative4DDF60(ind ntype.PlayerInd) {
	legacy.Nox_xxx_netPlayerIncomingServ_4DDF60(int(ind))
	finishPlayerIncomingNative4DDF60(
		s.Players.ByInd(ind),
		noxflags.HasGame(noxflags.GameModeCoopTeam),
	)
}

func finishPlayerIncomingNative4DDF60(pl *server.Player, coopTeam bool) {
	if pl == nil {
		return
	}
	pl.Field4700 = 0
	pl.Field3676 = 3
	if !coopTeam && pl.PlayerUnit != nil {
		pl.Pos3632Vec = pl.PlayerUnit.PosVec
	}
}
