package cfixture

/*
#include "GAME1_1.h"
#include "memmap.h"
extern uint32_t dword_587000_60044;
typedef const char global_record_const_char_t;
typedef char*(*global_record_get_C)(void);
typedef void(*global_record_set_C)(const char*);
typedef int(*global_record_out_C)(const char**,const char**);
typedef int(*global_record_word_C)(void);
_Static_assert(_Generic(&sub_41FA40,global_record_get_C:1,default:0),"global string pointer ABI");
_Static_assert(_Generic(&sub_41FA50,global_record_set_C:1,default:0),"nullable const string ABI");
_Static_assert(_Generic(&sub_41FBE0,global_record_out_C:1,default:0),"native const char pointer output ABI");
_Static_assert(_Generic(&sub_4200E0,global_record_word_C:1,default:0),"state signed DWORD ABI");
_Static_assert(_Generic(&sub_420100,global_record_word_C:1,default:0),"state shifted signed DWORD ABI");
static global_record_get_C volatile global_record_get_callback_C=sub_41FA40;
static global_record_set_C volatile global_record_set_callback_C=sub_41FA50;
static global_record_out_C volatile global_record_out_callback_C=sub_41FBE0;
static global_record_word_C volatile global_record_word_callbacks_C[]={sub_4200E0,sub_420100};
static uintptr_t global_record_call_C(unsigned route,unsigned entry,const char*source,const char**first,const char**second){
 if(route){switch(entry){case 0:{global_record_get_C f=global_record_get_callback_C;return(uintptr_t)f();}case 1:{global_record_set_C f=global_record_set_callback_C;f(source);return 0;}case 2:{global_record_out_C f=global_record_out_callback_C;return(uint32_t)f(first,second);}case 3:case 4:{global_record_word_C f=global_record_word_callbacks_C[entry-3];return(uint32_t)f();}default:abort();}}
 switch(entry){case 0:return(uintptr_t)sub_41FA40();case 1:sub_41FA50(source);return 0;case 2:return(uint32_t)sub_41FBE0(first,second);case 3:return(uint32_t)sub_4200E0();case 4:return(uint32_t)sub_420100();default:abort();}
}
static void global_record_layout_C(uintptr_t*v){
 v[0]=sizeof(void*);v[1]=sizeof(int);v[2]=sizeof(uint32_t);v[3]=sizeof(char);v[4]=sizeof(dword_587000_60044);v[5]=sizeof(global_record_get_C);v[6]=sizeof(global_record_set_C);v[7]=sizeof(global_record_out_C);v[8]=sizeof(global_record_word_C);v[9]=sizeof(const char*);v[10]=_Alignof(const char*);v[11]=24;v[12]=10;
}
static void global_record_owner_pointers_C(uintptr_t*v){v[0]=(uintptr_t)mem_getPtr(0x5d4594,0x650000-0x5d4594);v[1]=(uintptr_t)mem_getPtr(0x5d4594,531660);v[2]=(uintptr_t)mem_getPtr(0x5d4594,534756);v[3]=(uintptr_t)mem_getU32Ptr(0x587000,60072);v[4]=(uintptr_t)&dword_587000_60044;}
static uint32_t global_record_selected_C(int set,uint32_t v){if(set)dword_587000_60044=v;return dword_587000_60044;}
static uint32_t global_record_state_C(int set,uint32_t v){uint32_t*p=mem_getU32Ptr(0x587000,60072);if(set)*p=v;return*p;}
// Bounded copy of the actual unchanged caller. The owner service returns NULL;
// other owner/flag services must not run. Only its real typed output-slot call
// and 108-byte scratch clear receive coverage; full caller migration is uncredited.
static unsigned char*global_record_parent_scratch_C;
static uintptr_t global_record_parent_pointers_C[2];
static unsigned global_record_parent_owner_calls_C;
static int global_record_parent_output_C(const char**first,const char**second){int result=sub_41FBE0(first,second);if(result){global_record_parent_pointers_C[0]=(uintptr_t)*first;global_record_parent_pointers_C[1]=(uintptr_t)*second;}return result;}
static void*global_record_parent_mem_C(uintptr_t base,uintptr_t off){if(base!=0x85b3fc||off!=10308||!global_record_parent_scratch_C)abort();return global_record_parent_scratch_C;}
static int global_record_parent_owner_C(void){global_record_parent_owner_calls_C++;return 0;}
static void global_record_parent_unexpected_flag_C(unsigned flags){(void)flags;abort();}
#undef getMemAt
#define getMemAt global_record_parent_mem_C
#define nox_xxx_officialStringCmp_41FDE0 global_record_parent_copy_C
#define sub_41FBE0 global_record_parent_output_C
#define sub_4207E0 global_record_parent_owner_C
#define nox_xxx_setGameFlags_40A4D0 global_record_parent_unexpected_flag_C
#define nox_common_gameFlags_unset_40A540 global_record_parent_unexpected_flag_C
static int nox_xxx_officialStringCmp_41FDE0() {
	int v0;         // ebx
	size_t v1;      // eax
	const char* v3; // [esp+8h] [ebp-8h]
	const char* v4; // [esp+Ch] [ebp-4h]

	memset(getMemAt(0x85B3FC, 10308), 0, 0x6Cu);
	if (sub_41FBE0(&v3, &v4) != 1) {
		return 0;
	}
	v0 = sub_4207E0();
	if (v0) {
		strcpy((char*)(v0 + 228), v3);
		strcpy((char*)(v0 + 238), v4);
		sub_41FA50(v3);
		v1 = strcspn((const char*)(v0 + 24), ":");
		if (!strncmp("Official", (const char*)(v1 + v0 + 25), 8u)) {
			nox_xxx_setGameFlags_40A4D0(0x1000000);
		} else {
			nox_common_gameFlags_unset_40A540(0x1000000);
		}
	}
	return 1;
}
#undef getMemAt
#define getMemAt mem_getPtr
#undef nox_xxx_officialStringCmp_41FDE0
#undef sub_41FBE0
#undef sub_4207E0
#undef nox_xxx_setGameFlags_40A4D0
#undef nox_common_gameFlags_unset_40A540
static void global_record_parent_call_C(void*scratch,uintptr_t*v){global_record_parent_scratch_C=scratch;global_record_parent_pointers_C[0]=global_record_parent_pointers_C[1]=0;global_record_parent_owner_calls_C=0;v[0]=(uint32_t)global_record_parent_copy_C();v[1]=global_record_parent_pointers_C[0];v[2]=global_record_parent_pointers_C[1];v[3]=global_record_parent_owner_calls_C;global_record_parent_scratch_C=NULL;}
*/
import "C"

