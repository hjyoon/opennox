package legacy

import (
	"math"
	"runtime"
	"testing"
	"unsafe"

	"github.com/opennox/opennox/v1/server"
)

func TestUnitOrderExport533900PreservesNativePointersAndInt32Order(t *testing.T) {
	owner := new(server.Object)
	creature := new(server.Object)
	if unsafe.Sizeof(uintptr(0)) == 8 {
		if uintptr(unsafe.Pointer(owner)) <= math.MaxUint32 || uintptr(unsafe.Pointer(creature)) <= math.MaxUint32 {
			t.Fatalf("object pointers = %p/%p, want native addresses above 4 GiB", owner, creature)
		}
	}
	var pinner runtime.Pinner
	pinner.Pin(owner)
	pinner.Pin(creature)
	defer pinner.Unpin()

	oldCall := unitOrderCall533900
	t.Cleanup(func() { unitOrderCall533900 = oldCall })
	type call struct {
		owner, creature *server.Object
		order           uint32
	}
	var calls []call
	unitOrderCall533900 = func(owner, creature *server.Object, order uint32) {
		calls = append(calls, call{owner: owner, creature: creature, order: order})
	}

	unitOrderExportCall533900(owner, creature, math.MinInt32)
	Nox_xxx_orderUnit_533900(owner, nil, math.MaxUint32)
	unitOrderExportCall533900(nil, nil, -1)
	if len(calls) != 3 {
		t.Fatalf("calls = %d, want 3", len(calls))
	}
	if calls[0] != (call{owner: owner, creature: creature, order: 0x80000000}) {
		t.Fatalf("C export call = %#v", calls[0])
	}
	if calls[1] != (call{owner: owner, order: math.MaxUint32}) {
		t.Fatalf("Go wrapper call = %#v", calls[1])
	}
	if calls[2] != (call{order: math.MaxUint32}) {
		t.Fatalf("null export call = %#v", calls[2])
	}
	runtime.KeepAlive(owner)
	runtime.KeepAlive(creature)
}
