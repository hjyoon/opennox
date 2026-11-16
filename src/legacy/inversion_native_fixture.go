package legacy

/*
#include "GAME4_2.h"
#include "spell_accept_4fd400.h"
#include <stddef.h>
typedef int(*inversion_callback_C)(int,nox_object_t*,nox_object_t*,nox_object_t*);
_Static_assert(_Generic(&sub_52BEB0,inversion_callback_C:1,default:0),"original four-argument Inversion ABI");
static void inversion_layout_C(uintptr_t*v){
 v[0]=sizeof(void*);v[1]=sizeof(nox_object_t);v[2]=offsetof(nox_object_t,x);v[3]=offsetof(nox_object_t,y);
 v[4]=sizeof(nox_spell_accept_arg_t);v[5]=offsetof(nox_spell_accept_arg_t,obj);v[6]=offsetof(nox_spell_accept_arg_t,x);v[7]=offsetof(nox_spell_accept_arg_t,y);
}
static int inversion_call_C(int route,int id,nox_object_t*second,nox_object_t*owner,nox_object_t*caster){
 if(route==1){inversion_callback_C f=sub_52BEB0;return f(id,second,owner,caster);}
 return sub_52BEB0(id,second,owner,caster);
}
*/
import "C"

import (
	"github.com/opennox/libs/spell"
	"github.com/opennox/opennox/v1/server"
	"unsafe"
)

func inversionNativeLayoutC() (out [8]uintptr) {
	C.inversion_layout_C((*C.uintptr_t)(unsafe.Pointer(&out[0])))
	return
}

func inversionNativeCall(route int, id int32, second, owner, caster *server.Object, arg *server.SpellAcceptArg, level int32) int32 {
	if route == 2 {
		return int32(Sub_52BEB0(spell.ID(id), second, owner, caster, arg, int(level)))
	}
	return int32(C.inversion_call_C(C.int(route), C.int(id), asObjectC(second), asObjectC(owner), asObjectC(caster)))
}
