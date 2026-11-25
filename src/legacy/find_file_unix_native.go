//go:build !windows

package legacy

/*
#include <glob.h>
#include "windows_compat.h"
#include "cgo_const.h"

static inline HANDLE nox_find_invalid_handle(void) {
	return (HANDLE)(intptr_t)-1;
}
*/
import "C"

import (
	"os"
	"strings"
	"sync"
	"unsafe"

	"github.com/opennox/libs/ifs"
)

var findFileSearches = struct {
	sync.Mutex
	paths map[C.HANDLE][]string
}{paths: make(map[C.HANDLE][]string)}

func findFilePaths(pattern string) []string {
	pattern = strings.TrimSuffix(ifs.Normalize(pattern), ".*")
	input := CString(pattern)
	defer StrFree(input)
	var matches C.glob_t
	if C.glob(input, C.GLOB_NOESCAPE, nil, &matches) != 0 {
		C.globfree(&matches)
		return nil
	}
	defer C.globfree(&matches)

	paths := make([]string, int(matches.gl_pathc))
	for i, path := range unsafe.Slice(matches.gl_pathv, len(paths)) {
		paths[i] = GoString(path)
	}
	return paths
}

func fillFindFileData(path string, entry *C.WIN32_FIND_DATAA) {
	clear(unsafe.Slice((*byte)(unsafe.Pointer(entry)), int(unsafe.Sizeof(*entry))))
	entry.dwFileAttributes = C.FILE_ATTRIBUTE_NORMAL
	if info, err := os.Stat(path); err == nil {
		if info.IsDir() {
			entry.dwFileAttributes = C.FILE_ATTRIBUTE_DIRECTORY
		}
		size := uint64(info.Size())
		entry.nFileSizeHigh = C.uint32_t(size >> 32)
		entry.nFileSizeLow = C.uint32_t(size)
	}
	StrCopy(&entry.cFileName[0], len(entry.cFileName), path[strings.LastIndexByte(path, '/')+1:])
}

//export compatFindFirstFileA
func compatFindFirstFileA(pattern *C.nox_cgo_const_char, entry *C.WIN32_FIND_DATAA) C.HANDLE {
	if pattern == nil || entry == nil {
		return C.nox_find_invalid_handle()
	}
	paths := findFilePaths(GoString((*C.char)(pattern)))
	if len(paths) == 0 {
		return C.nox_find_invalid_handle()
	}
	handle := C.HANDLE(C.malloc(1))
	fillFindFileData(paths[0], entry)
	findFileSearches.Lock()
	findFileSearches.paths[handle] = paths[1:]
	findFileSearches.Unlock()
	return handle
}

//export compatFindNextFileA
func compatFindNextFileA(handle C.HANDLE, entry *C.WIN32_FIND_DATAA) C.int {
	if entry == nil {
		return 0
	}
	findFileSearches.Lock()
	paths := findFileSearches.paths[handle]
	if len(paths) == 0 {
		findFileSearches.Unlock()
		return 0
	}
	findFileSearches.paths[handle] = paths[1:]
	findFileSearches.Unlock()
	fillFindFileData(paths[0], entry)
	return 1
}

//export compatFindClose
func compatFindClose(handle C.HANDLE) C.int {
	findFileSearches.Lock()
	_, ok := findFileSearches.paths[handle]
	delete(findFileSearches.paths, handle)
	findFileSearches.Unlock()
	if !ok {
		return 0
	}
	C.free(unsafe.Pointer(handle))
	return 1
}
