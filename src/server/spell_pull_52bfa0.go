package server

import (
	"github.com/opennox/libs/spell"
	"github.com/opennox/libs/types"
)

type pullCastDeps52BFA0 struct {
	balance   func(string) float64
	pushUnits func(types.Pointf, float32, float32, float32, *Object, int, int)
	castAudio func(int32, *Object)
}

// pullCast52BFA0 restores GAME.EXE 0052BFA0..0052BFFD. Multiply the balance
// by the signed DWORD level, negate (FCHS), then spill to binary32, preserving
// negative zero. The origin is the caster's live native PosVec, not the
// acceptance target or the audio owner's position. No level/force clamp or
// silent nil-caster success existed in the original.
func pullCast52BFA0(id int32, owner, caster *Object, level int32, h pullCastDeps52BFA0) int32 {
	force := float32(-(h.balance("PullPowerCoeff") * float64(level)))
	h.pushUnits(caster.PosVec, 600, 10, force, nil, 0, 0)
	h.castAudio(id, owner)
	return 1
}

type PullCastRuntime52BFA0 struct {
	PushUnits func(types.Pointf, float32, float32, float32, *Object, int, int)
}

// CastPull52BFA0 preserves the spell-selector signature, including its unused
// second object and acceptance argument. The existing native radial service
// retains candidate filtering, missiles, ray tests and force application.
func (s *Server) CastPull52BFA0(id int32, _ *Object, owner, caster *Object, _ *SpellAcceptArg, level int32, runtime PullCastRuntime52BFA0) int32 {
	return pullCast52BFA0(id, owner, caster, level, pullCastDeps52BFA0{
		balance:   s.Balance.Float,
		pushUnits: runtime.PushUnits,
		castAudio: func(id int32, owner *Object) {
			s.Audio.EventObj(s.Spells.DefByInd(spell.ID(id)).GetCastSound(), owner, 0, 0)
		},
	})
}
