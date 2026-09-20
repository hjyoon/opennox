package opennox

import (
	"github.com/opennox/libs/object"

	noxflags "github.com/opennox/opennox/v1/common/flags"
	"github.com/opennox/opennox/v1/legacy"
	"github.com/opennox/opennox/v1/server"
)

// mapInfoSetFlagball417F30 initializes FlagBall mode without the PE32 object
// and player iterators used by GAME.EXE 00417EC0/004181F0.
func (s *Server) mapInfoSetFlagball417F30() int {
	srv := s.S()
	flagCount := 0
	for obj := srv.Objs.First(); obj != nil; obj = obj.Next() {
		if obj.ObjClass.Has(object.ClassFlag) {
			flagCount++
		}
	}
	legacy.SetFlagObjectCount417EC0(uint32(flagCount))
	if flagCount == 0 {
		return 0
	}

	if !noxflags.HasGame(noxflags.GameFlag16) {
		srv.TeamAutoAssign4181F0(server.TeamAutoAssignRuntime4181F0{
			ClientNetCode: func() uint32 {
				return uint32(legacy.ClientPlayerNetCode())
			},
			NoRendering: func() bool {
				return noxflags.HasEngine(noxflags.EngineNoRendering)
			},
			PreferConfiguredTeam: func() bool {
				return legacy.Sub_40A740() != 0
			},
			Attach: func(teamID server.TeamID, value *server.ObjectTeam, active int32, netCode uint32, flags int32) {
				legacy.Nox_xxx_createAtImpl_4191D0(teamID, value, int(active), int(netCode), int(flags))
			},
		})
	}
	legacy.Sub_455F60()
	return s.GameBallReset417F50(nil)
}
