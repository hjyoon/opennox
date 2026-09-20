package opennox

import (
	noxflags "github.com/opennox/opennox/v1/common/flags"
	"github.com/opennox/opennox/v1/legacy"
	"github.com/opennox/opennox/v1/server"
)

// mapInfoSetKotr4180D0 initializes KOTR with native-width Object and Team
// pointers instead of the PE32 integer walks in GAME.EXE 004180D0.
func (s *Server) mapInfoSetKotr4180D0() int {
	srv := s.S()
	ok := srv.MapInfoSetKotr4180D0(
		uint16(srv.Types.IndByID("Crown")),
		noxflags.HasGamePlay(noxflags.GameplayFlag4),
		server.KOTRMapSetupRuntime4180D0{
			DelayedDelete: s.DelayedDelete,
			RespawnRemove: legacy.Nox_xxx_respawnRemove_4EC6A0,
			MarkMinimapForAll: func(obj *server.Object) {
				for player := srv.Players.First(); player != nil; player = srv.Players.Next(player) {
					srv.Players.Nox_xxx_netMarkMinimapObject_417190(player.PlayerIndex(), obj, 1)
				}
			},
		},
	)
	if !ok {
		return 0
	}
	return 1
}
