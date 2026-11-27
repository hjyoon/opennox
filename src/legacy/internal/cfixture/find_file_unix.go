//go:build !windows

package cfixture

/*
#include "windows_compat.h"

typedef uintptr_t (*find_first_entry)(const char*, WIN32_FIND_DATAA*);
typedef int (*find_next_entry)(uintptr_t, WIN32_FIND_DATAA*);
typedef int (*find_close_entry)(uintptr_t);
_Static_assert(_Generic(&FindFirstFileA, find_first_entry: 1, default: 0), "find first ABI");
_Static_assert(_Generic(&FindNextFileA, find_next_entry: 1, default: 0), "find next ABI");
_Static_assert(_Generic(&FindClose, find_close_entry: 1, default: 0), "find close ABI");

static uintptr_t find_first_fixture(const char* path, WIN32_FIND_DATAA* entry) {
	find_first_entry volatile first = FindFirstFileA;
	return first(path, entry);
}

static int find_next_fixture(uintptr_t handle, WIN32_FIND_DATAA* entry) {
	find_next_entry volatile next = FindNextFileA;
	return next(handle, entry);
}

static int find_close_fixture(uintptr_t handle) {
	find_close_entry volatile close_search = FindClose;
	return close_search(handle);
}
*/
import "C"

import "unsafe"

func FindFileFirst(path *byte, entry unsafe.Pointer) uintptr {
	return uintptr(C.find_first_fixture((*C.char)(unsafe.Pointer(path)), (*C.WIN32_FIND_DATAA)(entry)))
}

func FindFileNext(handle uintptr, entry unsafe.Pointer) int {
	return int(C.find_next_fixture(C.uintptr_t(handle), (*C.WIN32_FIND_DATAA)(entry)))
}

func FindFileClose(handle uintptr) int {
	return int(C.find_close_fixture(C.uintptr_t(handle)))
}
