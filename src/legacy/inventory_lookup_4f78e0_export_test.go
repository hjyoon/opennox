package legacy

import (
	"math"
	"runtime"
	"testing"
	"unsafe"

	"github.com/opennox/opennox/v1/legacy/common/alloc"
	"github.com/opennox/opennox/v1/server"
)

func TestInventoryLookupExports4F78E0PreserveHighNativePointers(t *testing.T) {
	holder, freeHolder := alloc.New(server.Object{})
	defer freeHolder()
	first, freeFirst := alloc.New(server.Object{})
	defer freeFirst()
	matched, freeMatched := alloc.New(server.Object{})
	defer freeMatched()
	duplicate, freeDuplicate := alloc.New(server.Object{})
	defer freeDuplicate()

	first.NetCode = 0x0000ffff
	matched.NetCode = math.MaxUint32
	duplicate.NetCode = math.MaxUint32
	matched.InvHolder = holder
	duplicate.InvHolder = holder
	holder.InvFirstItem = first
	first.InvNextItem = matched
	matched.InvNextItem = duplicate

	if unsafe.Sizeof(uintptr(0)) == 8 {
		for index, pointer := range []unsafe.Pointer{
			unsafe.Pointer(holder),
			unsafe.Pointer(first),
			unsafe.Pointer(matched),
			unsafe.Pointer(duplicate),
		} {
			if uintptr(pointer) <= math.MaxUint32 {
				t.Fatalf("pointer %d = %p, want native high address", index, pointer)
			}
		}
	}

	if got := inventoryContainsExportCall4F78E0(holder, matched); got != 1 {
		t.Fatalf("contains result = %d, want 1", got)
	}
	if got := equippedItemByCodeExportCall4F7920(holder, math.MaxUint32); got != matched {
		t.Fatalf("lookup result = %p, want first match %p", got, matched)
	}

	matched.InvHolder = nil
	if got := inventoryContainsExportCall4F78E0(holder, matched); got != 0 {
		t.Fatalf("contains result after holder detach = %d, want 0", got)
	}

	runtime.KeepAlive(holder)
	runtime.KeepAlive(first)
	runtime.KeepAlive(matched)
	runtime.KeepAlive(duplicate)
}
