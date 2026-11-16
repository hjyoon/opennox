package legacy

/*
#include "GAME3_3.h"
#include "GAME4_3.h"
typedef const nox_object_t object_trace_const_object_t;
extern uint32_t dword_5d4594_2488620;
typedef uint32_t(*object_trace_net_C)(const nox_object_t*);
typedef uint16_t(*object_trace_type_C)(const nox_object_t*);
typedef void*(*object_trace_init_C)(const nox_object_t*);
typedef const char*(*object_trace_id_C)(const nox_object_t*);
typedef nox_object_t*(*object_trace_next_C)(nox_object_t*);
typedef int2*(*object_trace_point_C)(void);
typedef char(*object_trace_flags_C)(nox_object_t*);
_Static_assert(_Generic(&nox_xxx_objectGetNetCode_4E4C00,object_trace_net_C:1,default:0),"Object uint32 NetCode ABI");
_Static_assert(_Generic(&nox_xxx_objectGetTypeInd_4E4C10,object_trace_type_C:1,default:0),"Object uint16 type ABI");
_Static_assert(_Generic(&nox_object_getInitData_4E4C30,object_trace_init_C:1,default:0),"Object nullable init ABI");
_Static_assert(_Generic(&nox_xxx_objectGetID_4E4C80,object_trace_id_C:1,default:0),"Object const ID ABI");
_Static_assert(_Generic(&nox_xxx_inventoryGetNext_4E7990,object_trace_next_C:1,default:0),"Object next ABI");
_Static_assert(_Generic(&sub_537760,object_trace_point_C:1,default:0),"full native Point32 return ABI");
_Static_assert(_Generic(&sub_53B300,object_trace_flags_C:1,default:0),"Object flags char return ABI");
static object_trace_net_C volatile object_trace_net_callback_C=nox_xxx_objectGetNetCode_4E4C00;
static object_trace_type_C volatile object_trace_type_callback_C=nox_xxx_objectGetTypeInd_4E4C10;
static object_trace_init_C volatile object_trace_init_callback_C=nox_object_getInitData_4E4C30;
static object_trace_id_C volatile object_trace_id_callback_C=nox_xxx_objectGetID_4E4C80;
static object_trace_next_C volatile object_trace_next_callback_C=nox_xxx_inventoryGetNext_4E7990;
static object_trace_point_C volatile object_trace_point_callback_C=sub_537760;
static object_trace_flags_C volatile object_trace_flags_callback_C=sub_53B300;
typedef struct object_trace_result_C {uint32_t word;void*pointer;} object_trace_result_C;
static object_trace_result_C object_trace_call_C(unsigned route,unsigned index,nox_object_t*owner){
 object_trace_result_C r={0,NULL};
 if(route==1){
  switch(index){
  case 0:{object_trace_net_C f=object_trace_net_callback_C;r.word=f(owner);break;}
  case 1:{object_trace_type_C f=object_trace_type_callback_C;r.word=f(owner);break;}
  case 2:{object_trace_init_C f=object_trace_init_callback_C;r.pointer=f(owner);break;}
  case 3:{object_trace_id_C f=object_trace_id_callback_C;r.pointer=(void*)f(owner);break;}
  case 4:{object_trace_next_C f=object_trace_next_callback_C;r.pointer=f(owner);break;}
  case 5:{object_trace_point_C f=object_trace_point_callback_C;r.pointer=f();break;}
  case 6:{object_trace_flags_C f=object_trace_flags_callback_C;r.word=(uint8_t)f(owner);break;}
  default:abort();
  }
  return r;
 }
 switch(index){
 case 0:r.word=nox_xxx_objectGetNetCode_4E4C00(owner);break;
 case 1:r.word=nox_xxx_objectGetTypeInd_4E4C10(owner);break;
 case 2:r.pointer=nox_object_getInitData_4E4C30(owner);break;
 case 3:r.pointer=(void*)nox_xxx_objectGetID_4E4C80(owner);break;
 case 4:r.pointer=nox_xxx_inventoryGetNext_4E7990(owner);break;
 case 5:r.pointer=sub_537760();break;
 case 6:r.word=(uint8_t)sub_53B300(owner);break;
 default:abort();
 }
 return r;
}
static void object_trace_seed_C(void*p,size_t placement,const uint32_t*q){
 unsigned char*raw=p;nox_object_t*owner=(nox_object_t*)(raw+0x1000+placement);
 void*ids[]={NULL,raw+0x6000+placement,raw+0x7000+placement};
 void*inits[]={NULL,raw+0x8000+placement,raw+0x9000+placement};
 nox_object_t*nexts[]={NULL,(nox_object_t*)(raw+0xa000+placement),(nox_object_t*)(raw+0xb000+placement)};
 owner->net_code=q[2];owner->typ_ind=(uint16_t)q[3];owner->obj_flags=q[4];
 owner->id=(char*)ids[q[5]];owner->init_data=inits[q[6]];owner->inv_next_item=nexts[q[7]];
 dword_5d4594_2488620=q[8];int2*point=(int2*)getMemAt(0x5D4594,2488612);
 point->field_0=(int32_t)q[9];point->field_4=(int32_t)q[10];
}
static void object_trace_layout_C(uintptr_t*v){
 v[0]=sizeof(void*);v[1]=sizeof(nox_object_t);v[2]=offsetof(nox_object_t,id);v[3]=offsetof(nox_object_t,typ_ind);
 v[4]=offsetof(nox_object_t,obj_flags);v[5]=offsetof(nox_object_t,net_code);v[6]=offsetof(nox_object_t,inv_next_item);v[7]=offsetof(nox_object_t,init_data);
 v[8]=sizeof(((nox_object_t*)0)->id);v[9]=sizeof(((nox_object_t*)0)->typ_ind);v[10]=sizeof(((nox_object_t*)0)->obj_flags);v[11]=sizeof(((nox_object_t*)0)->net_code);
 v[12]=sizeof(((nox_object_t*)0)->inv_next_item);v[13]=sizeof(((nox_object_t*)0)->init_data);v[14]=sizeof(int2);v[15]=offsetof(int2,field_0);v[16]=offsetof(int2,field_4);
 v[17]=sizeof(dword_5d4594_2488620);v[18]=sizeof(char);v[19]=sizeof(((int2*)0)->field_0);
}
*/
import "C"

