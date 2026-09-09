package legacy

import (
	"math"
	"runtime"
	"testing"
	"unsafe"

	"github.com/opennox/opennox/v1/server"
)

func TestControlledCreatureCountExport500D10PreservesNativePointerAndInt32(t *testing.T) {
	owner := new(server.Object)
	if unsafe.Sizeof(uintptr(0)) == 8 && uintptr(unsafe.Pointer(owner)) <= math.MaxUint32 {
		t.Fatalf("owner pointer = %p, want native address above 4 GiB", owner)
	}

	oldCall := controlledCreatureCountCall500D10
	t.Cleanup(func() { controlledCreatureCountCall500D10 = oldCall })
	var gotOwner *server.Object
	nextResult := int32(math.MinInt32 + 0x500d10)
	controlledCreatureCountCall500D10 = func(owner *server.Object) int32 {
		gotOwner = owner
		return nextResult
	}

	var pin runtime.Pinner
	pin.Pin(owner)
	defer pin.Unpin()
	if got := controlledCreatureCountExportCall500D10(owner); got != nextResult {
		t.Fatalf("result = %d, want %d", got, nextResult)
	}
	if gotOwner != owner {
		t.Fatalf("owner = %p, want native pointer %p", gotOwner, owner)
	}

	nextResult = math.MaxInt32
	if got := controlledCreatureCountExportCall500D10(nil); got != math.MaxInt32 {
		t.Fatalf("nil-owner result = %d, want %d", got, int32(math.MaxInt32))
	}
	if gotOwner != nil {
		t.Fatalf("nil-owner call received %p", gotOwner)
	}
	runtime.KeepAlive(owner)
}
