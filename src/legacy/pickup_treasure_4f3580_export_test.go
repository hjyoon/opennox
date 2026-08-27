package legacy

import (
	"math"
	"runtime"
	"testing"
	"unsafe"

	"github.com/opennox/opennox/v1/legacy/common/alloc"
	"github.com/opennox/opennox/v1/server"
)

func TestPickupTreasureExport4F3580PreservesNativePointersFourArgumentsAndResult(t *testing.T) {
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

	old := Nox_xxx_pickupTreasure_4F3580
	t.Cleanup(func() { Nox_xxx_pickupTreasure_4F3580 = old })
	calls := 0
	Nox_xxx_pickupTreasure_4F3580 = func(gotOwner, gotItem *server.Object, arg3, arg4 int32) int32 {
		calls++
		if gotOwner != owner || gotItem != item {
			t.Fatalf("objects = %p/%p, want %p/%p", gotOwner, gotItem, owner, item)
		}
		switch calls {
		case 1:
			if arg3 != math.MinInt32 || arg4 != math.MaxInt32 {
				t.Fatalf("first scalars = %d/%d", arg3, arg4)
			}
			return math.MinInt32
		case 2:
			if arg3 != -17 || arg4 != -23 {
				t.Fatalf("second scalars = %d/%d", arg3, arg4)
			}
			return math.MaxInt32
		default:
			t.Fatalf("unexpected call %d", calls)
			return 0
		}
	}

	if got := pickupTreasureExportCall4F3580(owner, item, math.MinInt32, math.MaxInt32); got != math.MinInt32 {
		t.Fatalf("first result = %d, want %d", got, int32(math.MinInt32))
	}
	if got := pickupTreasureExportCall4F3580(owner, item, -17, -23); got != math.MaxInt32 {
		t.Fatalf("second result = %d, want %d", got, int32(math.MaxInt32))
	}
	if calls != 2 {
		t.Fatalf("calls = %d, want 2", calls)
	}
	runtime.KeepAlive(owner)
	runtime.KeepAlive(item)
}
