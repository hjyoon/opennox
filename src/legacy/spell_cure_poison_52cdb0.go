package legacy

import (
	"github.com/opennox/libs/spell"

	"github.com/opennox/opennox/v1/server"
)

func castCurePoisonNative52CDB0(id spell.ID, second, owner, caster *server.Object, arg *server.SpellAcceptArg, power int) int {
	return int(GetServer().S().CastCurePoison52CDB0(int32(id), second, owner, caster, arg, int32(power), server.CurePoisonCastRuntime52CDB0{
		RefundMana: playerManaRechargeCall4FD030,
	}))
}

// Retain six arguments while narrowing only the numeric spell ID and power.
// The old C callee read the acceptance target into a pointer-truncating int.
var curePoisonCastCall52CDB0 = castCurePoisonNative52CDB0
