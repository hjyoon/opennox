package legacy

import (
	"github.com/opennox/libs/object"
	"github.com/opennox/libs/spell"
	"github.com/opennox/libs/types"

	"github.com/opennox/opennox/v1/server"
)

func castMarkNative52CA80(id spell.ID, second, caster, aim *server.Object, arg *server.SpellAcceptArg, level int) int {
	outer := GetServer()
	return int(outer.S().CastMark52CA80(int32(id), second, caster, aim, arg, int32(level), server.MarkCastRuntime52CA80{
		GlyphTypeCache: Get_dword_5d4594_2487712_ptr(),
		Move: func(marker *server.Object, position *types.Pointf) {
			// Match 004E7010's original early return before reading pos.
			if marker == nil || marker.ObjClass&object.ClassImmobile != 0 {
				return
			}
			Nox_xxx_unitMove_4E7010(marker, *position)
		},
		CreateAt: func(marker, owner *server.Object, position types.Pointf) {
			outer.CreateObjectAt(marker, owner, position)
		},
	}))
}

var markCastCall52CA80 = castMarkNative52CA80
