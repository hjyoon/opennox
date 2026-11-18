package legacy

import (
	"testing"
	"unsafe"

	"github.com/opennox/libs/things"
	"github.com/opennox/libs/types"
)

func TestMakeScorch537AF0UsesNativeDispatch(t *testing.T) {
	old := makeScorchCall537AF0
	t.Cleanup(func() { makeScorchCall537AF0 = old })

	wantPos := types.Ptf(123.5, 456.25)
	called := false
	makeScorchCall537AF0 = func(pos types.Pointf, kind int) {
		called = true
		if pos != wantPos || kind != 2 {
			t.Fatalf("scorch args = (%v, %d), want (%v, 2)", pos, kind, wantPos)
		}
	}
	Nox_xxx_sMakeScorch_537AF0(wantPos, 2)
	if !called {
		t.Fatal("native scorch binding was not called")
	}
}

func TestMakeScorch537AF0LegacyEntryUsesNativeDispatch(t *testing.T) {
	old := makeScorchCall537AF0
	t.Cleanup(func() { makeScorchCall537AF0 = old })

	wantPos := types.Ptf(321.5, 654.25)
	pos := game4ConsumerValue(t, wantPos)
	if unsafe.Sizeof(uintptr(0)) > 4 && uint64(uintptr(unsafe.Pointer(pos))) <= 0xffffffff {
		t.Fatalf("position pointer is below 4 GiB: %p", pos)
	}
	called := 0
	makeScorchCall537AF0 = func(pos types.Pointf, kind int) {
		called++
		if pos != wantPos || kind != 1 {
			t.Fatalf("scorch args = (%v, %d), want (%v, 1)", pos, kind, wantPos)
		}
	}
	makeScorchLegacyEntry537AF0(pos, 1, false)
	makeScorchLegacyEntry537AF0(pos, 1, true)
	makeScorchLegacyEntry537AF0(nil, 1, true)
	if called != 2 {
		t.Fatalf("direct and indirect C dispatch count = %d, want 2", called)
	}
}

func TestScorchInit537BD0LegacyEntry(t *testing.T) {
	s := playerVitalReportWorld(t)
	s.FreeObjectTypes()
	t.Cleanup(s.FreeObjectTypes)
	for _, name := range []string{"ScorchMarkFloorSmallA", "ScorchMarkFloorMediumA", "ScorchMarkFloorLargeB"} {
		if err := s.Types.ReadObjectType(&things.Thing{Name: name}); err != nil {
			t.Fatal(err)
		}
	}
	want := s.Types.IndByID("ScorchMarkFloorLargeB")
	if got := scorchInitLegacyEntry537BD0(); want == 0 || got != want {
		t.Fatalf("initializer last type = %d, want %d", got, want)
	}
}
