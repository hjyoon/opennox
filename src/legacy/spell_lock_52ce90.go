package legacy

import (
	"github.com/opennox/libs/spell"

	"github.com/opennox/opennox/v1/server"
)

// The original selected-door/nearest/group-owner globals are shared by
// nested casts. Keep that lifetime while restoring pointer-width storage.
var lockCastState52CE90 server.LockCastState52CE90

func castLockNative52CE90(id spell.ID, second, caster, aim *server.Object, arg *server.SpellAcceptArg, power int) int {
	return int(GetServer().S().CastLock52CE90(int32(id), second, caster, aim, arg, int32(power), server.LockCastRuntime52CE90{
		State: &lockCastState52CE90,
	}))
}

var lockCastCall52CE90 = castLockNative52CE90
