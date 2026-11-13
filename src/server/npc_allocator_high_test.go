package server

import (
	"runtime"
	"testing"
	"unsafe"

	"github.com/opennox/opennox/v1/legacy/common/alloc"
)

func TestNPCNativeAllocatorPolicies(t *testing.T) {
	var ordinary, high serverNPCs
	ordinary.Init()
	high.InitHigh()
	t.Cleanup(func() {
		alloc.FreeSlice(ordinary.arr)
		alloc.FreeSlice(high.arr)
		ordinary.arr, high.arr = nil, nil
	})
	if ordinary.highAddresses || !high.highAddresses {
		t.Fatal("NPC allocation policy crossed owner boundaries")
	}
	a, b := ordinary.New(19), high.New(37)
	if a == nil || b == nil || a == b {
		t.Fatal("NPC owners did not allocate distinct records")
	}
	requireExplicitHighServerPointer(t, unsafe.Pointer(b))
	a.Color8[0], b.Color8[0] = 17, 29
	runtime.GC()
	if ordinary.ByID(19) != a || high.ByID(37) != b || a.Color8[0] != 17 || b.Color8[0] != 29 {
		t.Fatal("NPC owner lookup or storage changed across GC")
	}
	b.LiveVal = 0
	reused := high.New(41)
	if reused != b || reused.LiveVal != 1 || reused.IDVal != 41 || reused.Color8 != [6]uint32{} || a.Color8[0] != 17 {
		t.Fatal("NPC reuse failed to clear its record or modified another owner")
	}
	if high.ByID(37) != nil || high.ByID(41) != b || ordinary.ByID(19) != a {
		t.Fatal("NPC owner lookup retained the previous record identity")
	}
}
