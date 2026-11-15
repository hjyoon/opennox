package legacy

/*
#include "GAME4_2.h"
#include "spell_accept_4fd400.h"
#include <stddef.h>
typedef int(*poison_cast_callback_C)(int,nox_object_t*,nox_object_t*,nox_object_t*,void*,int);
_Static_assert(_Generic(&nox_xxx_castPoison_52C720,poison_cast_callback_C:1,default:0),"native six-argument Poison ABI");
static void poison_cast_layout_C(uintptr_t* v){
 v[0]=sizeof(void*);v[1]=sizeof(nox_spell_accept_arg_t);v[2]=offsetof(nox_spell_accept_arg_t,obj);
 v[3]=offsetof(nox_spell_accept_arg_t,x);v[4]=offsetof(nox_spell_accept_arg_t,y);
 v[5]=offsetof(nox_object_t,obj_class);v[6]=offsetof(nox_object_t,obj_flags);
}
static int poison_cast_call_C(int route,int id,nox_object_t* second,nox_object_t* owner,nox_object_t* caster,void* arg,int level){
 if(route==1){poison_cast_callback_C f=nox_xxx_castPoison_52C720;return f(id,second,owner,caster,arg,level);}
 return nox_xxx_castPoison_52C720(id,second,owner,caster,arg,level);
}
*/
import "C"

import (
	"unsafe"

	"github.com/opennox/libs/spell"
	"github.com/opennox/opennox/v1/server"
)

func poisonCastNativeLayoutC() (out [7]uintptr) {
	C.poison_cast_layout_C((*C.uintptr_t)(unsafe.Pointer(&out[0])))
	return
}

func poisonCastNativeCall(route int, id int32, second, owner, caster *server.Object, arg *server.SpellAcceptArg, level int32) int32 {
	if route == 2 {
		return int32(Nox_xxx_castPoison_52C720(spell.ID(id), second, owner, caster, arg, int(level)))
	}
	return int32(C.poison_cast_call_C(C.int(route), C.int(id), asObjectC(second), asObjectC(owner), asObjectC(caster), unsafe.Pointer(arg), C.int(level)))
}
