package legacy

import (
	"github.com/opennox/libs/spell"
	"github.com/opennox/libs/types"

	"github.com/opennox/opennox/v1/server"
)

func castPushNative52C000(id spell.ID, second, owner, caster *server.Object, arg *server.SpellAcceptArg, level int) int {
	return int(GetServer().S().CastPush52C000(int32(id), second, owner, caster, arg, int32(level), server.PushCastRuntime52C000{
		PushUnits: func(_ types.Pointf, outer, inner, force float32, source *server.Object, callback, arg int) {
			Nox_xxx_mapPushUnitsAroundPtr52E040(&caster.PosVec, outer, inner, force, source, callback, arg)
		},
	}))
}

// Avoid the six-int C callee: it narrowed both caster and audio owner and
// assumed the PE32 position offset, even when the bridge passed full pointers.
var pushCastCall52C000 = castPushNative52C000
