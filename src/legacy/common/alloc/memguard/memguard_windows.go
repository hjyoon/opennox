//go:build windows

package memguard

import (
	"unsafe"

	"golang.org/x/sys/windows"
)

func newPage(size int) ([]byte, func()) {
	return newPageFlags(size, windows.MEM_RESERVE)
}

func newPageHigh(size int) ([]byte, func()) {
	flags := uint32(windows.MEM_RESERVE)
	if unsafe.Sizeof(uintptr(0)) == 8 {
		flags |= windows.MEM_TOP_DOWN
	}
	return newPageFlags(size, flags)
}

func newPageFlags(size int, flags uint32) ([]byte, func()) {
	page := PageSize()
	num := size / page
	if size%page != 0 {
		num++
	}
	addr, err := windows.VirtualAlloc(0, uintptr(page*num), flags, windows.PAGE_NOACCESS)
	if err != nil {
		panic(err)
	}
	if flags&windows.MEM_TOP_DOWN != 0 && addr <= uintptr(^uint32(0)) {
		_ = windows.VirtualFree(addr, 0, windows.MEM_RELEASE)
		panic("high opaque fixture reservation is below 4 GiB")
	}
	b := unsafe.Slice((*byte)(unsafe.Pointer(addr)), size)
	return b, func() {
		if err := windows.VirtualFree(addr, 0, windows.MEM_RELEASE); err != nil {
			panic(err)
		}
	}
}
