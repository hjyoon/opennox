package legacy

/*
#include "GAME2_1.h"
#include "GAME3_3.h"
extern uintptr_t array_5D4594_1049872[9];
extern uint32_t dword_5d4594_1049864,dword_5d4594_1062488;
extern nox_window* dword_5d4594_1062452;
typedef nox_drawable*(*inventory_state_find_C)(int);
typedef int(*inventory_state_test_C)(void);
typedef int(*inventory_state_bounds_C)(int,int);
typedef void(*inventory_state_reset_C)(void);
typedef nox_window*(*inventory_state_window_C)(void);
typedef int(*inventory_state_set_C)(int);
typedef nox_object_t*(*inventory_state_first_C)(nox_object_t*);
_Static_assert(_Generic(&sub_461600,inventory_state_find_C:1,default:0),"inventory type query ABI");
_Static_assert(_Generic(&sub_461930,inventory_state_test_C:1,default:0),"inventory class query ABI");
_Static_assert(_Generic(&sub_464B40,inventory_state_bounds_C:1,default:0),"inventory grid ABI");
_Static_assert(_Generic(&sub_467680,inventory_state_reset_C:1,default:0),"inventory void reset ABI");
_Static_assert(_Generic(&nox_xxx_wndGetHandle_4676A0,inventory_state_window_C:1,default:0),"inventory window ABI");
_Static_assert(_Generic(&sub_467740,inventory_state_set_C:1,default:0),"inventory state setter ABI");
_Static_assert(_Generic(&nox_xxx_inventoryGetFirst_4E7980,inventory_state_first_C:1,default:0),"Object inventory head ABI");
static inventory_state_find_C volatile inventory_state_find_callback_C=sub_461600;
static inventory_state_test_C volatile inventory_state_test_callback_C=sub_461930;
static inventory_state_bounds_C volatile inventory_state_bounds_callback_C=sub_464B40;
static inventory_state_reset_C volatile inventory_state_reset_callback_C=sub_467680;
static inventory_state_window_C volatile inventory_state_window_callback_C=nox_xxx_wndGetHandle_4676A0;
static inventory_state_set_C volatile inventory_state_set_callback_C=sub_467740;
static inventory_state_first_C volatile inventory_state_first_callback_C=nox_xxx_inventoryGetFirst_4E7980;
typedef struct inventory_state_result_C {uint32_t word;void* pointer;} inventory_state_result_C;
static inventory_state_result_C inventory_state_call_C(unsigned route,unsigned index,nox_object_t*owner,int query,int x,int y,int value){
 inventory_state_result_C r={0,NULL};
 if(route==1){
  switch(index){
  case 0:{inventory_state_find_C f=inventory_state_find_callback_C;r.pointer=f(query);break;}
  case 1:{inventory_state_test_C f=inventory_state_test_callback_C;r.word=(uint32_t)f();break;}
  case 2:{inventory_state_bounds_C f=inventory_state_bounds_callback_C;r.word=(uint32_t)f(x,y);break;}
  case 3:{inventory_state_reset_C f=inventory_state_reset_callback_C;f();break;}
  case 4:{inventory_state_window_C f=inventory_state_window_callback_C;r.pointer=f();break;}
  case 5:{inventory_state_set_C f=inventory_state_set_callback_C;r.word=(uint32_t)f(value);break;}
  case 6:{inventory_state_first_C f=inventory_state_first_callback_C;r.pointer=f(owner);break;}
  default:abort();
  }
  return r;
 }
 switch(index){
 case 0:r.pointer=sub_461600(query);break;
 case 1:r.word=(uint32_t)sub_461930();break;
 case 2:r.word=(uint32_t)sub_464B40(x,y);break;
 case 3:sub_467680();break;
 case 4:r.pointer=nox_xxx_wndGetHandle_4676A0();break;
 case 5:r.word=(uint32_t)sub_467740(value);break;
 case 6:r.pointer=nox_xxx_inventoryGetFirst_4E7980(owner);break;
 default:abort();
 }
 return r;
}
static void inventory_state_seed_C(void*p,size_t placement,const uint32_t*q){
 unsigned char*raw=p;nox_drawable*nodes[12];
 for(unsigned j=0;j<12;j++)nodes[j]=(nox_drawable*)(raw+0x1000+j*0x400+placement);
 for(unsigned j=0;j<12;j++){
  nodes[j]->field_27=q[18+3*j];nodes[j]->flags28=q[19+3*j];
  int32_t next=(int32_t)q[20+3*j];nodes[j]->field_92=next<0?NULL:nodes[next];
 }
 for(unsigned j=0;j<9;j++){int32_t head=(int32_t)q[9+j];array_5D4594_1049872[j]=head<0?0:(uintptr_t)nodes[head];}
 nox_object_t*owner=(nox_object_t*)(raw+0x8000+placement);
 nox_object_t*items[]={NULL,(nox_object_t*)(raw+0x9000+placement),(nox_object_t*)(raw+0xa000+placement)};
 owner->inv_first_item=items[q[8]];
 dword_5d4594_1049864=q[1];dword_5d4594_1062488=q[3];
 dword_5d4594_1062452=q[7]?(nox_window*)(raw+0xc000+placement):NULL;
}
static void inventory_state_layout_C(uintptr_t*v){
 v[0]=sizeof(void*);v[1]=sizeof(nox_drawable);v[2]=offsetof(nox_drawable,field_27);v[3]=offsetof(nox_drawable,flags28);v[4]=offsetof(nox_drawable,field_92);
 v[5]=sizeof(((nox_drawable*)0)->field_27);v[6]=sizeof(((nox_drawable*)0)->flags28);v[7]=sizeof(((nox_drawable*)0)->field_92);
 v[8]=sizeof(nox_object_t);v[9]=offsetof(nox_object_t,inv_first_item);v[10]=sizeof(((nox_object_t*)0)->inv_first_item);v[11]=sizeof(nox_window);
 v[12]=sizeof(array_5D4594_1049872);v[13]=sizeof(array_5D4594_1049872[0]);v[14]=sizeof(array_5D4594_1049872)/sizeof(array_5D4594_1049872[0]);
 v[15]=sizeof(dword_5d4594_1049864);v[16]=sizeof(dword_5d4594_1062488);v[17]=sizeof(dword_5d4594_1062452);v[18]=sizeof(int);v[19]=sizeof(uint32_t);
}
*/
import "C"

