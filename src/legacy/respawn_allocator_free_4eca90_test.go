package legacy

import (
	"reflect"
	"testing"
	"unsafe"

	"github.com/opennox/opennox/v1/legacy/common/alloc"
	"github.com/opennox/opennox/v1/legacy/common/alloc/handles"
)

func TestRespawnAllocatorFreeNative4ECA90OrderAndCachedPointer(t *testing.T) {
	originalValue := byte(1)
	replacementValue := byte(2)
	original := unsafe.Pointer(&originalValue)
	replacement := unsafe.Pointer(&replacementValue)
	allocator := original
	loads := 0
	var events []string

	respawnAllocatorFreeNative4ECA90(respawnAllocatorFreeRuntime4ECA90{
		LoadAllocator: func() unsafe.Pointer {
			loads++
			events = append(events, "load-allocator")
			return allocator
		},
		FreeClass: func(got unsafe.Pointer) {
			events = append(events, "free-class")
			allocator = replacement
			if got != original {
				t.Fatalf("free-class allocator = %p, want cached %p", got, original)
			}
		},
	})

	if loads != 1 {
		t.Fatalf("allocator loads = %d, want 1", loads)
	}
	if allocator != replacement {
		t.Fatalf("allocator after free callback = %p, want %p", allocator, replacement)
	}
	if want := []string{"load-allocator", "free-class"}; !reflect.DeepEqual(events, want) {
		t.Fatalf("events = %#v, want %#v", events, want)
	}
}

func TestSub4ECA90NativeStorageAndFree(t *testing.T) {
	handles.Init()
	defer handles.Release()

	oldAllocator := respawnAllocatorLoad4ECA60()
	defer respawnAllocatorStore4ECA60(oldAllocator)

	respawnAllocatorStore4ECA60(nil)
	Sub_4ECA90()
	if got := respawnAllocatorLoad4ECA60(); got != nil {
		t.Fatalf("nil allocator global after free = %p, want nil", got)
	}

	class := alloc.NewClass("RespawnFree4ECA90Test", 16, 2)
	allocator := class.UPtr()
	respawnAllocatorStore4ECA60(allocator)
	Sub_4ECA90()

	if got := respawnAllocatorLoad4ECA60(); got != allocator {
		t.Fatalf("allocator global after free = %p, want unchanged %p", got, allocator)
	}
	if got := alloc.AsClass(allocator); got != nil {
		t.Fatalf("freed allocator still resolves to %v", got)
	}
}
