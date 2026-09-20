package legacy

import (
	"testing"
	"unsafe"

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
	if unsafe.Sizeof(uintptr(0)) <= 4 {
		t.Skip("the PE32 C entry remains authoritative on 32-bit hosts")
	}
	old := makeScorchCall537AF0
	t.Cleanup(func() { makeScorchCall537AF0 = old })

	wantPos := types.Ptf(321.5, 654.25)
	called := false
	makeScorchCall537AF0 = func(pos types.Pointf, kind int) {
		called = true
		if pos != wantPos || kind != 1 {
			t.Fatalf("scorch args = (%v, %d), want (%v, 1)", pos, kind, wantPos)
		}
	}
	makeScorchLegacyEntry537AF0(wantPos, 1)
	if !called {
		t.Fatal("legacy C scorch entry did not route to the native binding")
	}
}
