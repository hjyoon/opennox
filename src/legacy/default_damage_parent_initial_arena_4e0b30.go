//go:build darwin && arm64

package legacy

/*
#include <string.h>
#include <sys/mman.h>
#include <unistd.h>
#include "default_damage_parent_full_initial_4e0b30.h"
#ifndef MAP_ANON
#define MAP_ANON MAP_ANONYMOUS
#endif
static void* ddi_map_4e0b30(size_t page) {
 void* p=mmap(0,2*page,PROT_READ|PROT_WRITE,MAP_PRIVATE|MAP_ANON,-1,0);
 if(p==MAP_FAILED)return 0;
 if(mprotect((unsigned char*)p+page,page,PROT_NONE)!=0){munmap(p,2*page);return 0;}
 return p;
}
static void ddi_unmap_4e0b30(void* p,size_t page) {munmap(p,2*page);}
static void ddi_fill_4e0b30(void* p,size_t n,unsigned int b) {memset(p,(int)b,n);}
static void ddi_copy_4e0b30(void* dst,const void* src,size_t n) {memcpy(dst,src,n);}
static int ddi_put_4e0b30(void* base,uint32_t id,uintptr_t go_offset,const void* bytes,size_t n) {
 uintptr_t c=ddi_c_offset_4e0b30(id);
 if(c!=UINTPTR_MAX && c!=go_offset)return 0;
 // The four monster fields explicitly use a Go-derived adapter offset.
 memcpy((unsigned char*)base+(c==UINTPTR_MAX?go_offset:c),bytes,n);return 1;
}
static int ddi_put_pointer_4e0b30(void* base,uint32_t id,uintptr_t go_offset,void* value) {
 uintptr_t c=ddi_c_offset_4e0b30(id);
 if(!ddi_pointer_field_4e0b30(id)||c==UINTPTR_MAX||c!=go_offset)return 0;
 memcpy((unsigned char*)base+c,&value,sizeof(value));return 1;
}
static uintptr_t ddi_arena_c_offset_4e0b30(uint32_t id){return ddi_c_offset_4e0b30(id);}
static uintptr_t ddi_arena_callback_4e0b30(uint32_t kind){return ddi_callback_4e0b30(kind);}
static void* ddi_arena_callback_pointer_4e0b30(uint32_t kind){return(void*)ddi_callback_4e0b30(kind);}
*/
import "C"

import (
	"runtime"
	"unsafe"
)

func ddiMap4E0B30(page int) unsafe.Pointer      { return C.ddi_map_4e0b30(C.size_t(page)) }
func ddiUnmap4E0B30(p unsafe.Pointer, page int) { C.ddi_unmap_4e0b30(p, C.size_t(page)) }
func ddiFill4E0B30(p unsafe.Pointer, n int, value byte) {
	C.ddi_fill_4e0b30(p, C.size_t(n), C.uint(value))
}
func ddiCopyIn4E0B30(p unsafe.Pointer, data []byte) {
	if len(data) != 0 {
		C.ddi_copy_4e0b30(p, unsafe.Pointer(&data[0]), C.size_t(len(data)))
	}
	runtime.KeepAlive(data)
}
func ddiSnapshot4E0B30(p unsafe.Pointer, n int) []byte {
	out := make([]byte, n)
	if n != 0 {
		C.ddi_copy_4e0b30(unsafe.Pointer(&out[0]), p, C.size_t(n))
	}
	runtime.KeepAlive(p)
	return out
}
func ddiPut4E0B30(base unsafe.Pointer, id int, offset uintptr, data []byte) bool {
	if len(data) == 0 {
		return true
	}
	ok := C.ddi_put_4e0b30(base, C.uint32_t(id), C.uintptr_t(offset), unsafe.Pointer(&data[0]), C.size_t(len(data)))
	runtime.KeepAlive(base)
	runtime.KeepAlive(data)
	return ok != 0
}
func ddiPutPointer4E0B30(base unsafe.Pointer, id int, offset uintptr, value unsafe.Pointer) bool {
	ok := C.ddi_put_pointer_4e0b30(base, C.uint32_t(id), C.uintptr_t(offset), value)
	runtime.KeepAlive(base)
	runtime.KeepAlive(value)
	return ok != 0
}
func ddiCallback4E0B30(kind uint32) (unsafe.Pointer, uintptr) {
	return C.ddi_arena_callback_pointer_4e0b30(C.uint32_t(kind)), uintptr(C.ddi_arena_callback_4e0b30(C.uint32_t(kind)))
}
func ddiCOffset4E0B30(id int) uintptr { return uintptr(C.ddi_arena_c_offset_4e0b30(C.uint32_t(id))) }
