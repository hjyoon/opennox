package legacy

import (
	"github.com/opennox/libs/spell"
	"github.com/opennox/libs/types"

	"github.com/opennox/opennox/v1/common/memmap"
	"github.com/opennox/opennox/v1/server"
)

func castMeteorShowerNative52D8A0(id spell.ID, second, owner, caster *server.Object, arg *server.SpellAcceptArg, level int) int {
	outer := GetServer()
	return int(outer.S().CastMeteorShower52D8A0(int32(id), second, owner, caster, arg, int32(level), server.MeteorShowerCastRuntime52D8A0{
		TypeCache: memmap.PtrUint32(0x5D4594, 2487800),
		CreateAt: func(shower, owner *server.Object, position types.Pointf) {
			outer.CreateObjectAt(shower, owner, position)
		},
	}))
}

var meteorShowerCastCall52D8A0 = castMeteorShowerNative52D8A0
