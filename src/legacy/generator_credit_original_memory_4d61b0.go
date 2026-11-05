package legacy

/*
#include <stdint.h>
#include "defs.h"
#ifdef _WIN32
#include <windows.h>
static void* generator_credit_heap_4d61b0(size_t* guard){
 SYSTEM_INFO info;GetSystemInfo(&info);size_t page=info.dwPageSize;
 *guard=((sizeof(nox_playerInfo)+page-1)/page)*page;
 unsigned char* base=VirtualAlloc(NULL,*guard+0x10000,MEM_RESERVE|MEM_COMMIT|MEM_TOP_DOWN,PAGE_READWRITE);
 if(!base)return NULL;
 DWORD old;if(!VirtualProtect(base,*guard,PAGE_NOACCESS,&old)){VirtualFree(base,0,MEM_RELEASE);return NULL;}
 return base+*guard;
}
static void generator_credit_free_4d61b0(void* heap,size_t guard){VirtualFree((unsigned char*)heap-guard,0,MEM_RELEASE);}
#else
#include <sys/mman.h>
#include <unistd.h>
static void* generator_credit_heap_4d61b0(size_t* guard){
 size_t page=(size_t)sysconf(_SC_PAGESIZE);
 *guard=((sizeof(nox_playerInfo)+page-1)/page)*page;
 unsigned char* base=mmap(NULL,*guard+0x10000,PROT_READ|PROT_WRITE,MAP_PRIVATE|MAP_ANONYMOUS,-1,0);
 if(base==MAP_FAILED)return NULL;
 if(mprotect(base,*guard,PROT_NONE)){munmap(base,*guard+0x10000);return NULL;}
 return base+*guard;
}
static void generator_credit_free_4d61b0(void* heap,size_t guard){munmap((unsigned char*)heap-guard,guard+0x10000);}
#endif
// Form malformed/null addresses only in C. A whole native Player fits in the
// protected predecessor on both 4-KiB and 16-KiB page hosts.
static void* generator_credit_bad_4d61b0(void* heap,size_t guard,int slot){return (void*)((uintptr_t)heap-guard+32*(uintptr_t)slot);}
static void* generator_credit_field_4d61b0(void* player,int flags){return (void*)((uintptr_t)player+(flags?offsetof(nox_playerInfo,field_4692):offsetof(nox_playerInfo,field_4668)));}
*/
import "C"

import "unsafe"

func generatorCreditOriginalHeap4D61B0() (unsafe.Pointer, uintptr, func()) {
	var guard C.size_t
	heap := C.generator_credit_heap_4d61b0(&guard)
	return heap, uintptr(guard), func() {
		if heap != nil {
			C.generator_credit_free_4d61b0(heap, guard)
		}
	}
}

func generatorCreditOriginalBad4D61B0(heap unsafe.Pointer, guard uintptr, slot int) unsafe.Pointer {
	return C.generator_credit_bad_4d61b0(heap, C.size_t(guard), C.int(slot))
}

func generatorCreditOriginalField4D61B0(player unsafe.Pointer, flags bool) unsafe.Pointer {
	v := 0
	if flags {
		v = 1
	}
	return C.generator_credit_field_4d61b0(player, C.int(v))
}
