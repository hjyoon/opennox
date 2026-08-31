package legacy

/*
#include "GAME3_3.h"
*/
import "C"

import (
	"unsafe"

	"github.com/opennox/libs/object"

	"github.com/opennox/opennox/v1/server"
)

//export nox_xxx_collidePlayer_4E8460
func nox_xxx_collidePlayer_4E8460(player, other *C.nox_object_t, collision *C.float) {
	srv := GetServer()
	srv.S().PlayerCollide4E8460(
		asObjectS((*nox_object_t)(player)),
		asObjectS((*nox_object_t)(other)),
		unsafe.Pointer(collision),
		server.PlayerCollideRuntime4E8460{
			SetState: func(obj *server.Object, state server.PlayerState) {
				Nox_xxx_playerSetState_4FA020(obj, state)
			},
			DisableAbility: func(obj *server.Object, ability server.Ability) {
				Sub_4FC300(obj, int32(ability))
			},
			ApplyEnchant: func(obj *server.Object, enchant server.EnchantID, duration, power uint32) {
				Nox_xxx_buffApplyTo_4FF380(obj, enchant, int(duration), int(uint8(power)))
			},
			DamageMap: func(x, y, damage int32, damageType object.DamageType, source *server.Object) {
				srv.Nox_xxx_damageToMap_534BC0(int(x), int(y), int(damage), damageType, source)
			},
			DamageClear: func(obj *server.Object, damage int32) {
				Nox_xxx_unitDamageClear_4EE5E0(obj, int(damage))
			},
			Move: Nox_xxx_unitMove_4E7010,
			DisableEnchant: func(obj *server.Object, enchant server.EnchantID) {
				Nox_xxx_spellBuffOff_4FF5B0(obj, enchant)
			},
		},
	)
}
