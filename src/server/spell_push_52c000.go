package server

import (
	"github.com/opennox/libs/spell"
	"github.com/opennox/libs/types"
)

type pushCastDeps52C000 struct {
	balance   func(string) float64
	pushUnits func(types.Pointf, float32, float32, float32, *Object, int, int)
	castAudio func(int32, *Object)
}

// pushCast52C000 restores GAME.EXE 0052C000..0052C05D. The signed DWORD
// level multiplies the balance value before a binary32 spill. The origin is
// the caster's live PosVec, not the accepted target or the audio owner's
// position. Do not materialize caster +56: native PosVec is at +60 on wide
// hosts. The original neither clamps the level nor rejects an empty push.
func pushCast52C000(id int32, owner, caster *Object, level int32, h pushCastDeps52C000) int32 {
	force := float32(h.balance("PushPowerCoeff") * float64(level))
	h.pushUnits(caster.PosVec, 600, 10, force, nil, 0, 0)
	h.castAudio(id, owner)
	return 1
}

type PushCastRuntime52C000 struct {
	PushUnits func(types.Pointf, float32, float32, float32, *Object, int, int)
}

// CastPush52C000 retains the public spell-selector contract, including its
// unused second object and acceptance argument. The existing native radial
// push service owns candidate filtering, ray tests and force application.
func (s *Server) CastPush52C000(id int32, _ *Object, owner, caster *Object, _ *SpellAcceptArg, level int32, runtime PushCastRuntime52C000) int32 {
	return pushCast52C000(id, owner, caster, level, pushCastDeps52C000{
		balance:   s.Balance.Float,
		pushUnits: runtime.PushUnits,
		castAudio: func(id int32, owner *Object) {
			s.Audio.EventObj(s.Spells.DefByInd(spell.ID(id)).GetCastSound(), owner, 0, 0)
		},
	})
}
