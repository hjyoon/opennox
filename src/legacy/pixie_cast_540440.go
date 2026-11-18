package legacy

/*
#include "GAME4_3.h"
*/
import "C"

import (
	"github.com/opennox/libs/spell"
	"github.com/opennox/libs/things"
	"github.com/opennox/libs/types"
	"github.com/opennox/opennox/v1/common/memmap"
	"github.com/opennox/opennox/v1/server"
)

var pixieCastRuntime540440 = func(srv Server) server.PixieCastRuntime540440 {
	s := srv.S()
	return server.PixieCastRuntime540440{
		PixieType: s.Types.PixieID,
		CountOwned: func(owner *server.Object, typeID int32) int32 {
			return int32(owner.CountSubOfType(typeID))
		},
		Desired: func(level int) float64 {
			return s.Balance.FloatInd("PixieCount", level)
		},
		RandomInt: s.Rand.Logic.IntClamp,
		Direction: func(direction uint8) types.Pointf {
			offset := 8 * uintptr(direction)
			return types.Ptf(memmap.Float32(0x587000, 194136+offset), memmap.Float32(0x587000, 194140+offset))
		},
		TraceRay: func(from, to types.Pointf) bool {
			return s.MapTraceRay(from, to, server.MapTraceFlag1|server.MapTraceFlag3)
		},
		NewPixie: func() *server.Object {
			return s.NewObjectByTypeID("Pixie")
		},
		CreateAt: func(pixie, owner *server.Object, position types.Pointf) {
			srv.CreateObjectAt(pixie, owner, position)
		},
		FindTarget: func(pixie, owner *server.Object) *server.Object {
			return s.Nox_xxx_spellFlySearchTarget(nil, pixie, things.SpellOffensive, 600, 0, owner)
		},
		Frame: s.Frame, TickRate: s.TickRate,
		PlayCastAudio: func(id spell.ID, caster *server.Object) {
			s.Audio.EventObj(s.Spells.DefByInd(id).GetCastSound(), caster, 0, 0)
		},
	}
}

func CastPixies540440(srv Server, id spell.ID, owner, caster *server.Object, level int) int {
	if owner == nil || caster == nil {
		return 0
	}
	return server.CastPixiesNative540440(id, owner, caster, level, pixieCastRuntime540440(srv))
}

//export nox_xxx_castPixies_540440
func nox_xxx_castPixies_540440(id C.int, source, owner, caster *nox_object_t, args *C.nox_spell_accept_arg_t, level C.int) C.int {
	return C.int(CastPixies540440(GetServer(), spell.ID(id), asObjectS(owner), asObjectS(caster), int(level)))
}
