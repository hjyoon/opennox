package legacy

/*
#include "GAME4_2.h"
#include "spell_accept_4fd400.h"
#include <stddef.h>
typedef int (*meteor_cast_callback_C)(int,nox_object_t*,nox_object_t*,nox_object_t*,void*,int);
_Static_assert(_Generic(&nox_xxx_castMeteorShower_52D8A0,meteor_cast_callback_C:1,default:0),"native shower ABI");
_Static_assert(_Generic(&nox_xxx_castMeteor_52D9D0,meteor_cast_callback_C:1,default:0),"native meteor ABI");
typedef struct { int32_t damage; } meteor_cast_damage_prefix_C;
static void meteor_cast_layout_C(uintptr_t* v) {
 v[0]=offsetof(nox_object_t,typ_ind);v[1]=offsetof(nox_object_t,obj_class);v[2]=offsetof(nox_object_t,field_5);
 v[3]=offsetof(nox_object_t,x);v[4]=offsetof(nox_object_t,y);v[5]=offsetof(nox_object_t,z);v[6]=offsetof(nox_object_t,field_27);
 v[7]=offsetof(nox_object_t,field_128);v[8]=offsetof(nox_object_t,field_129);v[9]=offsetof(nox_object_t,data_update);v[10]=offsetof(nox_object_t,owner);
 for(int i=0;i<11;++i)v[11+i]=4;v[11]=sizeof(((nox_object_t*)0)->typ_ind);
 v[18]=sizeof(((nox_object_t*)0)->field_128);v[19]=sizeof(((nox_object_t*)0)->field_129);v[20]=sizeof(((nox_object_t*)0)->data_update);v[21]=sizeof(((nox_object_t*)0)->owner);
 v[22]=offsetof(nox_spell_accept_arg_t,obj);v[23]=offsetof(nox_spell_accept_arg_t,x);v[24]=offsetof(nox_spell_accept_arg_t,y);
 v[25]=sizeof(((nox_spell_accept_arg_t*)0)->obj);v[26]=sizeof(((nox_spell_accept_arg_t*)0)->x);v[27]=sizeof(((nox_spell_accept_arg_t*)0)->y);v[28]=sizeof(nox_spell_accept_arg_t);
 v[29]=offsetof(nox_player_update_data_t,player);v[30]=offsetof(nox_playerInfo,playerInd);
 v[31]=sizeof(nox_object_t);v[32]=sizeof(nox_player_update_data_t);v[33]=sizeof(nox_playerInfo);v[34]=sizeof(void*);
 v[35]=offsetof(meteor_cast_damage_prefix_C,damage);v[36]=sizeof(((meteor_cast_damage_prefix_C*)0)->damage);
}
static int meteor_cast_call_C(int route,int root,int id,nox_object_t* second,nox_object_t* owner,nox_object_t* caster,void* arg,int level) {
 if(route==1){meteor_cast_callback_C f=root?nox_xxx_castMeteor_52D9D0:nox_xxx_castMeteorShower_52D8A0;return f(id,second,owner,caster,arg,level);}
 return root?nox_xxx_castMeteor_52D9D0(id,second,owner,caster,arg,level):nox_xxx_castMeteorShower_52D8A0(id,second,owner,caster,arg,level);
}
*/
import "C"

import (
	"unsafe"

	"github.com/opennox/libs/spell"
	"github.com/opennox/opennox/v1/server"
)

func meteorCastLayoutC() (out [37]uintptr) {
	C.meteor_cast_layout_C((*C.uintptr_t)(unsafe.Pointer(&out[0])))
	return
}

func meteorCastActualCall(route, root int, id int32, second, owner, caster *server.Object, arg *server.SpellAcceptArg, level int32) int32 {
	if route == 2 {
		if root == 0 {
			return int32(Nox_xxx_castMeteorShower_52D8A0(spell.ID(id), second, owner, caster, arg, int(level)))
		}
		return int32(Nox_xxx_castMeteor_52D9D0(spell.ID(id), second, owner, caster, arg, int(level)))
	}
	return int32(C.meteor_cast_call_C(C.int(route), C.int(root), C.int(id), asObjectC(second), asObjectC(owner), asObjectC(caster), unsafe.Pointer(arg), C.int(level)))
}
