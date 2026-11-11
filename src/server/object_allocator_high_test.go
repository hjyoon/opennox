package server

import (
	"bytes"
	"runtime"
	"testing"
	"unsafe"

	"github.com/opennox/opennox/v1/legacy/common/alloc"
)

func TestServerObjectAllocatorHighNativeFrames(t *testing.T) {
	for _, high := range []bool{false, true} {
		name := "ordinary"
		if high {
			name = "high"
		}
		t.Run(name, func(t *testing.T) {
			var objects serverObjects
			initialize := objects.Init
			if high {
				initialize = objects.InitHigh
			}
			if !initialize(2) || objects.highAddresses != high {
				t.Fatal("object allocator did not keep its per-owner policy")
			}
			defer objects.FreeObjects()
			check := func(p unsafe.Pointer) {
				t.Helper()
				if p == nil || (high && runtime.GOOS == "windows" && unsafe.Sizeof(uintptr(0)) == 8 && uintptr(p) <= uintptr(^uint32(0))) {
					t.Fatalf("native frame address violates owner policy: %p", p)
				}
			}
			first := objects.alloc.NewObject()
			check(unsafe.Pointer(first))
			if first.NetCode == 0 {
				t.Fatal("initial class record lost its native NetCode")
			}
			code := first.NetCode
			objects.alloc.FreeObjectFirst(first)
			reused := objects.alloc.NewObject()
			if reused != first || reused.NetCode != code {
				t.Fatal("native class reuse changed identity or NetCode")
			}
			health := objects.newHealthData()
			defer alloc.Free(health)
			check(unsafe.Pointer(health))
			if *health != (HealthData{}) {
				t.Fatal("native health frame is not zero initialized")
			}
			data := objects.newNativeData(17)
			defer alloc.FreeSlice(data)
			check(unsafe.Pointer(&data[0]))
			if !bytes.Equal(data, make([]byte, 17)) {
				t.Fatal("native data frame is not zero initialized")
			}
			copy(data, []byte{1, 2, 3})
			if !bytes.Equal(data[:4], []byte{1, 2, 3, 0}) {
				t.Fatal("native data frame did not retain copied input")
			}
			memory := objects.newNativeMemory(23)
			defer alloc.FreePtr(memory)
			check(memory)
			if !bytes.Equal(unsafe.Slice((*byte)(memory), 23), make([]byte, 23)) {
				t.Fatal("native auxiliary frame is not zero initialized")
			}
		})
	}
}
