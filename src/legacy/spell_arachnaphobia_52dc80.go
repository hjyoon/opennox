package legacy

import (
	"github.com/opennox/libs/spell"
	"github.com/opennox/libs/types"

	"github.com/opennox/opennox/v1/common/memmap"
	"github.com/opennox/opennox/v1/server"
)

func castArachnaphobiaNative52DC80(id spell.ID, second, owner, caster *server.Object, arg *server.SpellAcceptArg, level int) int {
	outer := GetServer()
	return int(outer.S().CastArachnaphobia52DC80(int32(id), second, owner, caster, arg, int32(level), server.ArachnaphobiaCastRuntime52DC80{
		TypeCache: memmap.PtrUint32(0x5D4594, 2487812),
		CreateAt: func(focus, owner *server.Object, position types.Pointf) {
			outer.CreateObjectAt(focus, owner, position)
		},
	}))
}

var arachnaphobiaCastCall52DC80 = castArachnaphobiaNative52DC80
