package legacy

import (
	"github.com/opennox/libs/spell"
	"github.com/opennox/libs/types"

	"github.com/opennox/opennox/v1/server"
)

func castCleansingFlameNative52D5C0(id spell.ID, second, recipient, caster *server.Object, arg *server.SpellAcceptArg, level int) int {
	outer := GetServer()
	return int(outer.S().CastCleansingFlame52D5C0(int32(id), second, recipient, caster, arg, int32(level), server.CleansingFlameCastRuntime52D5C0{
		CreateAt:       func(flame, owner *server.Object, position types.Pointf) { outer.CreateObjectAt(flame, owner, position) },
		DelayedDelete:  outer.DelayedDelete,
		UpdateCallback: Get_nox_xxx_updateFlameCleanse_53D510(),
	}))
}

var cleansingFlameCastCall52D5C0 = castCleansingFlameNative52D5C0
