package legacy

import (
	"testing"
	"unsafe"

	"github.com/opennox/libs/types"
	"github.com/opennox/opennox/v1/server"
)

func TestMapPushUnitsAround52E040UsesNativeDispatch(t *testing.T) {
	origin := types.Ptf(123.5, 456.25)
	source := &server.Object{}
	if unsafe.Sizeof(uintptr(0)) > 4 && uintptr(unsafe.Pointer(source)) <= uintptr(^uint32(0)) {
		t.Skip("test object was not allocated above the PE32 address range")
	}

	old := mapPushUnitsAroundCall52E040
	t.Cleanup(func() { mapPushUnitsAroundCall52E040 = old })
	called := false
	mapPushUnitsAroundCall52E040 = func(gotOrigin types.Pointf, outerRadius, innerRadius, force float32, gotSource *server.Object, callback, callbackArg int) {
		called = true
		if gotOrigin != origin || outerRadius != 100 || innerRadius != 30 || force != 60 {
			t.Fatalf("push args = (%v, %v, %v, %v)", gotOrigin, outerRadius, innerRadius, force)
		}
		if gotSource != source {
			t.Fatalf("source pointer = %p, want %p", gotSource, source)
		}
		if callback != 7 || callbackArg != 9 {
			t.Fatalf("callback args = (%d, %d), want (7, 9)", callback, callbackArg)
		}
	}

	Nox_xxx_mapPushUnitsAround_52E040(origin, 100, 30, 60, source, 7, 9)
	if !called {
		t.Fatal("native radial push binding was not called")
	}
}

func TestMapPushUnitsAround52E040LegacyEntryUsesNativeDispatch(t *testing.T) {
	if unsafe.Sizeof(uintptr(0)) <= 4 {
		t.Skip("the PE32 C entry remains authoritative on 32-bit hosts")
	}
	origin := types.Ptf(321.5, 654.25)

	old := mapPushUnitsAroundCall52E040
	t.Cleanup(func() { mapPushUnitsAroundCall52E040 = old })
	called := false
	mapPushUnitsAroundCall52E040 = func(gotOrigin types.Pointf, outerRadius, innerRadius, force float32, gotSource *server.Object, callback, callbackArg int) {
		called = true
		if gotOrigin != origin || outerRadius != 96 || innerRadius != 10 || force != 100 {
			t.Fatalf("push args = (%v, %v, %v, %v)", gotOrigin, outerRadius, innerRadius, force)
		}
		if gotSource != nil {
			t.Fatalf("source pointer = %p, want nil", gotSource)
		}
		if callback != 11 || callbackArg != 13 {
			t.Fatalf("callback args = (%d, %d), want (11, 13)", callback, callbackArg)
		}
	}

	mapPushUnitsAroundLegacyEntry52E040(origin, 96, 10, 100, nil, 11, 13)
	if !called {
		t.Fatal("legacy C entry did not route to the native radial push binding")
	}
}
