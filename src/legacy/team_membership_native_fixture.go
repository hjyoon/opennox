package legacy

/*
#include "GAME1_1.h"
#include <stddef.h>
#include <string.h>
typedef nox_object_team_t*(*membership_head_fn)(nox_team_t*);
typedef nox_object_team_t*(*membership_next_fn)(nox_object_team_t*);
_Static_assert(_Generic(&nox_xxx_teamCheckSmth_418C60,membership_head_fn:1,default:0),"native Team head pointer ABI");
_Static_assert(_Generic(&sub_418C70,membership_next_fn:1,default:0),"native ObjectTeam next pointer ABI");
static uintptr_t team_membership_query_C(int route,int kind,void*value){
 if(kind==0){
  if(route==1){membership_head_fn volatile f=nox_xxx_teamCheckSmth_418C60;return(uintptr_t)f(value);}
  return(uintptr_t)nox_xxx_teamCheckSmth_418C60(value);
 }
 if(route==1){membership_next_fn volatile f=sub_418C70;return(uintptr_t)f(value);}
 return(uintptr_t)sub_418C70(value);
}
static void team_membership_seed_C(nox_team_t*array){
 memset(array,0,17*sizeof(*array));
 for(int i=1;i<=2;i++){array[i].field_57=(unsigned char)i;array[i].ind=(unsigned char)i;array[i].active=1;}
}
static uint32_t team_membership_count_C(nox_team_t*t){return t->field_48;}
static void team_membership_set_count_C(nox_team_t*t,uint32_t value){t->field_48=value;}
static int team_membership_parent_link_C(void*value,unsigned char id){return nox_server_team_attach_object_native(value,id);}
static uintptr_t team_membership_parent_detach_C(void*value){return(uintptr_t)nox_server_team_detach_object_native(value);}
static void team_membership_layout_C(uintptr_t*v){
 v[0]=sizeof(void*);v[1]=sizeof(nox_team_t);v[2]=_Alignof(nox_team_t);v[3]=offsetof(nox_team_t,field_44);v[4]=sizeof(((nox_team_t*)0)->field_44);
 v[5]=sizeof(nox_object_team_t);v[6]=_Alignof(nox_object_team_t);v[7]=offsetof(nox_object_team_t,legacy_next);v[8]=sizeof(((nox_object_team_t*)0)->legacy_next);
 v[9]=offsetof(nox_object_team_t,id);v[10]=sizeof(((nox_object_team_t*)0)->id);v[11]=sizeof(membership_head_fn);v[12]=sizeof(membership_next_fn);
}
*/
import "C"

import "unsafe"

func teamMembershipNativeLayoutC() (out [13]uintptr) {
	C.team_membership_layout_C((*C.uintptr_t)(unsafe.Pointer(&out[0])))
	return
}

func teamMembershipNativeSeedC(array unsafe.Pointer) {
	C.team_membership_seed_C((*C.nox_team_t)(array))
}

func teamMembershipNativeCountC(team unsafe.Pointer) uint32 {
	return uint32(C.team_membership_count_C((*C.nox_team_t)(team)))
}

func teamMembershipNativeSetCountC(team unsafe.Pointer, count uint32) {
	C.team_membership_set_count_C((*C.nox_team_t)(team), C.uint32_t(count))
}

func teamMembershipNativeParentLinkC(value unsafe.Pointer, id byte) bool {
	return C.team_membership_parent_link_C(value, C.uchar(id)) != 0
}

func teamMembershipNativeParentDetachC(value unsafe.Pointer) uintptr {
	return uintptr(C.team_membership_parent_detach_C(value))
}

func teamMembershipNativeQuery(route, kind int, value unsafe.Pointer) uintptr {
	if route == 2 {
		if kind == 0 {
			return uintptr(unsafe.Pointer(nox_xxx_teamCheckSmth_418C60((*C.nox_team_t)(value))))
		}
		return uintptr(unsafe.Pointer(sub_418C70((*C.nox_object_team_t)(value))))
	}
	return uintptr(C.team_membership_query_C(C.int(route), C.int(kind), value))
}
