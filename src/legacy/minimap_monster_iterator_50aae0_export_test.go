package legacy

import (
	"math"
	"runtime"
	"testing"
	"unsafe"

	"github.com/opennox/libs/types"

	"github.com/opennox/opennox/v1/server"
)

func TestMinimapMonsterIterator50AAE0CABIUsesNativePointer(t *testing.T) {
	unit := &server.Object{PosVec: types.Ptf(12.5, 34.75)}
	var pin runtime.Pinner
	pin.Pin(unit)
	defer pin.Unpin()
	if unsafe.Sizeof(uintptr(0)) == 8 && uintptr(unsafe.Pointer(unit)) <= math.MaxUint32 {
		t.Fatalf("unit pointer = %p, want native address above 4 GiB", unit)
	}

	oldFirst := minimapMonsterFirstImpl50AAE0
	oldNext := minimapMonsterNextImpl50AB10
	minimapMonsterFirstImpl50AAE0 = func() *server.Object { return unit }
	minimapMonsterNextImpl50AB10 = func() *server.Object { return nil }
	t.Cleanup(func() {
		minimapMonsterFirstImpl50AAE0 = oldFirst
		minimapMonsterNextImpl50AB10 = oldNext
	})

	got := minimapMonsterFirstExportCall50AAE0()
	want := unsafe.Pointer(&unit.PosVec)
	if got != want {
		t.Fatalf("C First position = %p, want native field %p", got, want)
	}
	pos := (*types.Pointf)(got)
	if *pos != unit.PosVec {
		t.Fatalf("C First position = %+v, want %+v", *pos, unit.PosVec)
	}
	if got := minimapMonsterNextExportCall50AB10(); got != nil {
		t.Fatalf("C Next = %p, want nil", got)
	}
	runtime.KeepAlive(unit)
}
