package handles

import (
	"fmt"
	"sync/atomic"
	"unsafe"

	"github.com/opennox/opennox/v1/legacy/common/alloc/memguard"
)

var (
	data []byte
	base uintptr
	end  uintptr
	cur  uint32
	free func()
)

// Init the handles memory region.
func Init() {
	const size = 16 * 1024 * 1024
	data, free = memguard.New(size)
	base = uintptr(unsafe.Pointer(&data[0]))
	end = base + size
	atomic.StoreUint32(&cur, 0)
}

const handleStride = uint32(unsafe.Alignof(uintptr(0)))

func nextOffset() uintptr {
	if base == 0 {
		panic("call Init first")
	}
	off := uintptr(atomic.AddUint32(&cur, handleStride))
	if base+off >= end {
		panic("no more handles") // TODO
	}
	return off
}

// New creates a new unique opaque handle for application to use.
func New() uintptr {
	return base + nextOffset()
}

// NewPtr creates a new unique opaque handle for application to use. It casts the value to a pointer.
func NewPtr() unsafe.Pointer {
	off := nextOffset()
	return unsafe.Pointer(&data[off])
}

// IsValid checks if a handle is valid.
func IsValid(h uintptr) bool {
	return h >= base && h < end
}

// AssertValid checks if a handle is valid and panic otherwise.
func AssertValid(h uintptr) {
	if h == 0 {
		panic("zero handle")
	}
	if h < base || h >= end {
		panic(fmt.Errorf("invalid handle: %v", h))
	}
}

// AssertValidPtr checks if a handle is valid and panic otherwise.
func AssertValidPtr(p unsafe.Pointer) {
	AssertValid(uintptr(p))
}

// Release all the handles and associated protected memory pages.
func Release() {
	if free != nil {
		free()
		free = nil
		data = nil
		base = 0
		end = 0
	}
}
