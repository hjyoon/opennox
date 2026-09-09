package legacy

import (
	"math"
	"runtime"
	"testing"
	"unsafe"

	"github.com/opennox/opennox/v1/server"
)

func TestMonsterWanderExport512930PreservesNativePointer(t *testing.T) {
	if unsafe.Sizeof(uintptr(0)) != 8 {
		t.Skip("native-width routing regression applies to 64-bit builds")
	}

	unit := new(server.Object)
	if uintptr(unsafe.Pointer(unit)) <= math.MaxUint32 {
		t.Fatalf("unit pointer = %p, want address above the ABI32 range", unit)
	}

	oldCall := monsterWanderCall512930
	t.Cleanup(func() { monsterWanderCall512930 = oldCall })
	got := make([]*server.Object, 0, 2)
	monsterWanderCall512930 = func(unit *server.Object) {
		got = append(got, unit)
	}

	monsterWanderExportCall512930(unit)
	monsterWanderExportCall512930(nil)
	if len(got) != 2 {
		t.Fatalf("export calls = %d, want 2", len(got))
	}
	if got[0] != unit {
		t.Fatalf("unit = %p, want %p", got[0], unit)
	}
	if got[1] != nil {
		t.Fatalf("null unit = %p, want nil", got[1])
	}
	runtime.KeepAlive(unit)
}

func TestMonsterWanderGoWrapper512930UsesNativeAdapter(t *testing.T) {
	unit := new(server.Object)
	oldCall := monsterWanderCall512930
	t.Cleanup(func() { monsterWanderCall512930 = oldCall })

	var got *server.Object
	monsterWanderCall512930 = func(unit *server.Object) { got = unit }
	Nox_xxx_scriptMonsterRoam_512930(unit)
	if got != unit {
		t.Fatalf("unit = %p, want %p", got, unit)
	}
}
