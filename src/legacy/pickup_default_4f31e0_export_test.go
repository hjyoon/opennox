package legacy

import (
	"math"
	"runtime"
	"testing"
	"unsafe"

	"github.com/opennox/opennox/v1/legacy/common/alloc"
	"github.com/opennox/opennox/v1/server"
)

func TestPickupDefaultExport4F31E0PreservesPointersAndFourArguments(t *testing.T) {
	owner, freeOwner := alloc.New(server.Object{})
	defer freeOwner()
	item, freeItem := alloc.New(server.Object{})
	defer freeItem()

	if unsafe.Sizeof(uintptr(0)) == 8 {
		for index, pointer := range []unsafe.Pointer{unsafe.Pointer(owner), unsafe.Pointer(item)} {
			if uintptr(pointer) <= math.MaxUint32 {
				t.Fatalf("pointer %d = %p, want native high address", index, pointer)
			}
		}
	}

	old := Nox_xxx_pickupDefault_4F31E0
	t.Cleanup(func() { Nox_xxx_pickupDefault_4F31E0 = old })
	calls := 0
	Nox_xxx_pickupDefault_4F31E0 = func(gotOwner, gotItem *server.Object, report, ignored int) bool {
		calls++
		if gotOwner != owner || gotItem != item {
			t.Fatalf("objects = %p/%p, want %p/%p", gotOwner, gotItem, owner, item)
		}
		switch calls {
		case 1:
			if report != math.MinInt32 || ignored != math.MaxInt32 {
				t.Fatalf("first scalars = %d/%d", report, ignored)
			}
			return true
		case 2:
			if report != -17 || ignored != -23 {
				t.Fatalf("second scalars = %d/%d", report, ignored)
			}
			return false
		default:
			t.Fatalf("unexpected call %d", calls)
			return false
		}
	}

	if got := pickupDefaultExportCall4F31E0(owner, item, math.MinInt32, math.MaxInt32); got != 1 {
		t.Fatalf("first result = %d, want 1", got)
	}
	if got := pickupDefaultExportCall4F31E0(owner, item, -17, -23); got != 0 {
		t.Fatalf("second result = %d, want 0", got)
	}
	if calls != 2 {
		t.Fatalf("calls = %d, want 2", calls)
	}
	runtime.KeepAlive(owner)
	runtime.KeepAlive(item)
}
