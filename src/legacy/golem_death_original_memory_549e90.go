package legacy

/*
#include <stdint.h>
#include <stddef.h>
#ifdef _WIN32
#include <windows.h>
static void* golem_fixture_guard_549e90(size_t* size) {
 SYSTEM_INFO info;GetSystemInfo(&info);*size=info.dwPageSize;
 return VirtualAlloc(NULL,*size,MEM_RESERVE|MEM_COMMIT,PAGE_NOACCESS);
}
static void golem_fixture_guard_free_549e90(void* p,size_t size) {VirtualFree(p,0,MEM_RELEASE);}
#else
#include <sys/mman.h>
#include <unistd.h>
static void* golem_fixture_guard_549e90(size_t* size) {
 *size=(size_t)sysconf(_SC_PAGESIZE);
 void* p=mmap(NULL,*size,PROT_NONE,MAP_PRIVATE|MAP_ANONYMOUS,-1,0);
 return p==MAP_FAILED?NULL:p;
}
static void golem_fixture_guard_free_549e90(void* p,size_t size) {munmap(p,size);}
#endif
// The original loaded PE occupies [0x400000,0x981000). Relocating a fixture
// base cannot declare unrelated native mappings part of that original image.
// Only originally unmapped projections use the protected fixture page. This
// is reference-memory transport, not an engine/table admission check.
static void* golem_fixture_mass_address_549e90(void* table,int32_t index,void* unmapped) {
 uint32_t address=0x5cd264u+(uint32_t)index;
 if(address<0x400000u || address>=0x981000u)return unmapped;
 return (char*)table+index;
}
static void* golem_fixture_slot_address_549e90(void* table,int32_t index,uint32_t original_base,void* unmapped) {
 uint32_t scaled=(uint32_t)index*4u;
 uint32_t address=original_base+scaled;
 if(address<0x400000u || address>=0x981000u)return unmapped;
 int32_t offset=(int32_t)scaled;
 return (char**)table+offset/4;
}
*/
import "C"
import "unsafe"

func golemOriginalGuardPage549E90() (unsafe.Pointer, func()) {
	var size C.size_t
	p := C.golem_fixture_guard_549e90(&size)
	return p, func() {
		if p != nil {
			C.golem_fixture_guard_free_549e90(p, size)
		}
	}
}
func golemOriginalMappedMassAddress549E90(table unsafe.Pointer, index int32, unmapped unsafe.Pointer) unsafe.Pointer {
	return C.golem_fixture_mass_address_549e90(table, C.int32_t(index), unmapped)
}

func golemOriginalMappedSlotAddress549E90(table unsafe.Pointer, index int32, originalBase uint32, unmapped unsafe.Pointer) unsafe.Pointer {
	return C.golem_fixture_slot_address_549e90(table, C.int32_t(index), C.uint32_t(originalBase), unmapped)
}
