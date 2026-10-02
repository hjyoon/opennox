package legacy

import (
	"github.com/opennox/libs/spell"

	"github.com/opennox/opennox/v1/server"
)

func castPullNative52BFA0(id spell.ID, second, owner, caster *server.Object, arg *server.SpellAcceptArg, level int) int {
	return int(GetServer().S().CastPull52BFA0(int32(id), second, owner, caster, arg, int32(level), server.PullCastRuntime52BFA0{
		PushUnits: Nox_xxx_mapPushUnitsAround_52E040,
	}))
}

// The six-int C callee narrowed caster/audio-owner pointers and assumed the
// PE32 caster +56 position. Keep native-width objects through the real entry.
var pullCastCall52BFA0 = castPullNative52BFA0
