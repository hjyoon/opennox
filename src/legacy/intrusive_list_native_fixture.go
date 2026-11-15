package legacy

/*
#include "GAME1_1.h"
#include <stddef.h>
#include <stdlib.h>
_Static_assert(_Generic(&nox_common_list_clear_425760,void (*)(nox_list_item_t*):1,default:0),"list clear ABI");
_Static_assert(_Generic(&sub_425770,void* (*)(void*):1,default:0),"list init ABI");
_Static_assert(_Generic(&sub_425790,int (*)(int*,uint32_t*):1,default:0),"list ascending ABI");
_Static_assert(_Generic(&sub_4257F0,void (*)(int*,uint32_t*):1,default:0),"list descending ABI");
_Static_assert(_Generic(&nox_common_list_getFirstSafe_425890,nox_list_item_t* (*)(nox_list_item_t*):1,default:0),"list first ABI");
_Static_assert(_Generic(&nox_common_list_getNextSafe_4258A0,nox_list_item_t* (*)(nox_list_item_t*):1,default:0),"list safe next ABI");
_Static_assert(_Generic(&sub_4258C0,uint32_t* (*)(uint32_t**,int):1,default:0),"list index pointer ABI");
_Static_assert(_Generic(&nox_common_list_append_4258E0,void (*)(nox_list_item_t*,nox_list_item_t*):1,default:0),"list before ABI");
_Static_assert(_Generic(&sub_425900,uint32_t* (*)(uint32_t*,uint32_t*):1,default:0),"list after ABI");
_Static_assert(_Generic(&nox_common_list_remove_425920,void (*)(void*):1,default:0),"list remove ABI");
_Static_assert(_Generic(&nox_common_list_getNext_425940,nox_list_item_t* (*)(nox_list_item_t*):1,default:0),"list next ABI");
_Static_assert(_Generic(&sub_425960,nox_list_item_t* (*)(nox_list_item_t*):1,default:0),"native reverse pointer ABI");
static uintptr_t intrusive_list_call_C(int mode,void* a,void* b,int32_t index) {
    switch(mode) {
    case 0:nox_common_list_clear_425760(a);return 0;
    case 1:return (uintptr_t)sub_425770(a);
    case 2:return (uint32_t)sub_425790((int*)a,(uint32_t*)b);
    case 3:sub_4257F0((int*)a,(uint32_t*)b);return 0;
    case 4:return (uintptr_t)nox_common_list_getFirstSafe_425890(a);
    case 5:return (uintptr_t)nox_common_list_getNextSafe_4258A0(a);
    case 6:return (uintptr_t)sub_4258C0((uint32_t**)a,index);
    case 7:nox_common_list_append_4258E0(a,b);return 0;
    case 8:return (uintptr_t)sub_425900((uint32_t*)a,(uint32_t*)b);
    case 9:nox_common_list_remove_425920(a);return 0;
    case 10:return (uintptr_t)nox_common_list_getNext_425940(a);
    case 11:return (uintptr_t)sub_425960(a);
    default:abort();
    }
}
static void intrusive_list_layout_C(uintptr_t* values) {
    values[0]=sizeof(nox_list_item_t);
    values[1]=offsetof(nox_list_item_t,field_0);
    values[2]=offsetof(nox_list_item_t,field_1);
    values[3]=offsetof(nox_list_item_t,field_2);
    values[4]=sizeof(void*);
}
*/
import "C"
import "unsafe"

func intrusiveListCallC(mode int, a, b unsafe.Pointer, index int32) uintptr {
	return uintptr(C.intrusive_list_call_C(C.int(mode), a, b, C.int32_t(index)))
}

func intrusiveListLayoutC() (out [5]uintptr) {
	C.intrusive_list_layout_C((*C.uintptr_t)(unsafe.Pointer(&out[0])))
	return out
}
