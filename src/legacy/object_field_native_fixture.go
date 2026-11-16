package legacy

/*
#include "GAME3_3.h"
#include "GAME4_3.h"
#include "GAME5_2.h"
typedef struct object_field_action_C {uint32_t action;uintptr_t args[4];uint32_t field5;} object_field_action_C;
typedef struct object_field_update_C {
 uint8_t fixed300[300];void* waypoints[16];uint32_t fields91_102[12];uint16_t health103[32];
 void* anim119;uint8_t animation[4];void* definition;void* soundset;uint32_t fields123_135[13];
 int8_t index;uint8_t byte136_1;uint16_t word136_2;uint32_t field137;object_field_action_C stack[24];
} object_field_update_C;
typedef int(*object_field_AI_callback_C)(nox_object_t*);
typedef void(*object_field_sync_callback_C)(nox_object_t*);
typedef int*(*object_field_mask_callback_C)(nox_object_t*,int,int,int);
typedef uint32_t(*object_field_buffs_callback_C)(nox_object_t*);
typedef nox_object_t*(*object_field_armor_callback_C)(nox_object_t*,nox_object_t*);
typedef unsigned int(*object_field_code_callback_C)(nox_drawable*);
typedef int(*object_field_clear_callback_C)(uint32_t);
typedef unsigned int(*object_field_bit_callback_C)(unsigned int);
_Static_assert(_Generic(&sub_5408A0,object_field_AI_callback_C:1,default:0),"native AI ABI");
_Static_assert(_Generic(&nox_xxx_unitNeedSync_4E44F0,object_field_sync_callback_C:1,default:0),"native sync ABI");
_Static_assert(_Generic(&sub_4E4500,object_field_mask_callback_C:1,default:0),"native field mask ABI");
_Static_assert(_Generic(&nox_xxx_objectGetBuffs_4E4A80,object_field_buffs_callback_C:1,default:0),"native buffs ABI");
_Static_assert(_Generic(&nox_xxx_armorHaveSameSubclass_53E7B0,object_field_armor_callback_C:1,default:0),"native armor ABI");
_Static_assert(_Generic(&nox_xxx_netGetUnitCodeCli_578B00,object_field_code_callback_C:1,default:0),"native Drawable ABI");
_Static_assert(_Generic(&nox_xxx_netClearHighBit_578B30,object_field_clear_callback_C:1,default:0),"full32 code argument ABI");
_Static_assert(_Generic(&nox_xxx_netTestHighBit_578B70,object_field_bit_callback_C:1,default:0),"full32 bit test ABI");
static object_field_AI_callback_C volatile object_field_AI_C=sub_5408A0;
static object_field_sync_callback_C volatile object_field_sync_C=nox_xxx_unitNeedSync_4E44F0;
static object_field_mask_callback_C volatile object_field_mask_C=sub_4E4500;
static object_field_buffs_callback_C volatile object_field_buffs_C=nox_xxx_objectGetBuffs_4E4A80;
static object_field_armor_callback_C volatile object_field_armor_C=nox_xxx_armorHaveSameSubclass_53E7B0;
static object_field_code_callback_C volatile object_field_code_C=nox_xxx_netGetUnitCodeCli_578B00;
static object_field_clear_callback_C volatile object_field_clear_C=nox_xxx_netClearHighBit_578B30;
static object_field_bit_callback_C volatile object_field_bit_C=nox_xxx_netTestHighBit_578B70;
typedef struct object_field_result_C {uint32_t word;void*pointer;}object_field_result_C;
static object_field_result_C object_field_call_C(unsigned route,unsigned i,nox_object_t*owner,nox_object_t*item,nox_drawable*drawable,uint32_t a,uint32_t b,uint32_t mode,uint32_t word){
 object_field_result_C result={0};
 if(route==1){
  switch(i){
  case 0:{object_field_AI_callback_C f=object_field_AI_C;result.word=(uint32_t)f(owner);break;}
  case 1:{object_field_sync_callback_C f=object_field_sync_C;f(owner);break;}
  case 2:{object_field_mask_callback_C f=object_field_mask_C;result.pointer=f(owner,(int)a,(int)b,(int)mode);break;}
  case 3:{object_field_buffs_callback_C f=object_field_buffs_C;result.word=f(owner);break;}
  case 4:{object_field_armor_callback_C f=object_field_armor_C;result.pointer=f(owner,item);break;}
  case 5:{object_field_code_callback_C f=object_field_code_C;result.word=f(drawable);break;}
  case 6:{object_field_clear_callback_C f=object_field_clear_C;result.word=(uint32_t)f(word);break;}
  case 7:{object_field_bit_callback_C f=object_field_bit_C;result.word=f(word);break;}
  default:abort();
  }
  return result;
 }
 switch(i){
 case 0:result.word=(uint32_t)sub_5408A0(owner);break;
 case 1:nox_xxx_unitNeedSync_4E44F0(owner);break;
 case 2:result.pointer=sub_4E4500(owner,(int)a,(int)b,(int)mode);break;
 case 3:result.word=nox_xxx_objectGetBuffs_4E4A80(owner);break;
 case 4:result.pointer=nox_xxx_armorHaveSameSubclass_53E7B0(owner,item);break;
 case 5:result.word=nox_xxx_netGetUnitCodeCli_578B00(drawable);break;
 case 6:result.word=(uint32_t)nox_xxx_netClearHighBit_578B30(word);break;
 case 7:result.word=nox_xxx_netTestHighBit_578B70(word);break;
 default:abort();
 }
 return result;
}
static void object_field_seed_C(uint8_t*arena,uintptr_t placement,const uint32_t*x){
 nox_object_t*owner=(nox_object_t*)(arena+0x1000+placement),*item=(nox_object_t*)(arena+0x5000+placement),*nodes[4];
 object_field_update_C*update=(object_field_update_C*)(arena+0x3000+placement);
 nox_drawable*drawable=(nox_drawable*)(arena+0xa000+placement);
 owner->data_update=update;owner->field_38=x[1];owner->buffs=x[1];owner->field_37=x[7];
 for(unsigned j=0;j<32;j++)owner->field_140[j]=x[12+j];
 update->index=(int8_t)x[2];
 if(update->index==-1)update->fields123_135[9]=x[3];
 else update->stack[(unsigned)update->index].action=x[3];
 item->obj_subclass=x[10];
 for(unsigned j=0;j<4;j++)nodes[j]=(nox_object_t*)(arena+0x6000+j*0x1000+placement);
 owner->inv_first_item=x[11]?nodes[0]:NULL;
 for(unsigned j=0;j<x[11];j++){
  nodes[j]->obj_flags=x[44+3*j];nodes[j]->obj_class=x[45+3*j];nodes[j]->obj_subclass=x[46+3*j];
  nodes[j]->inv_next_item=j+1<x[11]?nodes[j+1]:NULL;
 }
 drawable->flags28=x[7];drawable->field_32=x[8];
}
static void object_field_layout_C(uintptr_t*v){
 v[0]=sizeof(void*);v[1]=sizeof(nox_object_t);v[2]=offsetof(nox_object_t,data_update);
 v[3]=offsetof(nox_object_t,field_38);v[4]=offsetof(nox_object_t,field_37);v[5]=offsetof(nox_object_t,field_140);v[6]=sizeof(((nox_object_t*)0)->field_140);
 v[7]=offsetof(nox_object_t,buffs);v[8]=offsetof(nox_object_t,obj_class);v[9]=offsetof(nox_object_t,obj_subclass);v[10]=offsetof(nox_object_t,obj_flags);
 v[11]=offsetof(nox_object_t,inv_first_item);v[12]=offsetof(nox_object_t,inv_next_item);v[13]=sizeof(((nox_object_t*)0)->field_37);
 v[14]=offsetof(object_field_update_C,index);v[15]=offsetof(object_field_update_C,stack);v[16]=sizeof(object_field_action_C);
 v[17]=offsetof(object_field_action_C,action);v[18]=offsetof(object_field_action_C,args);v[19]=offsetof(object_field_action_C,field5);
 v[20]=offsetof(object_field_update_C,fields123_135)+9*4;v[21]=sizeof(((object_field_update_C*)0)->index);v[22]=sizeof(((object_field_update_C*)0)->stack);
 v[23]=offsetof(nox_drawable,flags28);v[24]=offsetof(nox_drawable,field_32);v[25]=sizeof(((nox_drawable*)0)->flags28);v[26]=sizeof(((nox_drawable*)0)->field_32);
 v[27]=sizeof(int);v[28]=sizeof(short);v[29]=sizeof(uint32_t);v[30]=sizeof(((nox_object_t*)0)->buffs);
 v[31]=sizeof(((nox_object_t*)0)->inv_first_item);v[32]=sizeof(((nox_object_t*)0)->data_update);v[33]=sizeof(((object_field_action_C*)0)->args[0]);v[34]=sizeof(int*);
}
*/
import "C"

