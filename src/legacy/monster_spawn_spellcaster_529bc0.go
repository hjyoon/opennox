package legacy

/*
#include "GAME4_2.h"
*/
import "C"

import (
	"github.com/opennox/libs/object"
	"github.com/opennox/opennox/v1/server"
)

//export nox_xxx_monsterOnSpawnSpellcaster_529BC0
func nox_xxx_monsterOnSpawnSpellcaster_529BC0(unit *nox_object_t) {
	monsterOnSpawnNative529BC0(asObjectS(unit), func() *server.Server { return GetServer().S() })
}

func monsterOnSpawnNative529BC0(obj *server.Object, getServer func() *server.Server) {
	if obj == nil {
		return
	}
	ud := obj.UpdateDataMonster()
	def := Nox_xxx_monsterDefByTT_517560(int(obj.TypeInd))
	if def == nil {
		return
	}
	if byte(ud.Field361>>8) == 0 {
		if obj.HealthData == nil {
			return
		}
		health := uint16(def.Health68)
		obj.HealthData.Cur = health
		obj.HealthData.Max = health
		obj.HealthData.Field2 = health
	}
	if byte(ud.Field361) == 1 {
		ud.StatusFlags = def.StatusFlags92
	} else {
		const mutable = object.MonsterStatus((1<<22)-1) &^ object.MonsterStatus(0x19c40)
		ud.StatusFlags = ud.StatusFlags&^mutable | def.StatusFlags92&mutable
	}
	if ud.StatusFlags&0x20 == 0 {
		ud.StatusFlags &^= 0x1800
	}
	if byte(ud.Field335) == 1 {
		ud.RetreatLevel = def.RetreatRatio80
	}
	if byte(ud.Field337) == 1 {
		ud.ResumeLevel = def.ResumeRatio84
	}
	if byte(ud.Field509) == 1 {
		getServer().MonsterAutoSpells54C0C0(obj)
	}
}
