package legacy

import (
	"math"
	"runtime"
	"testing"
	"unsafe"

	"github.com/opennox/opennox/v1/server"
)

func TestSummonedCreatureLimitExport500D70PreservesNativePointerAndInt32(t *testing.T) {
	owner := new(server.Object)
	if unsafe.Sizeof(uintptr(0)) == 8 && uintptr(unsafe.Pointer(owner)) <= math.MaxUint32 {
		t.Fatalf("owner pointer = %p, want native address above 4 GiB", owner)
	}

	oldCall := summonedCreatureLimitCall500D70
	t.Cleanup(func() { summonedCreatureLimitCall500D70 = oldCall })
	var gotOwner *server.Object
	var gotIndex int32
	nextResult := true
	summonedCreatureLimitCall500D70 = func(owner *server.Object, guideIndex int32) bool {
		gotOwner = owner
		gotIndex = guideIndex
		return nextResult
	}

	var pin runtime.Pinner
	pin.Pin(owner)
	defer pin.Unpin()
	if got := summonedCreatureLimitExportCall500D70(owner, math.MinInt32); got != 1 {
		t.Fatalf("true result = %d, want canonical one", got)
	}
	if gotOwner != owner || gotIndex != math.MinInt32 {
		t.Fatalf("call = (%p, %d), want (%p, %d)", gotOwner, gotIndex, owner, int32(math.MinInt32))
	}

	nextResult = false
	if got := summonedCreatureLimitExportCall500D70(nil, math.MaxInt32); got != 0 {
		t.Fatalf("false result = %d, want canonical zero", got)
	}
	if gotOwner != nil || gotIndex != math.MaxInt32 {
		t.Fatalf("nil-owner call = (%p, %d), want (nil, %d)", gotOwner, gotIndex, int32(math.MaxInt32))
	}
	runtime.KeepAlive(owner)
}
