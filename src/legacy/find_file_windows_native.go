//go:build windows

package legacy

import (
	"unsafe"

	"golang.org/x/sys/windows"
)

// The ANSI system calls write the WIN32_FIND_DATAA record used by saved
// players and audio lookup. The path and record are borrowed for one call.
var (
	findFileKernel     = windows.NewLazySystemDLL("kernel32.dll")
	findFirstFileAProc = findFileKernel.NewProc("FindFirstFileA")
	findNextFileAProc  = findFileKernel.NewProc("FindNextFileA")
	findCloseProc      = findFileKernel.NewProc("FindClose")
)

func FindFirstFileA(pattern *byte, entry *WIN32_FIND_DATAA) uintptr {
	handle, _, _ := findFirstFileAProc.Call(uintptr(unsafe.Pointer(pattern)), uintptr(unsafe.Pointer(entry)))
	return handle
}

func FindNextFileA(handle uintptr, entry *WIN32_FIND_DATAA) int {
	status, _, _ := findNextFileAProc.Call(handle, uintptr(unsafe.Pointer(entry)))
	return int(int32(status))
}

func closeFileSearch(handle uintptr) int {
	status, _, _ := findCloseProc.Call(handle)
	return int(int32(status))
}
