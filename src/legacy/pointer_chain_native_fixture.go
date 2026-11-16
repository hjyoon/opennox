package legacy

/*
#include "GAME3_2.h"
#include "GAME4_1.h"
#include <stddef.h>
typedef struct pc_node_C { nox_object_t*item;uint32_t cost;struct pc_node_C*next,*prev;}pc_node_C;
typedef struct pc_session_C {uint32_t f0,f4;nox_object_t*first,*second;uint32_t f16;pc_node_C*head20;uint32_t f24,f28;pc_node_C*headA,*headB;uint32_t valueA,valueB;nox_object_t*goldA,*goldB;struct pc_session_C*prev,*next;}pc_session_C;
typedef uint32_t*(*pc_attach_C)(uint32_t*,nox_object_t*);
typedef void*(*pc_detach_C)(uint32_t*,nox_object_t*);
typedef uint32_t*(*pc_stat_C)(nox_object_t*);
typedef void*(*pc_kill_C)(nox_object_t*);
typedef nox_object_t*(*pc_lookup_C)(nox_object_t*,int);
typedef void*(*pc_lesson_C)(nox_object_t*,int);
_Static_assert(_Generic(&nox_xxx_tradeSetPlayer_50F370,pc_attach_C:1,default:0),"native attach ABI");
_Static_assert(_Generic(&sub_50F490,pc_detach_C:1,default:0),"native detach ABI");
_Static_assert(_Generic(&sub_4D60E0,pc_stat_C:1,default:0),"native stat ABI");
_Static_assert(_Generic(&sub_4D6170,pc_kill_C:1,default:0),"native kill ABI");
_Static_assert(_Generic(&sub_510DE0,pc_lookup_C:1,default:0),"native item lookup ABI");
_Static_assert(_Generic(&nox_xxx_changeScore_4D8E90,pc_lesson_C:1,default:0),"native add lessons ABI");
_Static_assert(_Generic(&nox_xxx_playerSubLessons_4D8EC0,pc_lesson_C:1,default:0),"native subtract lessons ABI");
static pc_attach_C volatile pc_attach_ptr_C=nox_xxx_tradeSetPlayer_50F370;
static pc_detach_C volatile pc_detach_ptr_C=sub_50F490;
static pc_stat_C volatile pc_stat_ptr_C=sub_4D60E0;
static pc_kill_C volatile pc_kill_ptr_C=sub_4D6170;
static pc_lookup_C volatile pc_lookup_ptr_C=sub_510DE0;
static pc_lesson_C volatile pc_add_ptr_C=nox_xxx_changeScore_4D8E90,pc_sub_ptr_C=nox_xxx_playerSubLessons_4D8EC0;
static uintptr_t pc_call_C(int route,int index,void*session,nox_object_t*owner,int value){
 if(route==1){switch(index){
 case 2:{pc_attach_C f=pc_attach_ptr_C;return(uintptr_t)f(session,owner);}
 case 3:{pc_detach_C f=pc_detach_ptr_C;return(uintptr_t)f(session,owner);}
 case 4:{pc_stat_C f=pc_stat_ptr_C;return(uintptr_t)f(owner);}
 case 5:{pc_kill_C f=pc_kill_ptr_C;return(uintptr_t)f(owner);}
 case 6:{pc_lookup_C f=pc_lookup_ptr_C;return(uintptr_t)f(owner,value);}
 case 7:{pc_lesson_C f=pc_add_ptr_C;return(uintptr_t)f(owner,value);}
 case 8:{pc_lesson_C f=pc_sub_ptr_C;return(uintptr_t)f(owner,value);}
 }}
 switch(index){
 case 2:return(uintptr_t)nox_xxx_tradeSetPlayer_50F370(session,owner);
 case 3:return(uintptr_t)sub_50F490(session,owner);
 case 4:return(uintptr_t)sub_4D60E0(owner);
 case 5:return(uintptr_t)sub_4D6170(owner);
 case 6:return(uintptr_t)sub_510DE0(owner,value);
 case 7:return(uintptr_t)nox_xxx_changeScore_4D8E90(owner,value);
 case 8:return(uintptr_t)nox_xxx_playerSubLessons_4D8EC0(owner,value);
 default:return 0;
 }}
static void pc_seed_player_C(void*ptr,uint32_t value,uint32_t counter,uint32_t status,uint32_t gate){
 nox_playerInfo*p=ptr;p->lessons=(int32_t)value;p->field_4652=counter;p->field_4664=counter;p->field_4692=status;p->field_4792=gate;
}
static void pc_layout_C(uintptr_t*v){
 v[0]=sizeof(void*);v[1]=sizeof(nox_object_t);v[2]=offsetof(nox_object_t,obj_class);v[3]=sizeof(((nox_object_t*)0)->obj_class);v[4]=offsetof(nox_object_t,obj_flags);v[5]=sizeof(((nox_object_t*)0)->obj_flags);v[6]=offsetof(nox_object_t,data_update);v[7]=sizeof(((nox_object_t*)0)->data_update);
 v[8]=sizeof(nox_player_update_data_t);v[9]=offsetof(nox_player_update_data_t,player);v[10]=sizeof(((nox_player_update_data_t*)0)->player);v[11]=offsetof(nox_player_update_data_t,trade_70);v[12]=sizeof(((nox_player_update_data_t*)0)->trade_70);
 v[13]=sizeof(nox_playerInfo);v[14]=offsetof(nox_playerInfo,lessons);v[15]=sizeof(((nox_playerInfo*)0)->lessons);v[16]=offsetof(nox_playerInfo,field_4652);v[17]=sizeof(((nox_playerInfo*)0)->field_4652);v[18]=offsetof(nox_playerInfo,field_4664);v[19]=sizeof(((nox_playerInfo*)0)->field_4664);v[20]=offsetof(nox_playerInfo,field_4692);v[21]=sizeof(((nox_playerInfo*)0)->field_4692);v[22]=offsetof(nox_playerInfo,field_4792);v[23]=sizeof(((nox_playerInfo*)0)->field_4792);
 v[24]=sizeof(pc_session_C);v[25]=offsetof(pc_session_C,f0);v[26]=sizeof(((pc_session_C*)0)->f0);v[27]=offsetof(pc_session_C,head20);v[28]=sizeof(((pc_session_C*)0)->head20);
 v[29]=sizeof(pc_node_C);v[30]=offsetof(pc_node_C,item);v[31]=sizeof(((pc_node_C*)0)->item);v[32]=offsetof(pc_node_C,next);v[33]=sizeof(((pc_node_C*)0)->next);v[34]=offsetof(nox_object_t,net_code);v[35]=sizeof(((nox_object_t*)0)->net_code);v[36]=sizeof(int);v[37]=sizeof(pc_attach_C);
}
*/
import "C"

