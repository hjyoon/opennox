package legacy

/*
#include "GAME4_2.h"
#include "spell_accept_4fd400.h"
#include <stddef.h>
typedef int(*burn_cast_callback_C)(int,nox_object_t*,nox_object_t*,nox_object_t*,void*);
_Static_assert(_Generic(&nox_xxx_castBurn_52C3E0,burn_cast_callback_C:1,default:0),"original five-argument Burn native ABI");
static void burn_cast_layout_C(uintptr_t*v){
 v[0]=sizeof(void*);v[1]=sizeof(nox_object_t);v[2]=sizeof(nox_spell_accept_arg_t);
 v[3]=offsetof(nox_spell_accept_arg_t,obj);v[4]=offsetof(nox_spell_accept_arg_t,x);v[5]=offsetof(nox_spell_accept_arg_t,y);
 v[6]=offsetof(nox_object_t,typ_ind);v[7]=offsetof(nox_object_t,obj_class);v[8]=offsetof(nox_object_t,x);v[9]=offsetof(nox_object_t,y);
 v[10]=offsetof(nox_object_t,data_update);v[11]=sizeof(((nox_object_t*)0)->data_update);
 v[12]=sizeof(nox_player_update_data_t);v[13]=offsetof(nox_player_update_data_t,player);v[14]=sizeof(((nox_player_update_data_t*)0)->player);
 v[15]=sizeof(nox_playerInfo);v[16]=offsetof(nox_playerInfo,playerInd);v[17]=sizeof(((nox_playerInfo*)0)->playerInd);v[18]=sizeof(((nox_object_t*)0)->typ_ind);
}
static int burn_cast_call_C(int route,int id,nox_object_t*second,nox_object_t*owner,nox_object_t*caster,void*arg){
 if(route==1){burn_cast_callback_C f=nox_xxx_castBurn_52C3E0;return f(id,second,owner,caster,arg);}
 return nox_xxx_castBurn_52C3E0(id,second,owner,caster,arg);
}
*/
import "C"

import (
	"unsafe"

	"github.com/opennox/libs/spell"
	"github.com/opennox/opennox/v1/server"
)

func burnCastNativeLayoutC() (out [19]uintptr) {
	C.burn_cast_layout_C((*C.uintptr_t)(unsafe.Pointer(&out[0])))
	return
}

func burnCastNativeCall(route int, id int32, second, owner, caster *server.Object, arg *server.SpellAcceptArg, level int32) int32 {
	if route == 2 {
		return int32(Nox_xxx_castBurn_52C3E0(spell.ID(id), second, owner, caster, arg, int(level)))
	}
	return int32(C.burn_cast_call_C(C.int(route), C.int(id), asObjectC(second), asObjectC(owner), asObjectC(caster), unsafe.Pointer(arg)))
}
