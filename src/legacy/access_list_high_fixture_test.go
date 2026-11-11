package legacy

import (
	"unsafe"

	"github.com/opennox/opennox/v1/legacy/common/alloc"
)

// Preserve foreign ownership while explicitly requesting the high addresses
// required by these controls. Product default calloc/free remain unchanged.
func accessListNativeFixtureAlloc416720(size uintptr) unsafe.Pointer {
	p, _ := alloc.MallocHigh(size)
	return p
}

func accessListNativeFixtureFree416720(p unsafe.Pointer) {
	if p != nil {
		alloc.FreePtr(p)
	}
}
