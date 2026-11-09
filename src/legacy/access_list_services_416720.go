package legacy

/*
#include <stdlib.h>
#include <stddef.h>
#include "GAME1.h"
#include "GAME2.h"
#include "GAME3_3.h"
// Only storage remains in C. Initialization and field access live in Go.
extern nox_list_item_t nox_access_allowed_list_4168E0;
extern nox_list_item_t nox_access_banned_list_416900;
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
	"sync"
	"unsafe"
)

var accessListNativeOwnersOnce416720 sync.Once

func accessListNativeEnsure416720() {
	accessListNativeOwnersOnce416720.Do(func() {
		// Keep the existing C-owned, native-width sentinels for their entire
		// process lifetime. C links never retain pointers into the Go heap.
		accessListNativeClear416720((*accessListNode416720)(unsafe.Pointer(&C.nox_access_allowed_list_4168E0)))
		accessListNativeClear416720((*accessListNode416720)(unsafe.Pointer(&C.nox_access_banned_list_416900)))
	})
}

func accessListNativeDefaultHead416720(banned bool) unsafe.Pointer {
	accessListNativeEnsure416720()
	if banned {
		return unsafe.Pointer(&C.nox_access_banned_list_416900)
	}
	return unsafe.Pointer(&C.nox_access_allowed_list_4168E0)
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
