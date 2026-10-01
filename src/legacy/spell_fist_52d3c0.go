package legacy

import (
	"github.com/opennox/libs/spell"
	"github.com/opennox/libs/types"

	"github.com/opennox/opennox/v1/server"
)

func castFistNative52D3C0(id spell.ID, second, owner, caster *server.Object, arg *server.SpellAcceptArg, level int) int {
	outer := GetServer()
	return int(outer.S().CastFist52D3C0(int32(id), second, owner, caster, arg, int32(level), server.FistCastRuntime52D3C0{
		CreateAt: func(fist, caster *server.Object, position types.Pointf) {
			outer.CreateObjectAt(fist, caster, position)
		},
	}))
}

// Keep the existing Go selector signature without invoking the old six-int C
// function, which truncated the ownership root before reading Field129.
var fistCastCall52D3C0 = castFistNative52D3C0
