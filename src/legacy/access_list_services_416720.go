package legacy

/*
#include <stdlib.h>
#include <stddef.h>
#include "GAME1.h"
#include "GAME2.h"
#include "GAME3_3.h"
// Defined by the tiny accessor inserted after the existing static C owners.
extern void* nox_access_native_head_416720(int banned);
static void* access_list_config_416720(void) { return getMemAt(0x587000,54280); }
static void access_list_save_416720(void) { sub_4E43F0("ban.txt"); }
static void access_list_layout_416720(size_t* a) {
    a[0]=sizeof(nox_list_item_t);a[1]=sizeof(nox_access_allowed_entry_t);a[2]=sizeof(nox_access_banned_entry_t);
    a[3]=offsetof(nox_access_allowed_entry_t,name);a[4]=offsetof(nox_access_banned_entry_t,expires_at_ms);a[5]=offsetof(nox_access_banned_entry_t,serial);
}
*/
import "C"
import (
	"runtime"
	"unsafe"
)

func accessListNativeDefaultHead416720(banned bool) unsafe.Pointer {
	v := C.int(0)
	if banned {
		v = 1
	}
	return C.nox_access_native_head_416720(v)
}
func accessListNativeDefaultAlloc416720(size uintptr) unsafe.Pointer {
	return C.calloc(1, C.size_t(size))
}
func accessListNativeDefaultFree416720(p unsafe.Pointer)   { C.free(p); runtime.KeepAlive(p) }
func accessListNativeDefaultRefresh416720() unsafe.Pointer { return unsafe.Pointer(C.sub_455800()) }
func accessListNativeConfig416720() unsafe.Pointer         { return C.access_list_config_416720() }
func accessListNativeDefaultLoad416720(p unsafe.Pointer) int32 {
	v := int32(C.sub_4E41B0((*C.char)(p)))
	runtime.KeepAlive(p)
	return v
}
func accessListNativeDefaultSave416720() { C.access_list_save_416720() }
func accessListNativeCLayout416720() [6]uintptr {
	var a [6]C.size_t
	C.access_list_layout_416720(&a[0])
	var out [6]uintptr
	for i, v := range a {
		out[i] = uintptr(v)
	}
	return out
}
