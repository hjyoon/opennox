package opennox

import (
	"github.com/opennox/libs/object"

	noxflags "github.com/opennox/opennox/v1/common/flags"
	"github.com/opennox/opennox/v1/legacy"
)

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
		legacy.TeamAutoAssign4181F0(false)
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
