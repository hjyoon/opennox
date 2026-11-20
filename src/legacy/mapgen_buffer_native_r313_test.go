package legacy

import (
	"slices"
	"testing"
	"unsafe"
)

func TestMapgenBufferNativeR313(t *testing.T) {
	defer mapgenHighFixtureMemory()()
	defer nox_xxx_mapgenFreeBuffer_521400()
	allocate, release := game4SupportAllocationService, game4SupportReleaseService
	failedAllocation := false
	var successful, released int
	game4SupportAllocationService = func(count, size uintptr) unsafe.Pointer {
		if count != 1 || size != mapgenBufferSize {
			t.Errorf("scratch allocation %d x %d", count, size)
		}
		if failedAllocation {
			return nil
		}
		ptr := allocate(count, size)
		if ptr != nil {
			successful++
		}
		return ptr
	}
	game4SupportReleaseService = func(ptr unsafe.Pointer) {
		if ptr == nil {
			t.Error("scratch release received a nil address")
		} else {
			released++
		}
		release(ptr)
	}

	requireClean := func() {
		t.Helper()
		if mapgenBuffer != nil || successful != released {
			t.Fatalf("scratch retained address %p with %d allocations and %d releases", mapgenBuffer, successful, released)
		}
	}
	nox_xxx_mapgenFreeBuffer_521400()
	requireClean()
	failedAllocation = true
	if nox_xxx_mapgenAllocBuffer_5213E0() != 0 {
		t.Fatal("reported a successful scratch allocation after allocation failure")
	}
	requireClean()
	failedAllocation = false
	if buffer, token := mapgenAcquireBuffer(func(unsafe.Pointer) uint32 { return 0 }); buffer != nil || token != 0 {
		t.Fatalf("failed registration returned %p with token %#x", buffer, token)
	}
	requireClean()
	if successful != 1 {
		t.Fatalf("registration failure performed %d physical allocations", successful)
	}

	zeroed := make([]byte, mapgenBufferSize)
	for iteration := 0; iteration < 12; iteration++ {
		if nox_xxx_mapgenAllocBuffer_5213E0() != 1 || mapgenBuffer == nil {
			t.Fatalf("scratch allocation failed at cycle %d", iteration+1)
		}
		if unsafe.Sizeof(uintptr(0)) == 8 && uint64(uintptr(mapgenBuffer)) <= 0xffffffff {
			t.Fatalf("64-bit scratch address lost its high bits at cycle %d", iteration+1)
		}
		data := unsafe.Slice((*byte)(mapgenBuffer), mapgenBufferSize)
		if !slices.Equal(data, zeroed) {
			t.Fatalf("scratch buffer was not zeroed at cycle %d", iteration+1)
		}
		for i := range data {
			data[i] = 0xa5
		}
		nox_xxx_mapgenFreeBuffer_521400()
		nox_xxx_mapgenFreeBuffer_521400()
		requireClean()
	}

	if nox_xxx_mapgenAllocBuffer_5213E0() != 1 {
		t.Fatal("initial scratch allocation failed")
	}
	if nox_xxx_mapgenAllocBuffer_5213E0() != 1 || successful != released+1 {
		t.Fatal("replacement scratch allocation did not release the previous allocation")
	}
	failedAllocation = true
	if nox_xxx_mapgenAllocBuffer_5213E0() != 0 {
		t.Fatal("replacement failure reported success")
	}
	requireClean()
	t.Log("Mapgen buffer native R313: 12 repeated lifetimes and allocation failures PASS")
}
