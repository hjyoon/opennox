package legacy

/*
#include "GAME1_1.h"
#include "private_player_records_native.h"
#include <stddef.h>
#include <stdlib.h>
_Static_assert(sizeof(nox_private_player_record_t)==(sizeof(void*)==8?80:52),"private parent size");
_Static_assert(sizeof(nox_private_player_child_t)==(sizeof(void*)==8?32:16),"private child size");
_Static_assert(offsetof(nox_private_player_record_t,name)==3*sizeof(void*),"private name offset");
_Static_assert(offsetof(nox_private_player_record_t,id)==(sizeof(void*)==8?44:32),"private ID offset");
_Static_assert(offsetof(nox_private_player_record_t,team)==(sizeof(void*)==8?48:36),"private team offset");
_Static_assert(offsetof(nox_private_player_record_t,children)==(sizeof(void*)==8?56:40),"private children offset");
_Static_assert(offsetof(nox_private_player_child_t,peer)==3*sizeof(void*),"private peer offset");
_Static_assert(_Generic(&sub_4259C0,void (*)(void):1,default:0),"private init ABI");
_Static_assert(_Generic(&sub_4259F0,int* (*)(void):1,default:0),"private free ABI");
_Static_assert(_Generic(&sub_425A50,int* (*)(void):1,default:0),"private first ABI");
_Static_assert(_Generic(&sub_425A60,int* (*)(int*):1,default:0),"private next ABI");
_Static_assert(_Generic(&sub_425A70,int* (*)(int):1,default:0),"private find ABI");
_Static_assert(_Generic(&sub_425AA0,int (*)(int):1,default:0),"private exists ABI");
_Static_assert(_Generic(&sub_425AD0,wchar2_t* (*)(int,wchar2_t*):1,default:0),"private create ABI");
_Static_assert(_Generic(&sub_425B30,void (*)(void*,int):1,default:0),"private add ABI");
_Static_assert(_Generic(&sub_425B60,char* (*)(void*,int):1,default:0),"private remove ABI");
_Static_assert(_Generic(&sub_425BC0,int* (*)(void*):1,default:0),"private first child ABI");
_Static_assert(_Generic(&sub_425BE0,int* (*)(int*):1,default:0),"private next child ABI");
static uintptr_t private_records_call_C(int mode,void* pointer,int32_t value,uint16_t* name) {
    switch(mode) {
    case 0:sub_4259C0();return 0;
    case 1:return (uintptr_t)sub_4259F0();
    case 2:return (uintptr_t)sub_425A50();
    case 3:return (uintptr_t)sub_425A60(pointer);
    case 4:return (uintptr_t)sub_425A70(value);
    case 5:return (uint32_t)sub_425AA0(value);
    case 6:return (uintptr_t)sub_425AD0(value,name);
    case 7:sub_425B30(pointer,value);return 0;
    case 8:return (uintptr_t)sub_425B60(pointer,value);
    case 9:return (uintptr_t)sub_425BC0(pointer);
    case 10:return (uintptr_t)sub_425BE0(pointer);
    default:abort();
    }
}
static void private_records_layout_C(uintptr_t* v) {
    v[0]=sizeof(nox_private_player_record_t);v[1]=sizeof(nox_private_player_child_t);
    v[2]=offsetof(nox_private_player_record_t,name);v[3]=offsetof(nox_private_player_record_t,id);
    v[4]=offsetof(nox_private_player_record_t,team);v[5]=offsetof(nox_private_player_record_t,children);
    v[6]=offsetof(nox_private_player_child_t,peer);
}
static void private_records_consumer_C(void* record,void* team,uintptr_t* out) {
    nox_private_player_record_t* p=record;
    p->team=team;out[0]=(uintptr_t)p->name;out[1]=(uint32_t)p->id;out[2]=(uintptr_t)p->team;
}
*/
import "C"
import "unsafe"

func privatePlayerRecordsCallC(mode int, p unsafe.Pointer, value int32, name *uint16) uintptr {
	return uintptr(C.private_records_call_C(C.int(mode), p, C.int32_t(value), (*C.uint16_t)(unsafe.Pointer(name))))
}

func privatePlayerRecordsLayoutC() (out [7]uintptr) {
	C.private_records_layout_C((*C.uintptr_t)(unsafe.Pointer(&out[0])))
	return out
}

func privatePlayerRecordsConsumerC(p, team unsafe.Pointer) (out [3]uintptr) {
	C.private_records_consumer_C(p, team, (*C.uintptr_t)(unsafe.Pointer(&out[0])))
	return out
}
