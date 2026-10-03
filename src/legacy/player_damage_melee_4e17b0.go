package legacy

import (
	"github.com/opennox/libs/object"

	"github.com/opennox/opennox/v1/common/unit/ai"
	"github.com/opennox/opennox/v1/server"
)

func playerDamageMeleeRuntime4E17B0(s *server.Server) server.PlayerDamageMeleeRuntime4E17B0 {
	return server.PlayerDamageMeleeRuntime4E17B0{
		RandomInt:         s.Rand.Logic.IntClamp,
		StaffWalkingBlock: func() bool { return Get_gameex_flags()&4 != 0 },
		MonsterBlockAction: func(unit *server.Object) {
			unit.MonsterActionPushIfChanged50A360(ai.ACTION_WEAPON_BLOCK)
		},
		MonsterPopBlockAction: func(unit *server.Object) { unit.MonsterPopAction() },
		CanDamageBlockWeapon: func(item *server.Object) bool {
			if item == nil || item.HealthData == nil ||
				(item.Class().Has(object.ClassWeapon) && uint32(item.SubClass())&0x07800000 != 0) {
				return true
			}
			return canEquipDamageNative4E16D0(item)
		},
		DamageBlockWeapon: playerDamageWeaponNative4E1560,
	}
}
