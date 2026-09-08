package legacy

import (
	"math"
	"runtime"
	"testing"
	"unsafe"

	"github.com/opennox/opennox/v1/server"
)

func TestUnitFollowExport5158C0PreservesBothNativePointers(t *testing.T) {
	if unsafe.Sizeof(uintptr(0)) != 8 {
		t.Skip("native-width routing regression applies to 64-bit builds")
	}

	unit := new(server.Object)
	target := new(server.Object)
	if uintptr(unsafe.Pointer(unit)) <= math.MaxUint32 {
		t.Fatalf("unit pointer = %p, want address above the ABI32 range", unit)
	}
	if uintptr(unsafe.Pointer(target)) <= math.MaxUint32 {
		t.Fatalf("target pointer = %p, want address above the ABI32 range", target)
	}

	oldCall := unitFollowCall5158C0
	t.Cleanup(func() { unitFollowCall5158C0 = oldCall })
	type call struct{ unit, target *server.Object }
	var got []call
	unitFollowCall5158C0 = func(unit, target *server.Object) {
		got = append(got, call{unit: unit, target: target})
	}

	unitFollowExportCall5158C0(unit, target)
	unitFollowExportCall5158C0(nil, target)
	unitFollowExportCall5158C0(unit, nil)
	unitFollowExportCall5158C0(nil, nil)
	want := []call{
		{unit: unit, target: target},
		{unit: nil, target: target},
		{unit: unit, target: nil},
		{unit: nil, target: nil},
	}
	if len(got) != len(want) {
		t.Fatalf("export calls = %d, want %d", len(got), len(want))
	}
	for i := range want {
		if got[i] != want[i] {
			t.Errorf("call %d = (%p, %p), want (%p, %p)", i, got[i].unit, got[i].target, want[i].unit, want[i].target)
		}
	}
	runtime.KeepAlive(unit)
	runtime.KeepAlive(target)
}

func TestUnitFollowGoWrapper5158C0UsesNativeAdapter(t *testing.T) {
	unit := new(server.Object)
	target := new(server.Object)
	oldCall := unitFollowCall5158C0
	t.Cleanup(func() { unitFollowCall5158C0 = oldCall })

	var gotUnit, gotTarget *server.Object
	unitFollowCall5158C0 = func(unit, target *server.Object) {
		gotUnit, gotTarget = unit, target
	}
	Nox_xxx_unitSetFollow_5158C0(unit, target)
	if gotUnit != unit || gotTarget != target {
		t.Fatalf("objects = (%p, %p), want (%p, %p)", gotUnit, gotTarget, unit, target)
	}
}
