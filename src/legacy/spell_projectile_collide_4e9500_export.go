package legacy

/*
#include "spell_projectile_collide_4e9500.h"
*/
import "C"

import (
	"unsafe"

	"github.com/opennox/libs/spell"
	"github.com/opennox/libs/types"

	"github.com/opennox/opennox/v1/server"
)

//export nox_xxx_spellFlyCollide_4E9500
func nox_xxx_spellFlyCollide_4E9500(
	projectile, other *C.nox_object_t,
	collision *C.float,
) {
	srv := GetServer()
	srv.S().SpellProjectileCollide4E9500(
		asObjectS((*nox_object_t)(projectile)),
		asObjectS((*nox_object_t)(other)),
		(*types.Pointf)(unsafe.Pointer(collision)),
		server.SpellProjectileCollideRuntime4E9500{
			CheckDirection: func(first types.Pointf, direction int16, second types.Pointf) int32 {
				return twoPointsAndDirection4E6E50(first, int32(direction), second)
			},
			ChangeOwner:    Nox_xxx_changeOwner_52BE40,
			SetPlayerState: Nox_xxx_playerSetState_4FA020,
			SpellAccept: func(
				spellID spell.ID,
				source, owner, projectile *server.Object,
				arg *server.SpellAcceptArg,
				level int,
			) bool {
				return srv.Nox_xxx_spellAccept4FD400(spellID, source, owner, projectile, arg, level)
			},
			DelayedDelete:   srv.DelayedDelete,
			InversionEffect: InversionEffectPointer4E03D0(),
		},
	)
}
