package legacy

import (
	"math"
	"runtime"
	"testing"
	"unsafe"

	"github.com/opennox/opennox/v1/legacy/common/alloc"
	"github.com/opennox/opennox/v1/server"
)

func TestUnitBanishExport5017F0PreservesNativePointer(t *testing.T) {
	unit, freeUnit := alloc.New(server.Object{})
	defer freeUnit()
	if unsafe.Sizeof(uintptr(0)) == 8 && uintptr(unsafe.Pointer(unit)) <= math.MaxUint32 {
		t.Fatalf("unit pointer = %p, want native address above 4 GiB", unit)
	}

	oldCall := unitBanishCall5017F0
	t.Cleanup(func() { unitBanishCall5017F0 = oldCall })
	var calls []*server.Object
	unitBanishCall5017F0 = func(object *server.Object) {
		calls = append(calls, object)
	}

	unitBanishExportCall5017F0(unit)
	unitBanishExportCall5017F0(nil)
	Nox_xxx_banishUnit_5017F0(unit)
	if len(calls) != 3 || calls[0] != unit || calls[1] != nil || calls[2] != unit {
		t.Fatalf("calls = %v, want [%p nil %p]", calls, unit, unit)
	}
	runtime.KeepAlive(unit)
}
