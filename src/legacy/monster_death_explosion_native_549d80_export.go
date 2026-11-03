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

static float2* monster_death_live_point_549d80(nox_object_t* owner) {
 // Native-width arithmetic leaves a bad address to the world operation,
 // preserving the original flag-call prefix before a nil position fault.
 return (float2*)((uintptr_t)owner + offsetof(nox_object_t,x));
}
static void monster_death_push_549d80(nox_object_t* owner,float outer,float inner,float force,nox_object_t* source,uintptr_t callback,uintptr_t arg) {
 nox_xxx_mapPushUnitsAround_52E040(monster_death_live_point_549d80(owner),outer,inner,force,source,callback,arg);
}
static void monster_death_damage_549d80(nox_object_t* owner,float outer,float inner,int damage,int type,nox_object_t* source,nox_object_t* exclude) {
 nox_xxx_mapDamageUnitsAround_4E25B0((float*)monster_death_live_point_549d80(owner),outer,inner,damage,type,source,exclude);
}
static int monster_death_spark_549d80(nox_object_t* owner,uint32_t code) {
 return nox_xxx_netSparkExplosionFx_5231B0((float*)monster_death_live_point_549d80(owner),(char)code);
}
static int monster_death_point_549d80(uint32_t code,nox_object_t* owner) {
 return nox_xxx_netSendPointFx_522FF0((char)code,monster_death_live_point_549d80(owner));
}
*/
import "C"

import "github.com/opennox/opennox/v1/server"

func monsterDeathExplosionRuntime549D80() server.MonsterDeathExplosionRuntime549D80 {
	if override := monsterDeathOriginalRuntimeOverride549D80; override != nil {
		return *override
	}
	return server.MonsterDeathExplosionRuntime549D80{
		GameFlag: func(mask uint32) uint32 {
			if bool(C.nox_common_gameFlags_check_40A5C0(C.uint(mask))) {
				return 1
			}
			return 0
		},
		Push: func(owner *server.Object, a, b, force float32, source *server.Object, callback, arg uintptr) uint32 {
			C.monster_death_push_549d80(asObjectC(owner), C.float(a), C.float(b), C.float(force), asObjectC(source), C.uintptr_t(callback), C.uintptr_t(arg))
			return 0
		},
		Damage: func(owner *server.Object, a, b float32, damage, typ uint32, source, excluded *server.Object) uint32 {
			C.monster_death_damage_549d80(asObjectC(owner), C.float(a), C.float(b), C.int(int32(damage)), C.int(int32(typ)), asObjectC(source), asObjectC(excluded))
			return 0
		},
		Spark: func(owner *server.Object, code uint32) uint32 {
			return uint32(C.monster_death_spark_549d80(asObjectC(owner), C.uint32_t(code)))
		},
		Point: func(code uint32, owner *server.Object) uint32 {
			return uint32(C.monster_death_point_549d80(C.uint32_t(code), asObjectC(owner)))
		},
		Audio: func(id uint32, unit *server.Object, kind, arg uint32) uint32 {
			C.nox_xxx_aud_501960(C.int(int32(id)), asObjectC(unit), C.int(int32(kind)), C.int(int32(arg)))
			return 0
		},
		Delete: func(unit *server.Object) uint32 {
			C.nox_xxx_delayedDeleteObject_4E5CC0(asObjectC(unit))
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
