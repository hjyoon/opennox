package legacy

/*
#include "defs.h"
#include "memmap.h"
extern unsigned int dword_5d4594_1599628;
int sub_512E80(wchar2_t*);
typedef int(*wide_script_pointer_entry_C)(wchar2_t*);
_Static_assert(_Generic(&sub_512E80,wide_script_pointer_entry_C:1,default:0),"wide script pointer entry ABI");
static wide_script_pointer_entry_C volatile wide_script_pointer_callback_C=sub_512E80;
static int wide_script_pointer_call_C(int route,wchar2_t*value){
 if(route){wide_script_pointer_entry_C f=wide_script_pointer_callback_C;return f(value);}
 return sub_512E80(value);
}
static void wide_script_pointer_layout_C(uintptr_t*v){
 v[0]=sizeof(void*);v[1]=sizeof(int);v[2]=sizeof(wchar2_t);v[3]=sizeof(dword_5d4594_1599628);v[4]=sizeof(wide_script_pointer_entry_C);v[5]=sizeof(uint32_t);v[6]=_Alignof(wchar2_t);v[7]=1024;v[8]=4;v[9]=sizeof(uintptr_t);
}
static void wide_script_pointer_owners_C(uintptr_t*v){v[0]=(uintptr_t)&dword_5d4594_1599628;v[1]=(uintptr_t)getMemAt(0x973f18,0);v[2]=(uintptr_t)getMemAt(0x5d4594,1599628);}
static uint32_t wide_script_pointer_counter_C(int set,uint32_t value){if(set)dword_5d4594_1599628=value;return dword_5d4594_1599628;}
static uintptr_t wide_script_pointer_read_C(unsigned cell){return(uintptr_t)getMemPtr(0x973f18,26664+4*cell);}
*/
import "C"

import "unsafe"

func wideScriptPointerNativeLayoutC() (out [10]uintptr) {
	C.wide_script_pointer_layout_C((*C.uintptr_t)(unsafe.Pointer(&out[0])))
	return
}

func wideScriptPointerNativeOwnersC() (out [3]uintptr) {
	C.wide_script_pointer_owners_C((*C.uintptr_t)(unsafe.Pointer(&out[0])))
	return
}

func wideScriptPointerNativeCounterC(set bool, value uint32) uint32 {
	var mode C.int
	if set {
		mode = 1
	}
	return uint32(C.wide_script_pointer_counter_C(mode, C.uint32_t(value)))
}

func wideScriptPointerNativeCall(route int, value unsafe.Pointer) uint32 {
	if route == 2 {
		return uint32(sub_512E80((*C.wchar2_t)(value)))
	}
	return uint32(C.wide_script_pointer_call_C(C.int(route), (*C.wchar2_t)(value)))
}

func wideScriptPointerNativeReadC(cell int) uintptr {
	return uintptr(C.wide_script_pointer_read_C(C.uint(cell)))
}
