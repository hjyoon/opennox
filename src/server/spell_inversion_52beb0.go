package server

import (
	"github.com/opennox/libs/spell"
	"github.com/opennox/libs/types"
)

type inversionCastDeps52BEB0 struct {
	balance     func(string) float64
	eachMissile func(types.Pointf, float32, func(*Object) bool)
	changeOwner func(*Object, *Object)
	castAudio   func(int32, *Object)
}

// inversionCast52BEB0 restores GAME.EXE 0052BEB0..0052BEFB. The complete
// InversionRange is spilled to binary32 before the caster's live position is
// loaded. Every missile is passed to the existing 0052BE40 ownership service,
// which owns its magic/SWAP/target gates. Cast audio belongs to the caster,
// not to the potentially different callback owner. No level scaling or
// silent missing-caster success existed in the original.
func inversionCast52BEB0(id int32, owner, caster *Object, h inversionCastDeps52BEB0) int32 {
	radius := float32(h.balance("InversionRange"))
	h.eachMissile(caster.PosVec, radius, func(missile *Object) bool {
		h.changeOwner(missile, owner)
		return true
	})
	h.castAudio(id, caster)
	return 1
}

type InversionCastRuntime52BEB0 struct {
	ChangeOwner func(*Object, *Object)
}

func (s *Server) CastInversion52BEB0(id int32, _ *Object, owner, caster *Object, _ *SpellAcceptArg, _ int32, runtime InversionCastRuntime52BEB0) int32 {
	return inversionCast52BEB0(id, owner, caster, inversionCastDeps52BEB0{
		balance:     s.Balance.Float,
		eachMissile: s.Map.EachMissileInCircle,
		changeOwner: runtime.ChangeOwner,
		castAudio: func(id int32, caster *Object) {
			s.Audio.EventObj(s.Spells.DefByInd(spell.ID(id)).GetCastSound(), caster, 0, 0)
		},
	})
}
