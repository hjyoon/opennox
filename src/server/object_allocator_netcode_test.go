package server

import "testing"

func TestServerObjectAllocatorPreservesNativeNetCode(t *testing.T) {
	var objects serverObjects
	if !objects.Init(2) {
		t.Fatal("object allocator initialization failed")
	}
	defer objects.FreeObjects()

	first := objects.alloc.NewObject()
	if first == nil {
		t.Fatal("object allocator returned nil")
	}
	if first.NetCode == 0 {
		t.Fatal("first reused object lost its preassigned NetCode")
	}
	want := first.NetCode

	objects.alloc.FreeObjectFirst(first)
	second := objects.alloc.NewObject()
	if second != first {
		t.Fatalf("allocator returned %p after freeing %p", second, first)
	}
	if second.NetCode != want {
		t.Fatalf("reused object NetCode = %d, want %d", second.NetCode, want)
	}
}