import "unsafe"

func GlobalRecordNativeLayoutC() (out [13]uintptr) {
	C.global_record_layout_C((*C.uintptr_t)(unsafe.Pointer(&out[0])))
	return
}

func GlobalRecordNativeOwnerPointersC() (out [5]uintptr) {
	C.global_record_owner_pointers_C((*C.uintptr_t)(unsafe.Pointer(&out[0])))
	return
}

func GlobalRecordNativeSelectedC(set bool, value uint32) uint32 {
	var mode C.int
	if set {
		mode = 1
	}
	return uint32(C.global_record_selected_C(mode, C.uint32_t(value)))
}

func GlobalRecordNativeStateC(set bool, value uint32) uint32 {
	var mode C.int
	if set {
		mode = 1
	}
	return uint32(C.global_record_state_C(mode, C.uint32_t(value)))
}

func GlobalRecordNativeCall(route, entry int, source, first, second unsafe.Pointer) uintptr {
	return uintptr(C.global_record_call_C(C.uint(route), C.uint(entry), (*C.char)(source),
		(**C.global_record_const_char_t)(first), (**C.global_record_const_char_t)(second)))
}

func GlobalRecordNativeParentC(scratch unsafe.Pointer) (out [4]uintptr) {
	C.global_record_parent_call_C(scratch, (*C.uintptr_t)(unsafe.Pointer(&out[0])))
	return
}