import "unsafe"

func objectTraceNativeLayoutC() (out [20]uintptr) {
	C.object_trace_layout_C((*C.uintptr_t)(unsafe.Pointer(&out[0])))
	return
}

func objectTraceNativeSeedC(raw unsafe.Pointer, placement int, input *[11]uint32) {
	C.object_trace_seed_C(raw, C.size_t(placement), (*C.uint32_t)(unsafe.Pointer(&input[0])))
}

func objectTraceNativeCall(route, index int, owner unsafe.Pointer) (uint32, unsafe.Pointer) {
	o := (*C.nox_object_t)(owner)
	if route != 2 {
		r := C.object_trace_call_C(C.uint(route), C.uint(index), o)
		return uint32(r.word), r.pointer
	}
	readOnly := (*C.object_trace_const_object_t)(owner)
	switch index {
	case 0:
		return uint32(nox_xxx_objectGetNetCode_4E4C00(readOnly)), nil
	case 1:
		return uint32(nox_xxx_objectGetTypeInd_4E4C10(readOnly)), nil
	case 2:
		return 0, nox_object_getInitData_4E4C30(readOnly)
	case 3:
		return 0, unsafe.Pointer(nox_xxx_objectGetID_4E4C80(readOnly))
	case 4:
		return 0, unsafe.Pointer(nox_xxx_inventoryGetNext_4E7990(o))
	case 5:
		return 0, unsafe.Pointer(sub_537760())
	case 6:
		return uint32(uint8(sub_53B300(o))), nil
	}
	panic("invalid Object/trace entry")
}
