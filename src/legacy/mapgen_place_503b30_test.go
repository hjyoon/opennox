package legacy

import (
	"runtime"
	"testing"
	"unsafe"
)

func TestMapgenPendingScriptIDs503B30PreserveNativeObjectFields(t *testing.T) {
	got := mapgenPendingFixture503B30()
	if got.scriptIDs != [3]int32{} {
		t.Fatalf("pending script IDs = %v, want three zeroes", got.scriptIDs)
	}
	if got.extents != [3]uint32{0x10203040, 0x50607080, 0x90a0b0c0} {
		t.Errorf("object extents changed: %x", got.extents)
	}
	if got.field12 != [3]uint32{0x11223344, 0x55667788, 0x99aabbcc} {
		t.Errorf("adjacent fields changed: %x", got.field12)
	}
	if got.next != [2]uintptr{got.addresses[1], got.addresses[2]} {
		t.Errorf("pending-list links changed: %x, addresses %x", got.next, got.addresses)
	}
	if unsafe.Sizeof(uintptr(0)) == 8 && runtime.GOOS != "windows" && got.addresses[0] <= 0xffffffff {
		t.Fatalf("fixture object address = %#x, want above 4 GiB on this 64-bit host", got.addresses[0])
	}
}
