package legacy

import (
	"github.com/opennox/libs/spell"

	"github.com/opennox/opennox/v1/server"
)

func castPoisonNative52C720(id spell.ID, second, owner, caster *server.Object, arg *server.SpellAcceptArg, level int) int {
	return int(GetServer().S().CastPoison52C720(int32(id), second, owner, caster, arg, int32(level), server.PoisonCastRuntime52C720{
		ActivatePoison:          Nox_xxx_activatePoison_4EE7E0,
		RecordPlayerAttribution: recordPlayerAttributionRuntime4E7540,
	}))
}

// Keep the selector's native target and owner pointers through both restored
// services; the PE32 C callee loaded arg.Obj into a pointer-truncating int.
var poisonCastCall52C720 = castPoisonNative52C720
