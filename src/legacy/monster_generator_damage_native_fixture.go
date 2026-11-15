package legacy

/*
#include "GAME3_3.h"
#include "GAME4_2.h"
#include <stddef.h>
typedef int (*generator_damage_callback_C)(nox_object_t*,nox_object_t*,nox_object_t*,int,int);
_Static_assert(_Generic(&nox_xxx_damageMonsterGen_4E27D0,generator_damage_callback_C:1,default:0),"native generator damage ABI");
typedef struct generator_damage_health_C {uint16_t current,unused,maximum,field6;uint32_t field8,field12,field16;} generator_damage_health_C;
typedef struct generator_damage_update_C {
 nox_object_t* children[12];uint32_t damage_flags,damage_function,spawn_flags,spawn_function,death_flags,death_function;
 uint32_t collision[2];uint8_t spawn[3],quest_spawn[3],active,maximum;uint32_t frame88,field92;uint16_t health_samples[32];uint32_t field160;
} generator_damage_update_C;
static void generator_damage_layout_C(uintptr_t* v) {
 v[0]=sizeof(nox_object_t);v[1]=offsetof(nox_object_t,obj_flags);v[2]=offsetof(nox_object_t,field_5);
 v[3]=offsetof(nox_object_t,x);v[4]=offsetof(nox_object_t,y);v[5]=offsetof(nox_object_t,field_134);
 v[6]=offsetof(nox_object_t,health_data);v[7]=offsetof(nox_object_t,data_update);v[8]=offsetof(nox_object_t,func_damage);
 v[9]=sizeof(generator_damage_health_C);v[10]=offsetof(generator_damage_health_C,current);v[11]=offsetof(generator_damage_health_C,maximum);
 v[12]=sizeof(generator_damage_update_C);v[13]=offsetof(generator_damage_update_C,damage_flags);v[14]=offsetof(generator_damage_update_C,damage_function);
 v[15]=sizeof(((generator_damage_update_C*)0)->collision);v[16]=offsetof(generator_damage_health_C,unused);v[17]=sizeof(float2);v[18]=offsetof(float2,field_4);
}
static void generator_damage_bind_C(nox_object_t* unit) {unit->func_damage=nox_xxx_damageMonsterGen_4E27D0;}
static int generator_damage_call_C(int callback,nox_object_t* unit,nox_object_t* source,nox_object_t* weapon,int damage,int typ) {
 if(callback){generator_damage_callback_C f=nox_xxx_damageMonsterGen_4E27D0;return f(unit,source,weapon,damage,typ);}
 return nox_xxx_damageMonsterGen_4E27D0(unit,source,weapon,damage,typ);
}
// Independent pre-port C wire control, frozen at c283b909935784c9a61c4244191cebf1c4cb6d07.
// Only root/converter symbols are renamed; production exports may now move to Go.
static int generator_damage_float2int_frozen_419A70(float a1) {
	// Original419A70 returns integer-indefinite for invalid signed-DWORD input.
	// Native ARM64 conversion otherwise saturates or returns zero for NaN.
	if (!(a1 >= -2147483648.0f && a1 < 2147483648.0f)) {
		return INT32_MIN;
	}
	return (int)a1;
}
static int generator_damage_point_fx_frozen_523150(char a1, char a2, float* a3) {
	char v4[6]; // [esp+4h] [ebp-8h]

	v4[0] = a1;
	v4[1] = a2;
	*(uint16_t*)&v4[2] = generator_damage_float2int_frozen_419A70(*a3);
	*(uint16_t*)&v4[4] = generator_damage_float2int_frozen_419A70(a3[1]);
	return nox_xxx_netSendFxAllCli_523030((float2*)a3, v4, 6);
}
*/
import "C"
import "unsafe"

import (
	"github.com/opennox/libs/object"
	"github.com/opennox/libs/types"
	"github.com/opennox/opennox/v1/server"
)

func generatorDamageLayoutC() (v [19]uintptr) {
	C.generator_damage_layout_C((*C.uintptr_t)(unsafe.Pointer(&v[0])))
	return v
}

func generatorDamageBindC(unit *server.Object) {
	C.generator_damage_bind_C(asObjectC(unit))
}

func generatorDamagePointFXExistingC(op, subtype byte, point *types.Pointf) {
	C.generator_damage_point_fx_frozen_523150(C.char(op), C.char(subtype), (*C.float)(unsafe.Pointer(point)))
}

func generatorDamageActualCall(route int, unit, source, weapon *server.Object, damage, typ int32) int32 {
	if route == 2 {
		if unit.CallDamage(source, weapon, int(damage), object.DamageType(typ)) {
			return 1
		}
		return 0
	}
	return int32(C.generator_damage_call_C(C.int(route), asObjectC(unit), asObjectC(source), asObjectC(weapon), C.int(damage), C.int(typ)))
}
