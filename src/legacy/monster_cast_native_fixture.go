package legacy

/*
#include "GAME4_3.h"
#include <stddef.h>
typedef char (*monster_cast_callback_C)(int,nox_object_t*,nox_spell_accept_arg_t*);
typedef char (*monster_action_cast_callback_C)(nox_object_t*,int);
typedef void (*monster_recoil_callback_C)(nox_object_t*,nox_object_t*,float2*);
_Static_assert(_Generic(&nox_xxx_mobCast_541300,monster_cast_callback_C:1,default:0),"native direct cast ABI");
_Static_assert(_Generic(&nox_xxx_mobActionCast_5413B0,monster_action_cast_callback_C:1,default:0),"native action cast ABI");
_Static_assert(_Generic(&nox_xxx_mobCastRandomRecoil_541490,monster_recoil_callback_C:1,default:0),"native recoil ABI");
typedef struct monster_cast_action_layout_C {uint32_t action;uintptr_t args[4];uint32_t field5;} monster_cast_action_layout_C;
typedef struct monster_cast_update_layout_C {
 uint8_t fixed300[300];void* waypoints[16];uint32_t fields91_102[12];uint16_t health103[32];
 void* anim119;uint8_t animation[4];void* definition;void* soundset;uint32_t fields123_135[13];
 int8_t index;uint8_t byte136_1;uint16_t word136_2;uint32_t field137;monster_cast_action_layout_C stack[24];
 uint32_t stamina_and_seen_count;void* seen[16];void* current_enemy;uint32_t fields300_303[4];void* preferred_enemy;uint32_t field305;
 uint32_t scripts306_325[20];float fields326_359[34];uint32_t status;
} monster_cast_update_layout_C;
static void monster_cast_layout_C(uintptr_t* v) {
 v[0]=sizeof(nox_object_t);v[1]=offsetof(nox_object_t,obj_class);v[2]=offsetof(nox_object_t,obj_subclass);
 v[3]=offsetof(nox_object_t,x);v[4]=offsetof(nox_object_t,vel_x);v[5]=offsetof(nox_object_t,direction1);v[6]=offsetof(nox_object_t,direction2);v[7]=offsetof(nox_object_t,data_update);
 v[8]=sizeof(nox_spell_accept_arg_t);v[9]=offsetof(nox_spell_accept_arg_t,obj);v[10]=offsetof(nox_spell_accept_arg_t,x);
 v[11]=sizeof(monster_cast_action_layout_C);v[12]=offsetof(monster_cast_action_layout_C,args);v[13]=offsetof(monster_cast_action_layout_C,field5);
 v[14]=offsetof(monster_cast_update_layout_C,animation)+1;v[15]=offsetof(monster_cast_update_layout_C,animation)+2;
 v[16]=offsetof(monster_cast_update_layout_C,definition);v[17]=offsetof(monster_cast_update_layout_C,soundset);v[18]=offsetof(monster_cast_update_layout_C,index);v[19]=offsetof(monster_cast_update_layout_C,stack);
 v[20]=offsetof(monster_cast_update_layout_C,fields326_359)+4*sizeof(float);v[21]=offsetof(monster_cast_update_layout_C,status);v[22]=offsetof(nox_player_update_data_t,field_73);v[23]=216;
}
static char monster_direct_call_C(int callback,int id,nox_object_t* unit,nox_spell_accept_arg_t* arg) {
 if(callback){monster_cast_callback_C f=nox_xxx_mobCast_541300;return f(id,unit,arg);}return nox_xxx_mobCast_541300(id,unit,arg);
}
static char monster_action_call_C(int callback,nox_object_t* unit,int mode) {
 if(callback){monster_action_cast_callback_C f=nox_xxx_mobActionCast_5413B0;return f(unit,mode);}return nox_xxx_mobActionCast_5413B0(unit,mode);
}
static void monster_recoil_call_C(int callback,nox_object_t* unit,nox_object_t* target,float2* out) {
 if(callback){monster_recoil_callback_C f=nox_xxx_mobCastRandomRecoil_541490;f(unit,target,out);}else nox_xxx_mobCastRandomRecoil_541490(unit,target,out);
}
*/
import "C"
import "unsafe"

import (
	"github.com/opennox/libs/types"
	"github.com/opennox/opennox/v1/server"
)

func monsterCastLayoutC() (v [24]uintptr) {
	C.monster_cast_layout_C((*C.uintptr_t)(unsafe.Pointer(&v[0])))
	return v
}

func monsterCastActualC(route int, root string, id int32, unit, target *server.Object, arg *server.SpellAcceptArg, mode int32, out *types.Pointf) byte {
	if route == 2 {
		switch root {
		case "direct":
			return Nox_xxx_mobCast_541300(id, unit, arg)
		case "action":
			Nox_xxx_mobActionCast_5413B0(unit, int(mode))
		default:
			Nox_xxx_mobCastRandomRecoil_541490(unit, target, out)
		}
		return 0
	}
	switch root {
	case "direct":
		return byte(C.monster_direct_call_C(C.int(route), C.int(id), asObjectC(unit), (*C.nox_spell_accept_arg_t)(unsafe.Pointer(arg))))
	case "action":
		return byte(C.monster_action_call_C(C.int(route), asObjectC(unit), C.int(mode)))
	default:
		C.monster_recoil_call_C(C.int(route), asObjectC(unit), asObjectC(target), (*C.float2)(unsafe.Pointer(out)))
		return 0
	}
}
