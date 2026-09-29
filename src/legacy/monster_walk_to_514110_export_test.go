package legacy

import (
	"math"
	"runtime"
	"testing"
	"unsafe"

	"github.com/opennox/opennox/v1/server"
)

func TestMonsterWalkToExport514110PreservesNativePointerAndCoordinates(t *testing.T) {
	if unsafe.Sizeof(uintptr(0)) != 8 {
		t.Skip("native-width routing regression applies to 64-bit builds")
	}

	unit := new(server.Object)
	if uintptr(unsafe.Pointer(unit)) <= math.MaxUint32 {
		t.Fatalf("unit pointer = %p, want address above the ABI32 range", unit)
	}

	oldCall := monsterWalkToCall514110
	t.Cleanup(func() { monsterWalkToCall514110 = oldCall })
	type call struct {
		unit *server.Object
		x, y float32
	}
	var got []call
	monsterWalkToCall514110 = func(unit *server.Object, x, y float32) {
		got = append(got, call{unit: unit, x: x, y: y})
	}

	x := math.Float32frombits(0x45a87000)
	y := math.Float32frombits(0xc5430000)
	monsterWalkToExportCall514110(unit, x, y)
	if len(got) != 1 {
		t.Fatalf("export calls = %d, want 1", len(got))
	}
	if got[0].unit != unit {
		t.Fatalf("unit = %p, want %p", got[0].unit, unit)
	}
	if math.Float32bits(got[0].x) != math.Float32bits(x) || math.Float32bits(got[0].y) != math.Float32bits(y) {
		t.Fatalf("coordinates = (%08x, %08x), want (%08x, %08x)",
			math.Float32bits(got[0].x), math.Float32bits(got[0].y), math.Float32bits(x), math.Float32bits(y))
	}
	runtime.KeepAlive(unit)
}

func TestMonsterWalkToGoWrapper514110UsesNativeAdapter(t *testing.T) {
	unit := new(server.Object)
	oldCall := monsterWalkToCall514110
	t.Cleanup(func() { monsterWalkToCall514110 = oldCall })

	var gotUnit *server.Object
	var gotX, gotY float32
	monsterWalkToCall514110 = func(unit *server.Object, x, y float32) {
		gotUnit, gotX, gotY = unit, x, y
	}
	Nox_xxx_monsterWalkTo_514110(unit, 8640, 3120)
	if gotUnit != unit || gotX != 8640 || gotY != 3120 {
		t.Fatalf("call = (%p, %v, %v), want (%p, 8640, 3120)", gotUnit, gotX, gotY, unit)
	}
}
