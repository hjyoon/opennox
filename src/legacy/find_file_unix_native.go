//go:build !windows

package legacy

/*
#include <glob.h>
#include "windows_compat.h"
#include "cgo_const.h"
*/
import "C"

import (
	"os"
	"strings"
	"sync"
	"unsafe"

	"github.com/opennox/libs/ifs"

	"github.com/opennox/opennox/v1/legacy/common/alloc"
)

type fileSearch struct {
	remaining []string
	release   func()
}

var findFileSearches = struct {
	sync.Mutex
	active map[uintptr]*fileSearch
}{active: make(map[uintptr]*fileSearch)}

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

func fillFindFileData(path string, entry *WIN32_FIND_DATAA) {
	clear(unsafe.Slice((*byte)(unsafe.Pointer(entry)), int(unsafe.Sizeof(*entry))))
	entry.FileAttributes = C.FILE_ATTRIBUTE_NORMAL
	if info, err := os.Stat(path); err == nil {
		if info.IsDir() {
			entry.FileAttributes = C.FILE_ATTRIBUTE_DIRECTORY
		}
		size := uint64(info.Size())
		entry.nFileSizeHigh = uint32(size >> 32)
		entry.nFileSizeLow = uint32(size)
	}
	StrCopyBytes(entry.cFileName[:], path[strings.LastIndexByte(path, '/')+1:])
}

func FindFirstFileA(pattern *byte, entry *WIN32_FIND_DATAA) uintptr {
	if pattern == nil || entry == nil {
		return invalidFindFileHandle
	}
	paths := findFilePaths(alloc.GoString(pattern))
	if len(paths) == 0 {
		return invalidFindFileHandle
	}
	raw, release := alloc.Malloc(1)
	handle := uintptr(raw)
	fillFindFileData(paths[0], entry)
	findFileSearches.Lock()
	findFileSearches.active[handle] = &fileSearch{remaining: paths[1:], release: release}
	findFileSearches.Unlock()
	return handle
}

func FindNextFileA(handle uintptr, entry *WIN32_FIND_DATAA) int {
	if entry == nil {
		return 0
	}
	findFileSearches.Lock()
	search := findFileSearches.active[handle]
	if search == nil || len(search.remaining) == 0 {
		findFileSearches.Unlock()
		return 0
	}
	path := search.remaining[0]
	search.remaining = search.remaining[1:]
	findFileSearches.Unlock()
	fillFindFileData(path, entry)
	return 1
}

func closeFileSearch(handle uintptr) int {
	findFileSearches.Lock()
	search := findFileSearches.active[handle]
	delete(findFileSearches.active, handle)
	findFileSearches.Unlock()
	if search == nil {
		return 0
	}
	search.release()
	return 1
}

//export compatFindFirstFileA
func compatFindFirstFileA(pattern *C.nox_cgo_const_char, entry *C.WIN32_FIND_DATAA) C.HANDLE {
	return C.HANDLE(FindFirstFileA((*byte)(unsafe.Pointer(pattern)), (*WIN32_FIND_DATAA)(unsafe.Pointer(entry))))
}

//export compatFindNextFileA
func compatFindNextFileA(handle C.HANDLE, entry *C.WIN32_FIND_DATAA) C.int {
	return C.int(FindNextFileA(uintptr(handle), (*WIN32_FIND_DATAA)(unsafe.Pointer(entry))))
}

//export compatFindClose
func compatFindClose(handle C.HANDLE) C.int {
	return C.int(closeFileSearch(uintptr(handle)))
}
