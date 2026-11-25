//go:build amd64 || arm64

package alloc

import (
	"math"
	"unsafe"

	"golang.org/x/sys/unix"
)

// The address passed to mmap is a placement hint for an anonymous mapping.
//
//go:nocheckptr
func mallocHighMapping(size uintptr) (unsafe.Pointer, func()) {
	const addressHint = 1 << 36
	ptr, err := unix.MmapPtr(-1, 0, unsafe.Pointer(uintptr(addressHint)), size,
		unix.PROT_READ|unix.PROT_WRITE, unix.MAP_PRIVATE|unix.MAP_ANONYMOUS)
	if err != nil {
		panic(err)
	}
	release := func() {
		if err := unix.MunmapPtr(ptr, size); err != nil {
			panic(err)
		}
	}
	if uintptr(ptr) <= math.MaxUint32 {
		release()
		panic("high native mapping is below 4 GiB")
	}
	return ptr, release
}
