package legacy

/*
#include "GAME4_2.h"
#include <stddef.h>
typedef int (*fireball_callback_C)(int,void*,nox_object_t*,nox_object_t*,void*,int);
_Static_assert(_Generic(&nox_xxx_castFireball_52C790,fireball_callback_C:1,default:0),"native Fireball callback ABI");
static fireball_callback_C fireball_entry_C(void) { return nox_xxx_castFireball_52C790; }
static int fireball_call_C(int id,void* second,nox_object_t* caster,nox_object_t* firing,void* arg,int level) {
    return nox_xxx_castFireball_52C790(id,second,caster,firing,arg,level);
}
static void fireball_layout_C(uintptr_t* v) {
    v[0]=sizeof(nox_object_t);v[1]=offsetof(nox_object_t,x);v[2]=offsetof(nox_object_t,y);
    v[3]=offsetof(nox_object_t,vel_x);v[4]=offsetof(nox_object_t,vel_y);
    v[5]=offsetof(nox_object_t,shape)+offsetof(nox_shape,circle_r);v[6]=offsetof(nox_object_t,speed_cur);
    v[7]=offsetof(nox_object_t,direction1);v[8]=offsetof(nox_object_t,direction2);v[9]=sizeof(void*);
}
*/
import "C"
import (
	"unsafe"

	"github.com/opennox/libs/spell"
	"github.com/opennox/opennox/v1/server"
)

func fireballLayoutC() (out [10]uintptr) {
	C.fireball_layout_C((*C.uintptr_t)(unsafe.Pointer(&out[0])))
	return out
}

func fireballActualC(route int, id int32, second, caster, firing *server.Object, arg *server.SpellAcceptArg, level int32) int32 {
	if route == 0 {
		return int32(C.fireball_call_C(C.int(id), unsafe.Pointer(second), asObjectC(caster), asObjectC(firing), unsafe.Pointer(arg), C.int(level)))
	}
	if route == 1 {
		return int32(Nox_spells_call_intint6_go(unsafe.Pointer(C.fireball_entry_C()), spell.ID(id), second, caster, firing, arg, int(level)))
	}
	return int32(Nox_xxx_castFireball_52C790(spell.ID(id), second, caster, firing, arg, int(level)))
}
