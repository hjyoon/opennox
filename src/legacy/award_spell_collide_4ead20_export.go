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

//export nox_xxx_collideSpellPedestal_4EAD20
func nox_xxx_collideSpellPedestal_4EAD20(
	source, target *C.nox_object_t,
	collision *C.float,
) C.int {
	srv := GetServer()
	result := srv.S().AwardSpellCollide4EAD20(
		asObjectS((*nox_object_t)(source)),
		asObjectS((*nox_object_t)(target)),
		(*types.Pointf)(unsafe.Pointer(collision)),
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
	return C.int(result)
}
