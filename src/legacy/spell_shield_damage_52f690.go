package legacy

/*
#include "GAME4_2.h"
*/
import "C"

import (
	"unsafe"

	"github.com/opennox/libs/object"
	"github.com/opennox/libs/types"

	"github.com/opennox/opennox/v1/common/sound"
	"github.com/opennox/opennox/v1/server"
)

//export nox_xxx_netSendShieldFx_523670
func nox_xxx_netSendShieldFx_523670(target *nox_object_t, sourcePos *C.float2) C.int {
	GetServer().S().SendShieldFX523670(asObjectS(target), (*types.Pointf)(unsafe.Pointer(sourcePos)))
	return 0
}

//export nox_xxx_unitShield_52F690
func nox_xxx_unitShield_52F690(target *nox_object_t, amount C.int32_t) {
	spellShieldDamageNative52F690(GetServer().S(), asObjectS(target), int32(amount))
}

//export nox_xxx_unitShieldReduceDamage_52F710
func nox_xxx_unitShieldReduceDamage_52F710(target *nox_object_t, damage *C.int32_t, typ C.int, source *nox_object_t) {
	if target == nil || damage == nil {
		return
	}
	spellShieldReduceDamageNative52F710(GetServer().S(), asObjectS(target), (*int32)(damage),
		object.DamageType(typ), asObjectS(source))
}

func spellShieldDamageNative52F690(s *server.Server, target *server.Object, amount int32) {
	server.SpellShieldDamage52F690(
		s.Spells.Dur.SpellDurationFirst4FE930(), target, amount,
		server.SpellShieldDamageRuntime52F690{
			Cancel: func(record *server.DurSpell) {
				_ = s.Spells.Dur.SpellDurationCancel4FE9D0(record)
			},
			BuffOff: Nox_xxx_spellBuffOff_4FF5B0,
		},
	)
}

func spellShieldReduceDamageNative52F710(
	s *server.Server,
	target *server.Object,
	damage *int32,
	typ object.DamageType,
	source *server.Object,
) {
	server.SpellShieldReduceDamage52F710(
		target, damage, typ, source,
		server.SpellShieldReduceDamageRuntime52F710{
			Audio: func(id int, object *server.Object) {
				s.Audio.EventObj(sound.ID(id), object, 0, 0)
			},
			ShieldFX: func(target, source *server.Object) {
				var sourcePos *types.Pointf
				if source != nil {
					sourcePos = &source.PosVec
				}
				s.SendShieldFX523670(target, sourcePos)
			},
			CurrentHP: server.UnitGetHP4EE780,
			SetHP:     Nox_xxx_unitSetHP_4E4560,
			ShieldDamage: func(target *server.Object, amount int32) {
				spellShieldDamageNative52F690(s, target, amount)
			},
			FrontBlock: func(target, source *server.Object) bool {
				return Nox_server_testTwoPointsAndDirection_4E6E50(
					target.PosVec, int16(target.Direction1), source.PrevPos,
				)&1 != 0
			},
			Frame: s.Frame,
		},
	)
}