import "unsafe"

func objectFieldNativeLayoutC() (out [35]uintptr) {
	C.object_field_layout_C((*C.uintptr_t)(unsafe.Pointer(&out[0])))
	return
}

func objectFieldNativeSeedC(raw unsafe.Pointer, placement int, words *[56]uint32) {
	// C owns every pointer slot in these foreign canary records.
	C.object_field_seed_C((*C.uint8_t)(raw), C.uintptr_t(placement), (*C.uint32_t)(unsafe.Pointer(&words[0])))
}

func objectFieldNativeCall(route, index int, owner, item, drawable unsafe.Pointer, value, mask, mode, word uint32) (uint32, unsafe.Pointer) {
	o, it, dr := (*C.nox_object_t)(owner), (*C.nox_object_t)(item), (*C.nox_drawable)(drawable)
	if route != 2 {
		result := C.object_field_call_C(C.uint(route), C.uint(index), o, it, dr, C.uint32_t(value), C.uint32_t(mask), C.uint32_t(mode), C.uint32_t(word))
		return uint32(result.word), result.pointer
	}
	switch index {
	case 0:
		return uint32(sub_5408A0(o)), nil
	case 1:
		nox_xxx_unitNeedSync_4E44F0(o)
		return 0, nil
	case 2:
		return 0, unsafe.Pointer(sub_4E4500(o, C.int(value), C.int(mask), C.int(mode)))
	case 3:
		return uint32(nox_xxx_objectGetBuffs_4E4A80(o)), nil
	case 4:
		return 0, unsafe.Pointer(nox_xxx_armorHaveSameSubclass_53E7B0(o, it))
	case 5:
		return uint32(nox_xxx_netGetUnitCodeCli_578B00(dr)), nil
	case 6:
		return uint32(nox_xxx_netClearHighBit_578B30(C.uint(word))), nil
	case 7:
		return uint32(nox_xxx_netTestHighBit_578B70(C.uint(word))), nil
	}
	panic("invalid object field entry")
}
