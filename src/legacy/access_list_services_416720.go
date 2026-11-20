package legacy

/*
#include <stdlib.h>
#include <stddef.h>
#include "GAME1.h"
#include "GAME2.h"
#include "GAME3_3.h"
static void* access_list_config_416720(void) { return getMemAt(0x587000,54280); }
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

	"github.com/opennox/opennox/v1/legacy/common/alloc"
)

var (
	accessListNativeOwnersOnce416720 sync.Once
	accessListHeads416720, _         = alloc.New([2]accessListNode416720{})
)

func accessListNativeEnsure416720() {
	accessListNativeOwnersOnce416720.Do(func() {
		// The native sentinels remain at stable addresses for the process lifetime.
		for i := range accessListHeads416720 {
			accessListNativeClear416720(&accessListHeads416720[i])
		}
	})
}

func accessListNativeDefaultHead416720(banned bool) unsafe.Pointer {
	accessListNativeEnsure416720()
	if banned {
		return unsafe.Pointer(&accessListHeads416720[1])
	}
	return unsafe.Pointer(&accessListHeads416720[0])
}
func accessListNativeDefaultAlloc416720(size uintptr) unsafe.Pointer {
	return C.calloc(1, C.size_t(size))
}
func accessListNativeDefaultFree416720(p unsafe.Pointer)   { C.free(p); runtime.KeepAlive(p) }
func accessListNativeDefaultRefresh416720() unsafe.Pointer { return unsafe.Pointer(C.sub_455800()) }
func accessListNativeConfig416720() unsafe.Pointer         { return C.access_list_config_416720() }
func accessListNativeDefaultLoad416720(p unsafe.Pointer) int32 {
	loaded := loadAccessList4E41B0(GoString((*C.char)(p)))
	runtime.KeepAlive(p)
	return int32(bool2int(loaded))
}
func accessListNativeDefaultSave416720() { saveAccessList4E43F0("ban.txt") }

//export sub_4E41B0
func sub_4E41B0(path *C.char) C.int {
	return C.int(bool2int(loadAccessList4E41B0(GoString(path))))
}

//export sub_4E42C0
func sub_4E42C0(file *FILE) C.int {
	return C.int(bool2int(loadAccessBanned4E42C0(fileByHandle(file))))
}

//export sub_4E4390
func sub_4E4390(file *FILE) C.int {
	loadAccessAllowed4E4390(fileByHandle(file))
	return 1
}

func accessListLoadC4E41B0(path string) int {
	p := CString(path)
	defer C.free(unsafe.Pointer(p))
	return int(C.sub_4E41B0(p))
}

func accessListLoadSectionC4E41B0(file *FILE, banned bool) int {
	if banned {
		return int(C.sub_4E42C0(file))
	}
	return int(C.sub_4E4390(file))
}

//export sub_4E43F0
func sub_4E43F0(path *C.char) C.int {
	return C.int(bool2int(saveAccessList4E43F0(GoString(path))))
}

func accessListSaveC4E43F0(path string) int {
	p := CString(path)
	defer C.free(unsafe.Pointer(p))
	return int(C.sub_4E43F0(p))
}

func accessListNativeCLayout416720() [6]uintptr {
	var a [6]C.size_t
	C.access_list_layout_416720(&a[0])
	var out [6]uintptr
	for i, v := range a {
		out[i] = uintptr(v)
	}
	return out
}
