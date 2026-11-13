package alloc

import (
	"runtime"
	"testing"
	"unsafe"
)

func TestTryFreePtrRetainsForeignOwnership(t *testing.T) {
	unknown := new(uint64)
	*unknown = 0x1122334455667788
	if TryFreePtr(nil) || TryFreePtr(unsafe.Pointer(unknown)) {
		t.Fatal("unregistered memory was accepted for release")
	}
	if *unknown != 0x1122334455667788 {
		t.Fatal("unregistered memory was modified")
	}
	runtime.KeepAlive(unknown)
	for _, high := range []bool{false, true} {
		t.Run(map[bool]string{false: "ordinary", true: "high"}[high], func(t *testing.T) {
			var p unsafe.Pointer
			if high {
				p, _ = MallocHigh(19)
			} else {
				p, _ = Malloc(19)
			}
			if !TryFreePtr(p) {
				t.Fatal("registered owner was not released")
			}
			if TryFreePtr(p) {
				t.Fatal("owner was released twice")
			}
			allocMu.Lock()
			_, retained := allocs[p]
			_, retainedRelease := allocReleases[p]
			allocMu.Unlock()
			if retained || retainedRelease {
				t.Fatal("released ownership remained registered")
			}
		})
	}
}

func TestCStringHighInternOwnershipIsolation(t *testing.T) {
	const text = "native UTF8 한글"
	ordinary := InternCString(text)
	high := InternCStringHigh(text)
	if ordinary == high || ordinary != InternCString(text) || high != InternCStringHigh(text) {
		t.Fatal("ordinary and high intern owners lost independent stable identities")
	}
	if GoString(ordinary) != text || GoString(high) != text {
		t.Fatal("native interned text changed")
	}
	requireFixtureAddress(t, unsafe.Pointer(high))
	FreeStrings()
	allocMu.Lock()
	_, ordinaryRetained := allocs[unsafe.Pointer(ordinary)]
	_, highRetained := allocs[unsafe.Pointer(high)]
	allocMu.Unlock()
	if ordinaryRetained || highRetained {
		t.Fatal("intern cleanup retained a foreign owner")
	}
}
