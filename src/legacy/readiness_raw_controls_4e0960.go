package legacy

/*
#include <stddef.h>
#include <stdint.h>
#include "readiness_raw_4e0960.h"
void nullsub_22(void);
#ifdef _WIN32
#include <windows.h>
static void* readiness_raw_heap_4e0960(size_t* guard) {
 SYSTEM_INFO info;GetSystemInfo(&info);size_t page=info.dwPageSize;*guard=page;
 unsigned char* base=VirtualAlloc(NULL,0x10000+2*page,MEM_RESERVE|MEM_COMMIT|MEM_TOP_DOWN,PAGE_READWRITE);
 if(!base)return NULL;
 DWORD old;
 if(!VirtualProtect(base,page,PAGE_NOACCESS,&old)||!VirtualProtect(base+page+0x10000,page,PAGE_NOACCESS,&old)){VirtualFree(base,0,MEM_RELEASE);return NULL;}
 return base+page;
}
static void readiness_raw_free_4e0960(void* heap,size_t guard) {VirtualFree((unsigned char*)heap-guard,0,MEM_RELEASE);}
#else
#include <sys/mman.h>
#include <unistd.h>
static void* readiness_raw_heap_4e0960(size_t* guard) {
 size_t page=(size_t)sysconf(_SC_PAGESIZE);*guard=page;
 unsigned char* base=mmap(NULL,0x10000+2*page,PROT_READ|PROT_WRITE,MAP_PRIVATE|MAP_ANONYMOUS,-1,0);
 if(base==MAP_FAILED)return NULL;
 if(mprotect(base,page,PROT_NONE)||mprotect(base+page+0x10000,page,PROT_NONE)){munmap(base,0x10000+2*page);return NULL;}
 return base+page;
}
static void readiness_raw_free_4e0960(void* heap,size_t guard) {munmap((unsigned char*)heap-guard,0x10000+2*guard);}
#endif
static void* readiness_raw_bad_4e0960(void* heap,size_t guard) {return (void*)((uintptr_t)heap-guard+64);}
static void* readiness_raw_callback_4e0960(void) {return (void*)nullsub_22;}

// An opaque comparison token, never dereferenced/invoked. On64-bit hosts it
// deliberately aliases the actual callback's low DWORD while changing only
// the high DWORD. Creation remains in C: no Go uintptr->pointer round trip.
static void* readiness_raw_callback_collision_4e0960(void) {
#if UINTPTR_MAX > UINT32_MAX
 const uintptr_t real=(uintptr_t)nullsub_22;
 const uintptr_t step=(uintptr_t)UINT64_C(0x100000000);
 return (void*)(real<=UINTPTR_MAX-step?real+step:real-step);
#else
 return NULL;
#endif
}
static int readiness_raw_callback_truncation_control_4e0960(void) {
#if UINTPTR_MAX > UINT32_MAX
 const uintptr_t real=(uintptr_t)nullsub_22;
 const uintptr_t token=(uintptr_t)readiness_raw_callback_collision_4e0960();
 return real!=token && (uint32_t)real==(uint32_t)token;
#else
 return 0;
#endif
}

typedef int (*readiness_raw_old_4e0960)(int);
typedef int32_t (*readiness_raw_native_4e0960)(nox_object_t*);
static int32_t readiness_raw_old_call_4e0960(readiness_raw_old_4e0960 fn,nox_object_t* item) {
 return fn((int)(uintptr_t)item);
}
static int32_t readiness_raw_native_call_4e0960(readiness_raw_native_4e0960 fn,nox_object_t* item) {
 return fn(item);
}
static int32_t readiness_raw_public_4e0960(nox_object_t* item) {
 return _Generic(&nox_xxx_itemCheckReadinessEffect_4E0960,
  readiness_raw_old_4e0960:readiness_raw_old_call_4e0960,
  readiness_raw_native_4e0960:readiness_raw_native_call_4e0960)
 (&nox_xxx_itemCheckReadinessEffect_4E0960,item);
}
static int readiness_raw_abi_4e0960(void) {
 return _Generic(&nox_xxx_itemCheckReadinessEffect_4E0960,readiness_raw_old_4e0960:1,readiness_raw_native_4e0960:2);
}
static uintptr_t readiness_raw_layout_4e0960(int index) {
 switch(index) {
 case 0:return offsetof(nox_object_t,init_data);
 case 1:return offsetof(nox_object_t,obj_class);
 case 2:return offsetof(nox_modifier_attrs_t,modifiers[0]);
 case 3:return offsetof(nox_modifier_attrs_t,modifiers[1]);
 case 4:return offsetof(nox_modifier_attrs_t,modifiers[2]);
 case 5:return offsetof(nox_modifier_attrs_t,modifiers[3]);
 case 6:return offsetof(obj_412ae0_t,field_10);
 case 7:return offsetof(obj_412ae0_t,field_12);
 case 8:return sizeof(nox_object_t);
 case 9:return sizeof(nox_modifier_attrs_t);
 case 10:return sizeof(obj_412ae0_t);
 case 11:return sizeof(void*);
 case 12:return _Alignof(nox_object_t);
 case 13:return _Alignof(nox_modifier_attrs_t);
 default:return _Alignof(obj_412ae0_t);
 }
}
// These C getters are used ONLY in independently readable native domains.
// Callback values are observed as pointers, never invoked or truncated.
static uintptr_t readiness_raw_initial_read_4e0960(void* base,int kind,int index) {
 switch(kind) {
 case 0:return (uintptr_t)((nox_object_t*)base)->init_data;
 case 1:return ((nox_object_t*)base)->obj_class;
 case 2:return (uintptr_t)((nox_modifier_attrs_t*)base)->modifiers[index];
 case 3:return (uintptr_t)((obj_412ae0_t*)base)->field_10;
 default:return (uint32_t)((obj_412ae0_t*)base)->field_12;
 }
}
*/
import "C"

import "unsafe"

func readinessRawHeap4E0960() (unsafe.Pointer, uintptr, func()) {
	var guard C.size_t
	heap := C.readiness_raw_heap_4e0960(&guard)
	return heap, uintptr(guard), func() {
		if heap != nil {
			C.readiness_raw_free_4e0960(heap, guard)
		}
	}
}

func readinessRawBad4E0960(heap unsafe.Pointer, guard uintptr) unsafe.Pointer {
	return C.readiness_raw_bad_4e0960(heap, C.size_t(guard))
}

func readinessRawCallback4E0960() unsafe.Pointer { return C.readiness_raw_callback_4e0960() }

func readinessRawCollisionCallback4E0960() unsafe.Pointer {
	return C.readiness_raw_callback_collision_4e0960()
}

func readinessRawTruncationNegativeControl4E0960() bool {
	return C.readiness_raw_callback_truncation_control_4e0960() != 0
}

func readinessRawLayout4E0960() (out [15]uintptr) {
	for i := range out {
		out[i] = uintptr(C.readiness_raw_layout_4e0960(C.int(i)))
	}
	return
}

func readinessRawABI4E0960() int { return int(C.readiness_raw_abi_4e0960()) }

func readinessRawInitialC4E0960(base unsafe.Pointer, kind, index int) uintptr {
	return uintptr(C.readiness_raw_initial_read_4e0960(base, C.int(kind), C.int(index)))
}

func readinessRawPublic4E0960(item unsafe.Pointer) int32 {
	return int32(C.readiness_raw_public_4e0960((*C.nox_object_t)(item)))
}
