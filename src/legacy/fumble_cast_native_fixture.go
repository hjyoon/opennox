package legacy

/*
#include "GAME4_2.h"
#include "spell_accept_4fd400.h"
#include <stddef.h>
typedef int(*fumble_cast_callback_C)(int,nox_object_t*,nox_object_t*,nox_object_t*,void*);
_Static_assert(_Generic(&nox_xxx_castFumble_52C060,fumble_cast_callback_C:1,default:0),"original five-argument Fumble native ABI");
static void fumble_cast_layout_C(uintptr_t*v){
 v[0]=sizeof(void*);v[1]=sizeof(nox_object_t);v[2]=sizeof(nox_spell_accept_arg_t);
 v[3]=offsetof(nox_spell_accept_arg_t,obj);v[4]=offsetof(nox_spell_accept_arg_t,x);v[5]=offsetof(nox_spell_accept_arg_t,y);
 v[6]=offsetof(nox_object_t,typ_ind);v[7]=offsetof(nox_object_t,obj_class);v[8]=offsetof(nox_object_t,obj_subclass);v[9]=offsetof(nox_object_t,obj_flags);
 v[10]=offsetof(nox_object_t,x);v[11]=offsetof(nox_object_t,y);v[12]=offsetof(nox_object_t,inv_first_item);v[13]=offsetof(nox_object_t,inv_next_item);
 v[14]=offsetof(nox_object_t,field_129);v[15]=offsetof(nox_object_t,field_128);
}
static int fumble_cast_call_C(int route,int id,nox_object_t*second,nox_object_t*owner,nox_object_t*caster,void*arg){
 if(route==1){fumble_cast_callback_C f=nox_xxx_castFumble_52C060;return f(id,second,owner,caster,arg);}
 return nox_xxx_castFumble_52C060(id,second,owner,caster,arg);
}
*/
import "C"

import (
	"unsafe"

	"github.com/opennox/libs/spell"
	"github.com/opennox/opennox/v1/server"
)

func fumbleCastNativeLayoutC() (out [16]uintptr) {
	C.fumble_cast_layout_C((*C.uintptr_t)(unsafe.Pointer(&out[0])))
	return
}

func fumbleCastNativeCall(route int, id int32, second, owner, caster *server.Object, arg *server.SpellAcceptArg, level int32) int32 {
	if route == 2 {
		return int32(Nox_xxx_castFumble_52C060(spell.ID(id), second, owner, caster, arg, int(level)))
	}
	return int32(C.fumble_cast_call_C(C.int(route), C.int(id), asObjectC(second), asObjectC(owner), asObjectC(caster), unsafe.Pointer(arg)))
}
