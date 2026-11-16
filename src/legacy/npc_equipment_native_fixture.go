package legacy

/*
#include "GAME4_2.h"
#include <stddef.h>
typedef int(*npc_equipment_callback_C)(nox_object_t*);
_Static_assert(_Generic(&sub_52BA70,npc_equipment_callback_C:1,default:0),"native one-pointer NPC equipment ABI");
static void npc_equipment_layout_C(uintptr_t*v){
 v[0]=sizeof(void*);v[1]=sizeof(nox_object_t);
 v[2]=offsetof(nox_object_t,obj_class);v[3]=offsetof(nox_object_t,obj_subclass);v[4]=offsetof(nox_object_t,obj_flags);
 v[5]=offsetof(nox_object_t,inv_next_item);v[6]=offsetof(nox_object_t,inv_first_item);
 v[7]=sizeof(((nox_object_t*)0)->obj_class);v[8]=sizeof(((nox_object_t*)0)->obj_subclass);v[9]=sizeof(((nox_object_t*)0)->obj_flags);
 v[10]=sizeof(((nox_object_t*)0)->inv_next_item);v[11]=sizeof(((nox_object_t*)0)->inv_first_item);
 v[12]=_Alignof(nox_object_t);v[13]=sizeof(npc_equipment_callback_C);
}
static int npc_equipment_call_C(int route,nox_object_t*obj){
 if(route==1){npc_equipment_callback_C f=sub_52BA70;return f(obj);}
 return sub_52BA70(obj);
}
*/
import "C"

import (
	"unsafe"

	"github.com/opennox/opennox/v1/server"
)

func npcEquipmentNativeLayoutC() (out [14]uintptr) {
	C.npc_equipment_layout_C((*C.uintptr_t)(unsafe.Pointer(&out[0])))
	return
}

func npcEquipmentNativeCall(route int, obj *server.Object) int32 {
	if route == 2 {
		npcNormalizeEquipped52BA70(obj)
		return 0
	}
	return int32(C.npc_equipment_call_C(C.int(route), asObjectC(obj)))
}
