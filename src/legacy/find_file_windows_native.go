//go:build windows

package legacy

/*
#include <stdint.h>
#include "compat.h"

uintptr_t nox_win_find_first_file_word(const char*, WIN32_FIND_DATAA*);
int nox_win_find_next_file_word(uintptr_t, WIN32_FIND_DATAA*);
int nox_win_find_close_word(uintptr_t);
*/
import "C"

import "unsafe"

func FindFirstFileA(pattern *byte, entry *WIN32_FIND_DATAA) uintptr {
	return uintptr(C.nox_win_find_first_file_word((*C.char)(unsafe.Pointer(pattern)), (*_WIN32_FIND_DATAA)(unsafe.Pointer(entry))))
}

func FindNextFileA(handle uintptr, entry *WIN32_FIND_DATAA) int {
	return int(C.nox_win_find_next_file_word(C.uintptr_t(handle), (*_WIN32_FIND_DATAA)(unsafe.Pointer(entry))))
}

func closeFileSearch(handle uintptr) int {
	return int(C.nox_win_find_close_word(C.uintptr_t(handle)))
}
