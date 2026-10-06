package legacy

import (
	"github.com/opennox/libs/spell"
	"github.com/opennox/libs/types"

	"github.com/opennox/opennox/v1/common/memmap"
	"github.com/opennox/opennox/v1/server"
)

func castToxicCloudNative52DB60(id spell.ID, second, owner, caster *server.Object, arg *server.SpellAcceptArg, level int) int {
	outer := GetServer()
	return int(outer.S().CastToxicCloud52DB60(int32(id), second, owner, caster, arg, int32(level), server.ToxicCloudCastRuntime52DB60{
		TypeCache: memmap.PtrUint32(0x5D4594, 2487808),
		CreateAt: func(cloud, owner *server.Object, position types.Pointf) {
			outer.CreateObjectAt(cloud, owner, position)
		},
	}))
}

// The public selector keeps native object/argument pointers instead of
// entering the retired five-int C callee.
var toxicCloudCastCall52DB60 = castToxicCloudNative52DB60