import (
	"unsafe"

	"github.com/opennox/opennox/v1/server"
)

func pointerChainNativeLayoutC() (out [38]uintptr) {
	C.pc_layout_C((*C.uintptr_t)(unsafe.Pointer(&out[0])))
	return
}

func pointerChainNativeSeedPlayer(p *server.Player, value, counter, status, gate uint32) {
	C.pc_seed_player_C(unsafe.Pointer(p), C.uint32_t(value), C.uint32_t(counter), C.uint32_t(status), C.uint32_t(gate))
}

func pointerChainNativeCall(route, index int, session *server.TradeSession, owner *server.Object, value uint32) uintptr {
	s, o := (*C.uint32_t)(unsafe.Pointer(session)), (*C.nox_object_t)(unsafe.Pointer(owner))
	if route != 2 {
		return uintptr(C.pc_call_C(C.int(route), C.int(index), unsafe.Pointer(session), o, C.int(value)))
	}
	switch index {
	case 2:
		return uintptr(unsafe.Pointer(nox_xxx_tradeSetPlayer_50F370(s, o)))
	case 3:
		return uintptr(sub_50F490(s, o))
	case 4:
		return uintptr(unsafe.Pointer(sub_4D60E0(o)))
	case 5:
		return uintptr(sub_4D6170(o))
	case 6:
		return uintptr(unsafe.Pointer(sub_510DE0(o, C.int(value))))
	case 7:
		return uintptr(nox_xxx_changeScore_4D8E90(o, C.int(value)))
	case 8:
		return uintptr(nox_xxx_playerSubLessons_4D8EC0(o, C.int(value)))
	}
	panic("invalid pointer chain entry")
}
