package legacy

import (
	"github.com/opennox/libs/spell"
	"github.com/opennox/libs/types"

	"github.com/opennox/opennox/v1/common/memmap"
	"github.com/opennox/opennox/v1/server"
)

func castBurnNative52C3E0(id spell.ID, second, owner, caster *server.Object, arg *server.SpellAcceptArg, level int) int {
	outer := GetServer()
	return int(outer.S().CastBurn52C3E0(int32(id), second, owner, caster, arg, int32(level), server.BurnCastRuntime52C3E0{
		GlyphTypeCache: Get_dword_5d4594_2487712_ptr(),
		FlameTypeCache: memmap.PtrUint32(0x5D4594, 2487732),
		CreateAt: func(flame, owner *server.Object, position types.Pointf) {
			outer.CreateObjectAt(flame, owner, position)
		},
	}))
}

var burnCastCall52C3E0 = castBurnNative52C3E0
