package legacy

import (
	"github.com/opennox/libs/spell"

	"github.com/opennox/opennox/v1/server"
)

func castInversionNative52BEB0(id spell.ID, second, owner, caster *server.Object, arg *server.SpellAcceptArg, level int) int {
	return int(GetServer().S().CastInversion52BEB0(int32(id), second, owner, caster, arg, int32(level), server.InversionCastRuntime52BEB0{
		ChangeOwner: Nox_xxx_changeOwner_52BE40,
	}))
}

// The old four-int C callee narrowed both the query center and callback
// owner, despite receiving full-width pointers from the six-argument bridge.
var inversionCastCall52BEB0 = castInversionNative52BEB0
