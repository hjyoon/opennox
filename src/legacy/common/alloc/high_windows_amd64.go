package alloc

/*
#include <stdint.h>
#include <windows.h>
static void* nox_fixture_alloc_high(size_t size) {
 void*p=VirtualAlloc(NULL,size,MEM_RESERVE|MEM_COMMIT|MEM_TOP_DOWN,PAGE_READWRITE);
 if(p&&(uintptr_t)p<=UINT32_MAX){VirtualFree(p,0,MEM_RELEASE);return NULL;}
 return p;
}
static int nox_fixture_free_high(void*p) { return VirtualFree(p,0,MEM_RELEASE)!=0; }
*/
import "C"

import "unsafe"

func mallocHighMapping(size uintptr) (unsafe.Pointer, func()) {
	ptr := C.nox_fixture_alloc_high(C.size_t(size))
	return ptr, func() {
		if C.nox_fixture_free_high(ptr) == 0 {
			panic("cannot release high native mapping")
		}
	}
}
