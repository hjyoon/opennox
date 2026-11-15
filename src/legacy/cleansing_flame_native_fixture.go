package legacy

/*
#include "GAME4_2.h"
#include <stddef.h>
typedef int (*cleansing_flame_callback_C)(int,nox_object_t*,nox_object_t*,nox_object_t*,void*,int);
_Static_assert(_Generic(&nox_xxx_spellCastCleansingFlame_52D5C0,cleansing_flame_callback_C:1,default:0),"native CleansingFlame ABI");
static void cleansing_flame_layout_C(uintptr_t* v) {
 v[0]=offsetof(nox_object_t,typ_ind);v[1]=offsetof(nox_object_t,obj_class);
 v[2]=offsetof(nox_object_t,x);v[3]=offsetof(nox_object_t,y);
 v[4]=offsetof(nox_object_t,vel_x);v[5]=offsetof(nox_object_t,vel_y);v[6]=offsetof(nox_object_t,float_28);
 v[7]=offsetof(nox_object_t,direction1);v[8]=offsetof(nox_object_t,direction2);v[9]=offsetof(nox_object_t,field_34);
 v[10]=offsetof(nox_object_t,float_39);v[11]=offsetof(nox_object_t,float_40);v[12]=offsetof(nox_object_t,shape)+offsetof(nox_shape,circle_r);
 v[13]=offsetof(nox_object_t,field_128);v[14]=offsetof(nox_object_t,field_129);v[15]=offsetof(nox_object_t,func_update);
 for(int i=0;i<16;++i)v[16+i]=4;
 v[16]=sizeof(((nox_object_t*)0)->typ_ind);v[23]=sizeof(((nox_object_t*)0)->direction1);v[24]=sizeof(((nox_object_t*)0)->direction2);
 v[29]=sizeof(((nox_object_t*)0)->field_128);v[30]=sizeof(((nox_object_t*)0)->field_129);v[31]=sizeof(((nox_object_t*)0)->func_update);
 v[32]=sizeof(nox_object_t);v[33]=sizeof(void*);
}
static int cleansing_flame_call_C(int route,int id,nox_object_t* second,nox_object_t* recipient,nox_object_t* caster,void* arg,int level) {
 if(route==1){cleansing_flame_callback_C f=nox_xxx_spellCastCleansingFlame_52D5C0;return f(id,second,recipient,caster,arg,level);}
 return nox_xxx_spellCastCleansingFlame_52D5C0(id,second,recipient,caster,arg,level);
}
*/
import "C"

import (
	"unsafe"

	"github.com/opennox/libs/spell"
	"github.com/opennox/opennox/v1/server"
)

func cleansingFlameLayoutC() (out [34]uintptr) {
	C.cleansing_flame_layout_C((*C.uintptr_t)(unsafe.Pointer(&out[0])))
	return
}

func cleansingFlameActualCall(route int, id int32, second, recipient, caster *server.Object, arg *server.SpellAcceptArg, level int32) int32 {
	if route == 2 {
		return int32(Nox_xxx_spellCastCleansingFlame_52D5C0(spell.ID(id), second, recipient, caster, arg, int(level)))
	}
	return int32(C.cleansing_flame_call_C(C.int(route), C.int(id), asObjectC(second), asObjectC(recipient), asObjectC(caster), unsafe.Pointer(arg), C.int(level)))
}