import "unsafe"

func inventoryStateNativeLayoutC() (out [20]uintptr) {
	C.inventory_state_layout_C((*C.uintptr_t)(unsafe.Pointer(&out[0])))
	return
}

func inventoryStateNativeSeedC(raw unsafe.Pointer, placement int, input *[54]uint32) {
	C.inventory_state_seed_C(raw, C.size_t(placement), (*C.uint32_t)(unsafe.Pointer(&input[0])))
}

func inventoryStateNativeCall(route, index int, owner unsafe.Pointer, query, x, y, value int32) (uint32, unsafe.Pointer) {
	o := (*C.nox_object_t)(owner)
	if route != 2 {
		r := C.inventory_state_call_C(C.uint(route), C.uint(index), o, C.int(query), C.int(x), C.int(y), C.int(value))
		return uint32(r.word), r.pointer
	}
	switch index {
	case 0:
		return 0, unsafe.Pointer(sub_461600(C.int(query)))
	case 1:
		return uint32(sub_461930()), nil
	case 2:
		return uint32(sub_464B40(C.int(x), C.int(y))), nil
	case 3:
		sub_467680()
		return 0, nil
	case 4:
		return 0, unsafe.Pointer(nox_xxx_wndGetHandle_4676A0())
	case 5:
		return uint32(sub_467740(C.int(value))), nil
	case 6:
		return 0, unsafe.Pointer(nox_xxx_inventoryGetFirst_4E7980(o))
	}
	panic("invalid inventory state entry")
}
