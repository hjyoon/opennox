package legacy

import (
	"testing"
	"unsafe"

	"github.com/opennox/opennox/v1/common/memmap"
)

func TestLegacyMemoryPointerPreservesBlobOrder(t *testing.T) {
	for _, base := range legacyMemoryBlobOrder {
		blob := memmap.BlobByAddr(base)
		address := unsafe.Pointer(&blob.Data[0])
		if got := legacyMemoryPointer(base, 0, 1); got != address {
			t.Fatalf("blob %#x starts at %p, want %p", base, got, address)
		}
		end := uintptr(len(blob.Data))
		if got := legacyMemoryPointer(base, end, 0); got != unsafe.Add(address, end) {
			t.Fatalf("blob %#x zero-length end = %p", base, got)
		}
	}

	offset := uintptr(len(memmap.BlobByAddr(0x581450).Data))
	expect := memmap.PtrOff(0x587000, offset)
	if got := legacyMemoryPointer(0x581450, offset, 1); got != expect {
		t.Fatalf("fallthrough pointer = %p, want %p", got, expect)
	}

	for _, tc := range []struct {
		base, off, size uintptr
	}{
		{0x581451, 0, 1},
		{0x581450, ^uintptr(0), 2},
		{0x973A20, 0, ^uintptr(0)},
	} {
		if got := legacyMemoryPointer(tc.base, tc.off, tc.size); got != nil {
			t.Fatalf("out-of-bounds %#x+%d[%d] = %p", tc.base, tc.off, tc.size, got)
		}
	}
}
