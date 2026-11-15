package legacy

/*
#include "GAME3_3.h"
*/
import "C"

import (
	"github.com/opennox/opennox/v1/common/sound"
	"github.com/opennox/opennox/v1/server"
)

var monsterGeneratorDamageRuntimeForCall4E27D0 = monsterGeneratorDamageDefaultRuntime4E27D0

func monsterGeneratorDamageDefaultRuntime4E27D0() server.MonsterGeneratorDamageNativeRuntime4E27D0 {
	legacyServer := GetServer()
	s := legacyServer.S()
	return server.MonsterGeneratorDamageNativeRuntime4E27D0{
		Frame:   s.Frame,
		PointFX: s.SendMonsterGeneratorDamagePointFX4E27D0,
		Audio: func(id int32, unit *server.Object) {
			s.Audio.EventObj(sound.ID(id), unit, 0, 0)
		},
		GetHP: server.UnitGetHP4EE780,
		Default: func(unit, source, weapon *server.Object, damage, typ int32) int32 {
			return int32(nox_xxx_damageDefaultProc_4E0B30_go(asObjectC(unit), asObjectC(source), asObjectC(weapon), C.int(damage), C.int(typ)))
		},
		Script: func(block *server.ScriptCallback, source, unit *server.Object, event server.ScriptEventType) {
			legacyServer.NoxScriptC().ScriptCallback(block, source, unit, event)
		},
		SetStatus:   (*server.Object).SetXStatus,
		UnsetStatus: (*server.Object).UnsetXStatus,
	}
}

func Nox_xxx_damageMonsterGen_4E27D0(unit, source, weapon *server.Object, damage, typ int32) int32 {
	return server.MonsterGeneratorDamageNative4E27D0(unit, source, weapon, damage, typ, monsterGeneratorDamageRuntimeForCall4E27D0())
}

//export nox_xxx_damageMonsterGen_4E27D0
func nox_xxx_damageMonsterGen_4E27D0(unit, source, weapon *C.nox_object_t, damage, typ C.int) C.int {
	return C.int(Nox_xxx_damageMonsterGen_4E27D0(asObjectS((*nox_object_t)(unit)), asObjectS((*nox_object_t)(source)), asObjectS((*nox_object_t)(weapon)), int32(damage), int32(typ)))
}
