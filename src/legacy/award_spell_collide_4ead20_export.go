package legacy

/*
#include "award_spell_collide_4ead20.h"
*/
import "C"

import (
	"unsafe"

	"github.com/opennox/libs/spell"
	"github.com/opennox/libs/types"

	"github.com/opennox/opennox/v1/server"
)

var awardSpellCollideCall4EAD20 = func(source, target *server.Object, collision unsafe.Pointer) int32 {
	return GetServer().S().AwardSpellCollide4EAD20(
		source,
		target,
		(*types.Pointf)(collision),
		server.AwardSpellCollideRuntime4EAD20{
			GrantSpell: func(obj *server.Object, spellID uint32, mode, fourth, fifth int32) int32 {
				return int32(Nox_xxx_spellGrantToPlayer_4FB550(
					obj,
					spell.ID(spellID),
					int(mode),
					int(fourth),
					int(fifth),
				))
			},
		},
	)
}

//export nox_xxx_collideSpellPedestal_4EAD20
func nox_xxx_collideSpellPedestal_4EAD20(
	source, target *C.nox_object_t,
	collision *C.float,
) C.int {
	return C.int(awardSpellCollideCall4EAD20(
		asObjectS((*nox_object_t)(source)),
		asObjectS((*nox_object_t)(target)),
		unsafe.Pointer(collision),
	))
}
