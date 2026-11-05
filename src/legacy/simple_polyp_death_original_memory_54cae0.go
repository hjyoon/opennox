package legacy

/*
#include <stddef.h>
#include <stdint.h>
#include "defs.h"
#ifdef _WIN32
#include <windows.h>
static void* simple_polyp_heap_54cae0(size_t* page) {
 SYSTEM_INFO info;GetSystemInfo(&info);*page=info.dwPageSize;
 unsigned char* base=VirtualAlloc(NULL,0x10000+*page,MEM_RESERVE|MEM_COMMIT|MEM_TOP_DOWN,PAGE_READWRITE);
 if(!base)return NULL;
 DWORD old;
 if(!VirtualProtect(base,*page,PAGE_NOACCESS,&old)){VirtualFree(base,0,MEM_RELEASE);return NULL;}
 return base+*page;
}
static void simple_polyp_heap_free_54cae0(void* heap,size_t page) {VirtualFree((unsigned char*)heap-page,0,MEM_RELEASE);}
#else
#include <sys/mman.h>
#include <unistd.h>
static void* simple_polyp_heap_54cae0(size_t* page) {
 *page=(size_t)sysconf(_SC_PAGESIZE);
 unsigned char* base=mmap(NULL,0x10000+*page,PROT_READ|PROT_WRITE,MAP_PRIVATE|MAP_ANONYMOUS,-1,0);
 if(base==MAP_FAILED)return NULL;
 if(mprotect(base,*page,PROT_NONE)){munmap(base,0x10000+*page);return NULL;}
 return base+*page;
}
static void simple_polyp_heap_free_54cae0(void* heap,size_t page) {munmap((unsigned char*)heap-page,0x10000+page);}
#endif
// Relocate the original unmapped role to an actual protected native page.
// Literal 0x41000000 can be mapped in an unrelated native address space.
// Preserve real nil and use independently compiled C offsets for partial
// objects, avoiding a nil+offset Go pointer live across callbacks or GC.
static void* simple_polyp_bad_54cae0(void* heap,size_t page,int role) {
 if(role==-1)return NULL;
 if(role==-2)return (unsigned char*)heap-page;
 if(role==-3)return (unsigned char*)heap- offsetof(nox_object_t,y);
 if(role==-4) {
  uintptr_t align=_Alignof(nox_object_t);
  uintptr_t source=(uintptr_t)heap- offsetof(nox_object_t,obj_flags);
  return (void*)((source+align-1)&~(align-1));
 }
 return NULL;
}
*/
import "C"

import "unsafe"

func simplePolypOriginalHeap54CAE0() (unsafe.Pointer, uintptr, func()) {
	var page C.size_t
	heap := C.simple_polyp_heap_54cae0(&page)
	return heap, uintptr(page), func() {
		if heap != nil {
			C.simple_polyp_heap_free_54cae0(heap, page)
		}
	}
}

func simplePolypOriginalBad54CAE0(heap unsafe.Pointer, page uintptr, role int) unsafe.Pointer {
	return C.simple_polyp_bad_54cae0(heap, C.size_t(page), C.int(role))
}
