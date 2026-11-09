//go:build darwin && arm64

package legacy

/*
#include <sys/mman.h>
#include <unistd.h>
static void* sp_guard_map_4134d0(void) {
    size_t page=(size_t)sysconf(_SC_PAGESIZE);
    void* p=mmap(0,page*3,PROT_NONE,MAP_PRIVATE|MAP_ANON,-1,0);
    if(p==MAP_FAILED)return 0;
    void* middle=(char*)p+page;
    if(mprotect(middle,page,PROT_READ|PROT_WRITE)){munmap(p,page*3);return 0;}
    return middle;
}
static void sp_guard_unmap_4134d0(void* middle) {
    size_t page=(size_t)sysconf(_SC_PAGESIZE);munmap((char*)middle-page,page*3);
}
*/
import "C"
import "unsafe"

func startupPauseGuardMap4134D0() unsafe.Pointer    { return C.sp_guard_map_4134d0() }
func startupPauseGuardUnmap4134D0(p unsafe.Pointer) { C.sp_guard_unmap_4134d0(p) }
