package legacy

/*
#include "client__gui__window.h"
#include "GAME5_2.h"
#include <stddef.h>
#include <string.h>
typedef int (*flag_window_fn)(nox_window*);
typedef int (*flag_waypoint_fn)(nox_waypoint_t*,unsigned char);
_Static_assert(_Generic(&nox_xxx_wndGetFlags_46ADA0,flag_window_fn:1,default:0),"unchanged window flags ABI");
_Static_assert(_Generic(&sub_579EE0,flag_waypoint_fn:1,default:0),"unchanged waypoint flag mask ABI");
static int flag_predicate_call_C(int route,int kind,void* owner,unsigned char mask){
 if(kind==0){
  if(route==1){flag_window_fn volatile f=nox_xxx_wndGetFlags_46ADA0;return f(owner);}
  return nox_xxx_wndGetFlags_46ADA0(owner);
 }
 if(route==1){flag_waypoint_fn volatile f=sub_579EE0;return f(owner,mask);}
 return sub_579EE0(owner,mask);
}
static void flag_predicate_seed_C(int kind,void*owner,uint32_t value,uint32_t opaque){
 if(kind==0){nox_window*w=owner;memcpy(&w->id,&opaque,4);w->flags=value;}
 else{nox_waypoint_t*w=owner;w->ind=opaque;w->flags_2=(unsigned char)value;}
}
static void flag_predicate_layout_C(uintptr_t*v){
 v[0]=sizeof(void*);v[1]=sizeof(nox_window);v[2]=_Alignof(nox_window);
 v[3]=offsetof(nox_window,flags);v[4]=sizeof(((nox_window*)0)->flags);
 v[5]=sizeof(nox_waypoint_t);v[6]=_Alignof(nox_waypoint_t);
 v[7]=offsetof(nox_waypoint_t,flags_2);v[8]=sizeof(((nox_waypoint_t*)0)->flags_2);
 v[9]=sizeof(flag_window_fn);v[10]=sizeof(flag_waypoint_fn);
}
*/
import "C"

import "unsafe"

func flagPredicateNativeLayoutC() (out [11]uintptr) {
	C.flag_predicate_layout_C((*C.uintptr_t)(unsafe.Pointer(&out[0])))
	return
}

func flagPredicateNativeSeedC(kind int, owner unsafe.Pointer, value, opaque uint32) {
	C.flag_predicate_seed_C(C.int(kind), owner, C.uint32_t(value), C.uint32_t(opaque))
}

func flagPredicateNativeCall(route, kind int, owner unsafe.Pointer, mask byte) uint32 {
	if route == 2 {
		if kind == 0 {
			return uint32(nox_xxx_wndGetFlags_46ADA0((*C.nox_window)(owner)))
		}
		return uint32(sub_579EE0((*C.nox_waypoint_t)(owner), C.uchar(mask)))
	}
	return uint32(C.flag_predicate_call_C(C.int(route), C.int(kind), owner, C.uchar(mask)))
}
