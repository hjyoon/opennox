package opennox

import (
	"github.com/opennox/libs/object"

	noxflags "github.com/opennox/opennox/v1/common/flags"
	"github.com/opennox/opennox/v1/legacy"
	"github.com/opennox/opennox/v1/server"
)

func (s *Server) teamAutoAssignRuntime4181F0() server.TeamAutoAssignRuntime4181F0 {
	return server.TeamAutoAssignRuntime4181F0{
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
	}
}

// mapInfoSetFlags417EC0 counts flag objects and assigns teamless players
// without the PE32 object and player iterators used by GAME.EXE 00417EC0 and
// 004181F0.
func (s *Server) mapInfoSetFlags417EC0() bool {
	srv := s.S()
	flagCount := 0
	for obj := srv.Objs.First(); obj != nil; obj = obj.Next() {
		if obj.ObjClass.Has(object.ClassFlag) {
			flagCount++
		}
	}
	legacy.SetFlagObjectCount417EC0(uint32(flagCount))
	if flagCount == 0 {
		return false
	}

	if !noxflags.HasGame(noxflags.GameFlag16) {
		srv.TeamAutoAssign4181F0(s.teamAutoAssignRuntime4181F0())
	}
	return true
}

// mapInfoSetCapflag417EA0 initializes CTF without routing native object and
// player pointers through the original PE32 setup routine.
func (s *Server) mapInfoSetCapflag417EA0() int {
	if !s.mapInfoSetFlags417EC0() {
		return 0
	}
	legacy.Sub_455A50(2)
	return 1
}

// mapInfoSetFlagball417F30 initializes FlagBall mode on top of the shared
// native-width flag setup.
func (s *Server) mapInfoSetFlagball417F30() int {
	if !s.mapInfoSetFlags417EC0() {
		return 0
	}
	legacy.Sub_455F60()
	return s.GameBallReset417F50(nil)
}
