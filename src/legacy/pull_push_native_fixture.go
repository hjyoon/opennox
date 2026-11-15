package legacy

/*
#include "GAME4_2.h"
#include "spell_accept_4fd400.h"
#include <stddef.h>
typedef int(*pull_push_callback_C)(int,nox_object_t*,nox_object_t*,nox_object_t*,void*,int);
_Static_assert(_Generic(&nox_xxx_castPull_52BFA0,pull_push_callback_C:1,default:0),"native Pull ABI");
_Static_assert(_Generic(&nox_xxx_castPush_52C000,pull_push_callback_C:1,default:0),"native Push ABI");
static void pull_push_layout_C(uintptr_t*v){
 v[0]=sizeof(void*);v[1]=sizeof(nox_spell_accept_arg_t);v[2]=offsetof(nox_spell_accept_arg_t,obj);
 v[3]=offsetof(nox_spell_accept_arg_t,x);v[4]=offsetof(nox_spell_accept_arg_t,y);
 v[5]=offsetof(nox_object_t,x);v[6]=offsetof(nox_object_t,y);
}
static int pull_push_call_C(int route,int cast,int id,nox_object_t*second,nox_object_t*owner,nox_object_t*caster,void*arg,int level){
 if(route==1){pull_push_callback_C f=cast?nox_xxx_castPush_52C000:nox_xxx_castPull_52BFA0;return f(id,second,owner,caster,arg,level);}
 if(cast)return nox_xxx_castPush_52C000(id,second,owner,caster,arg,level);
 return nox_xxx_castPull_52BFA0(id,second,owner,caster,arg,level);
}
*/
import "C"

import (
	"unsafe"

	"github.com/opennox/libs/spell"
	"github.com/opennox/opennox/v1/server"
)

func pullPushNativeLayoutC() (out [7]uintptr) {
	C.pull_push_layout_C((*C.uintptr_t)(unsafe.Pointer(&out[0])))
	return
}

func pullPushNativeCall(route, cast int, id int32, second, owner, caster *server.Object, arg *server.SpellAcceptArg, level int32) int32 {
	if route == 2 {
		if cast != 0 {
			return int32(Nox_xxx_castPush_52C000(spell.ID(id), second, owner, caster, arg, int(level)))
		}
		return int32(Nox_xxx_castPull_52BFA0(spell.ID(id), second, owner, caster, arg, int(level)))
	}
	return int32(C.pull_push_call_C(C.int(route), C.int(cast), C.int(id), asObjectC(second), asObjectC(owner), asObjectC(caster), unsafe.Pointer(arg), C.int(level)))
}
