package legacy

import (
	"github.com/opennox/libs/spell"

	"github.com/opennox/opennox/v1/server"
)

func castFumbleNative52C060(id spell.ID, second, owner, caster *server.Object, arg *server.SpellAcceptArg, level int) int {
	s := GetServer()
	return int(s.S().CastFumble52C060(int32(id), second, owner, caster, arg, int32(level), server.FumbleCastRuntime52C060{
		ForceDrop:  objectForceDropCall4ED930,
		DropAll:    Nox_xxx_dropAllItems_4EDA40,
		ApplyForce: s.ApplyForce,
	}))
}

// The PE32 callee narrowed the target, inventory and owned-list pointers to
// DWORDs and used caster +56. Keep the real selector entry native-width.
var fumbleCastCall52C060 = castFumbleNative52C060
