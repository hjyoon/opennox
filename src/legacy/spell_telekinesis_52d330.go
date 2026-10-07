package legacy

import (
	"github.com/opennox/libs/spell"
	"github.com/opennox/libs/types"

	"github.com/opennox/opennox/v1/server"
)

func castTelekinesisNative52D330(id spell.ID, second, owner, caster *server.Object, arg *server.SpellAcceptArg, power int) int {
	outer := GetServer()
	return int(outer.S().CastTelekinesis52D330(int32(id), second, owner, caster, arg, int32(power), server.TelekinesisCastRuntime52D330{
		CreateAt: func(hand, target *server.Object, point types.Pointf) {
			outer.CreateObjectAt(hand, target, point)
		},
		BuffApply: buffApplyExportCall4FF380,
	}))
}

// Preserve the public selector's native object and acceptance pointers rather
// than sending them through the original six-int C entry.
var telekinesisCastCall52D330 = castTelekinesisNative52D330
