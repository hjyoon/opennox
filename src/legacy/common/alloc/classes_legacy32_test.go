package alloc

import (
	"testing"

	"github.com/opennox/opennox/v1/legacy/common/alloc/handles"
)

func TestClassLegacy32ResolvesHandleAndActiveObject(t *testing.T) {
	handles.Init()
	t.Cleanup(handles.Release)

	class := NewClass("Legacy32Resolver", 12, 2)
	t.Cleanup(class.Free)
	handle := class.UPtr()
	handleToken := uint32(uintptr(handle))
	if got := AsClassLegacy32(handleToken); got != class {
		t.Fatalf("class for token %#x = %p, want %p", handleToken, got, class)
	}

	first := class.NewObject()
	second := class.NewObject()
	if first == nil || second == nil {
		t.Fatal("class did not allocate both static objects")
	}
	firstToken := uint32(uintptr(first))
	secondToken := uint32(uintptr(second))
	if got := class.ActivePointerLegacy32(firstToken); got != first {
		t.Fatalf("first object for token %#x = %p, want %p", firstToken, got, first)
	}
	if got := class.ActivePointerLegacy32(secondToken); got != second {
		t.Fatalf("second pointer for token %#x = %p, want %p", secondToken, got, second)
	}

	class.FreeObjectFirst(first)
	if got := class.ActivePointerLegacy32(firstToken); got != nil {
		t.Fatalf("inactive pointer for token %#x = %p, want nil", firstToken, got)
	}
	if got := AsClassLegacy32(handleToken ^ 0x80000000); got != nil {
		t.Fatalf("unknown class token resolved to %p", got)
	}
}
