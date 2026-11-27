package legacy

/*
#include <stddef.h>
#include <stdint.h>
#include "defs.h"
#include "GAME1.h"
#include "GAME3_3.h"
#include "GAME4.h"
#include "GAME4_2.h"
#include "GAME5.h"
#include "map_push_units_52e040.h"

static float2* monster_death_nil_point_549d80(void) {
 // Pass the original nil-object field address to the consuming operation.
 return (float2*)(uintptr_t)offsetof(nox_object_t,x);
}
static void monster_death_push_549d80(float outer,float inner,float force,nox_object_t* source,uintptr_t callback,uintptr_t arg) {
 nox_xxx_mapPushUnitsAround_52E040(monster_death_nil_point_549d80(),outer,inner,force,source,callback,arg);
}
static void monster_death_damage_549d80(float outer,float inner,int damage,int type,nox_object_t* source,nox_object_t* exclude) {
 nox_xxx_mapDamageUnitsAround_4E25B0((float*)monster_death_nil_point_549d80(),outer,inner,damage,type,source,exclude);
}
static int monster_death_spark_549d80(uint32_t code) {
 return nox_xxx_netSparkExplosionFx_5231B0((float*)monster_death_nil_point_549d80(),(char)code);
}
*/
import "C"

import (
	"unsafe"

	"github.com/opennox/libs/object"
	noxflags "github.com/opennox/opennox/v1/common/flags"
	"github.com/opennox/opennox/v1/server"
)

func monsterDeathExplosionRuntime549D80() server.MonsterDeathExplosionRuntime549D80 {
	if override := monsterDeathOriginalRuntimeOverride549D80; override != nil {
		return *override
	}
	return server.MonsterDeathExplosionRuntime549D80{
		GameFlag: func(mask uint32) uint32 {
			if noxflags.HasGame(noxflags.GameFlag(mask)) {
				return 1
			}
			return 0
		},
		Push: func(owner *server.Object, a, b, force float32, source *server.Object, callback, arg uintptr) uint32 {
			if owner != nil {
				Nox_xxx_mapPushUnitsAroundPtr52E040(&owner.PosVec, a, b, force, source, int(callback), int(arg))
			} else {
				C.monster_death_push_549d80(C.float(a), C.float(b), C.float(force), asObjectC(source), C.uintptr_t(callback), C.uintptr_t(arg))
			}
			return 0
		},
		Damage: func(owner *server.Object, a, b float32, damage, typ uint32, source, excluded *server.Object) uint32 {
			if owner != nil {
				mapDamageUnitsAround4E25B0(&owner.PosVec, a, b, int32(damage), object.DamageType(int32(typ)), source, excluded)
			} else {
				C.monster_death_damage_549d80(C.float(a), C.float(b), C.int(int32(damage)), C.int(int32(typ)), asObjectC(source), asObjectC(excluded))
			}
			return 0
		},
		Spark: func(owner *server.Object, code uint32) uint32 {
			if owner != nil {
				return uint32(server.NetSparkExplosionFXNative5231B0(&owner.PosVec, byte(code), pointFXRuntimeForCall522FF0()))
			}
			return uint32(C.monster_death_spark_549d80(C.uint32_t(code)))
		},
		Point: func(code uint32, owner *server.Object) uint32 {
			if owner != nil {
				return uint32(server.NetSendPointFXNative522FF0(byte(code), &owner.PosVec, pointFXRuntimeForCall522FF0()))
			}
			return debrisPointFXAtAddress(code, unsafe.Offsetof(server.Object{}.PosVec))
		},
		Audio: func(id uint32, unit *server.Object, kind, arg uint32) uint32 {
			nox_xxx_aud_501960(int32(id), asObjectC(unit), int32(kind), int32(arg))
			return 0
		},
		Delete: func(unit *server.Object) uint32 {
			nox_xxx_delayedDeleteObject_4E5CC0(asObjectC(unit))
			return 0
		},
	}
}

//export sub_549D80
func sub_549D80(unit *nox_object_t) C.int {
	return C.int(asObjectS(unit).MonsterDeathExplosionNative549D80(0x549d80, monsterDeathExplosionRuntime549D80()))
}

//export sub_549E00
func sub_549E00(unit *nox_object_t) C.int {
	return C.int(asObjectS(unit).MonsterDeathExplosionNative549D80(0x549e00, monsterDeathExplosionRuntime549D80()))
}

//export sub_549E70
func sub_549E70(unit *nox_object_t) C.int {
	return C.int(asObjectS(unit).MonsterDeathExplosionNative549D80(0x549e70, monsterDeathExplosionRuntime549D80()))
}

//export sub_54A250
func sub_54A250(unit *nox_object_t) C.int {
	return C.int(asObjectS(unit).MonsterDeathExplosionNative549D80(0x54a250, monsterDeathExplosionRuntime549D80()))
}
