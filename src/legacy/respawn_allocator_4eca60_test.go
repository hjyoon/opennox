package legacy

import (
	"testing"
	"unsafe"

	"github.com/opennox/opennox/v1/legacy/common/alloc"
	"github.com/opennox/opennox/v1/legacy/common/alloc/handles"
)

func TestRespawnAllocatorNative4ECA60LayoutAndResult(t *testing.T) {
	wantSize := uintptr(104)
	if unsafe.Sizeof(uintptr(0)) == 4 {
		wantSize = 60
	}
	sentinel := byte(0)
	tests := []struct {
		name      string
		allocator unsafe.Pointer
		want      int
	}{
		{name: "success", allocator: unsafe.Pointer(&sentinel), want: 1},
		{name: "failure", allocator: nil, want: 0},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			stored := unsafe.Pointer(&sentinel)
			got := respawnAllocatorNative4ECA60(respawnAllocatorRuntime4ECA60{
				NewClass: func(name string, recordSize uintptr, capacity int) unsafe.Pointer {
					if name != "Respawn" || recordSize != wantSize || capacity != 384 {
						t.Fatalf("allocation request = (%q, %d, %d), want (Respawn, %d, 384)", name, recordSize, capacity, wantSize)
					}
					return tc.allocator
				},
				StoreAllocator: func(allocator unsafe.Pointer) {
					stored = allocator
				},
			})
			if got != tc.want {
				t.Fatalf("result = %d, want %d", got, tc.want)
			}
			if stored != tc.allocator {
				t.Fatalf("stored allocator = %p, want %p", stored, tc.allocator)
			}
		})
	}
}

func TestNoxXxxAllocItemRespawnArray4ECA60NativeStorage(t *testing.T) {
	handles.Init()
	defer handles.Release()

	oldAllocator := respawnAllocatorLoad4ECA60()
	var allocator unsafe.Pointer
	defer func() {
		respawnAllocatorStore4ECA60(oldAllocator)
		alloc.AsClass(allocator).Free()
	}()

	if got := Nox_xxx_allocItemRespawnArray_4ECA60(); got != 1 {
		t.Fatalf("result = %d, want 1", got)
	}
	allocator = respawnAllocatorLoad4ECA60()
	if allocator == nil {
		t.Fatal("stored allocator is nil")
	}
	class := alloc.AsClass(allocator)
	if class == nil {
		t.Fatalf("stored allocator %p is not a live allocation class", allocator)
	}
	record := class.NewObject()
	if record == nil {
		t.Fatal("new Respawn allocation class did not provide a record")
	}
}
