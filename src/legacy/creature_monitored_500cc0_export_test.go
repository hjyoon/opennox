package legacy

import (
	"math"
	"runtime"
	"testing"
	"unsafe"

	"github.com/opennox/opennox/v1/server"
)

func TestCreatureMonitoredExport500CC0PreservesNativePointersAndCanonicalResult(t *testing.T) {
	owner := new(server.Object)
	unit := new(server.Object)
	if unsafe.Sizeof(uintptr(0)) == 8 {
		if ptr := uintptr(unsafe.Pointer(owner)); ptr <= math.MaxUint32 {
			t.Fatalf("owner pointer = %#x, want native address above 4 GiB", ptr)
		}
		if ptr := uintptr(unsafe.Pointer(unit)); ptr <= math.MaxUint32 {
			t.Fatalf("unit pointer = %#x, want native address above 4 GiB", ptr)
		}
	}

	oldCall := creatureMonitoredCall500CC0
	t.Cleanup(func() { creatureMonitoredCall500CC0 = oldCall })
	type call struct {
		owner *server.Object
		unit  *server.Object
	}
	var calls []call
	result := true
	creatureMonitoredCall500CC0 = func(owner, unit *server.Object) bool {
		calls = append(calls, call{owner: owner, unit: unit})
		return result
	}

	var pin runtime.Pinner
	pin.Pin(owner)
	pin.Pin(unit)
	defer pin.Unpin()
	if got := creatureMonitoredExportCall500CC0(owner, unit); got != 1 {
		t.Fatalf("true result = %d, want canonical 1", got)
	}
	result = false
	if got := creatureMonitoredExportCall500CC0(nil, nil); got != 0 {
		t.Fatalf("false result = %d, want canonical 0", got)
	}
	if len(calls) != 2 || calls[0] != (call{owner: owner, unit: unit}) || calls[1] != (call{}) {
		t.Fatalf("calls = %#v, want native pair then two nil pointers", calls)
	}
	runtime.KeepAlive(owner)
	runtime.KeepAlive(unit)
}
