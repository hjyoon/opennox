package legacy

import (
	"github.com/opennox/libs/spell"
	"github.com/opennox/libs/types"

	"github.com/opennox/opennox/v1/server"
)

func castMeteorNative52D9D0(id spell.ID, second, owner, caster *server.Object, arg *server.SpellAcceptArg, level int) int {
	outer := GetServer()
	return int(outer.S().CastMeteor52D9D0(int32(id), second, owner, caster, arg, int32(level), server.MeteorCastRuntime52D9D0{
		TypeCache: Get_dword_5d4594_2487804_ptr(),
		CreateAt: func(meteor, owner *server.Object, position types.Pointf) {
			outer.CreateObjectAt(meteor, owner, position)
		},
	}))
}

// Preserve the public selector signature without entering the six-int C
// callee, which truncates the ownership root before reading Field129.
var meteorCastCall52D9D0 = castMeteorNative52D9D0
