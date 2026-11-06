package legacy

/*
#include <stddef.h>
#include <stdint.h>
#include "regeneration_raw_4e01d0.h"
#ifdef _WIN32
#include <windows.h>
static void* regeneration_raw_heap_4e01d0(size_t* guard) {
 SYSTEM_INFO info;GetSystemInfo(&info);size_t page=info.dwPageSize;*guard=page;
 unsigned char* base=VirtualAlloc(NULL,0x10000+2*page,MEM_RESERVE|MEM_COMMIT|MEM_TOP_DOWN,PAGE_READWRITE);
 if(!base)return NULL;
 DWORD old;
 if(!VirtualProtect(base,page,PAGE_NOACCESS,&old)||!VirtualProtect(base+page+0x10000,page,PAGE_NOACCESS,&old)){VirtualFree(base,0,MEM_RELEASE);return NULL;}
 return base+page;
}
static void regeneration_raw_free_4e01d0(void* heap,size_t guard) {VirtualFree((unsigned char*)heap-guard,0,MEM_RELEASE);}
#else
#include <sys/mman.h>
#include <unistd.h>
static void* regeneration_raw_heap_4e01d0(size_t* guard) {
 size_t page=(size_t)sysconf(_SC_PAGESIZE);*guard=page;
 unsigned char* base=mmap(NULL,0x10000+2*page,PROT_READ|PROT_WRITE,MAP_PRIVATE|MAP_ANONYMOUS,-1,0);
 if(base==MAP_FAILED)return NULL;
 if(mprotect(base,page,PROT_NONE)||mprotect(base+page+0x10000,page,PROT_NONE)){munmap(base,0x10000+2*page);return NULL;}
 return base+page;
}
static void regeneration_raw_free_4e01d0(void* heap,size_t guard) {munmap((unsigned char*)heap-guard,0x10000+2*guard);}
#endif
static void* regeneration_raw_bad_4e01d0(void* heap,size_t guard) {return (void*)((uintptr_t)heap-guard+64);}

typedef void (*regeneration_raw_old_4e01d0)(int,int);
typedef void (*regeneration_raw_native_4e01d0)(void*,nox_object_t*);
static void regeneration_raw_old_call_4e01d0(regeneration_raw_old_4e01d0 fn,void* effect,nox_object_t* item) {
 fn((int)(uintptr_t)effect,(int)(uintptr_t)item);
}
static void regeneration_raw_native_call_4e01d0(regeneration_raw_native_4e01d0 fn,void* effect,nox_object_t* item) {
 fn(effect,item);
}
static void regeneration_raw_public_4e01d0(void* effect,nox_object_t* item) {
 _Generic(&nox_xxx_effectRegeneration_4E01D0,
  regeneration_raw_old_4e01d0:regeneration_raw_old_call_4e01d0,
  regeneration_raw_native_4e01d0:regeneration_raw_native_call_4e01d0)
 (&nox_xxx_effectRegeneration_4E01D0,effect,item);
}
static int regeneration_raw_abi_4e01d0(void) {
 return _Generic(&nox_xxx_effectRegeneration_4E01D0,regeneration_raw_old_4e01d0:1,regeneration_raw_native_4e01d0:2);
}
static uintptr_t regeneration_raw_layout_4e01d0(int index) {
 switch(index) {
 case 0:return offsetof(nox_object_t,inv_holder);
 case 1:return offsetof(nox_object_t,obj_class);
 case 2:return offsetof(nox_object_t,health_data);
 case 3:return offsetof(nox_object_t,field_134);
 case 4:return offsetof(nox_object_t,obj_flags);
 case 5:return offsetof(obj_412ae0_t,field_27);
 case 6:return 0; // exact ushort current-HP field in the C getter contract
 case 7:return 4; // exact ushort maximum-HP field in the C getter contract
 case 8:return sizeof(nox_object_t);
 case 9:return sizeof(obj_412ae0_t);
 default:return sizeof(void*);
 }
}
// Only call these field getters for independently readable native domains.
static uintptr_t regeneration_raw_initial_read_4e01d0(void* base,int kind) {
 switch(kind) {
 case 0:return (uintptr_t)((nox_object_t*)base)->inv_holder;
 case 1:return ((nox_object_t*)base)->obj_class;
 case 2:return (uintptr_t)((nox_object_t*)base)->health_data;
 case 3:return ((nox_object_t*)base)->field_134;
 case 4:return ((nox_object_t*)base)->obj_flags;
 case 5:return (uint32_t)((obj_412ae0_t*)base)->field_27;
 case 6:return ((uint16_t*)base)[0];
 default:return ((uint16_t*)base)[2];
 }
}
*/
import "C"

import "unsafe"

func regenerationRawHeap4E01D0() (unsafe.Pointer, uintptr, func()) {
	var guard C.size_t
	heap := C.regeneration_raw_heap_4e01d0(&guard)
	return heap, uintptr(guard), func() {
		if heap != nil {
			C.regeneration_raw_free_4e01d0(heap, guard)
		}
	}
}

func regenerationRawBad4E01D0(heap unsafe.Pointer, guard uintptr) unsafe.Pointer {
	return C.regeneration_raw_bad_4e01d0(heap, C.size_t(guard))
}

func regenerationRawLayout4E01D0() (out [11]uintptr) {
	for i := range out {
		out[i] = uintptr(C.regeneration_raw_layout_4e01d0(C.int(i)))
	}
	return
}

func regenerationRawABI4E01D0() int { return int(C.regeneration_raw_abi_4e01d0()) }

func regenerationRawInitialC4E01D0(base unsafe.Pointer, kind int) uintptr {
	return uintptr(C.regeneration_raw_initial_read_4e01d0(base, C.int(kind)))
}

func regenerationRawPublic4E01D0(effect, item unsafe.Pointer) {
	C.regeneration_raw_public_4e01d0(effect, (*C.nox_object_t)(item))
}
