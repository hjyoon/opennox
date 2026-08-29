package handles

import (
	"testing"
	"unsafe"
)

func TestOpaqueHandlesRemainUniqueAndPointerAlignedWhenInterleaved(t *testing.T) {
	Init()
	t.Cleanup(Release)

	firstNumber := New()
	firstPointer := NewPtr()
	secondNumber := New()
	secondPointer := NewPtr()

	values := []uintptr{
		firstNumber,
		uintptr(firstPointer),
		secondNumber,
		uintptr(secondPointer),
	}
	seen := make(map[uintptr]struct{}, len(values))
	for _, value := range values {
		if !IsValid(value) {
			t.Fatalf("handle %#x is outside the active arena", value)
		}
		if _, ok := seen[value]; ok {
			t.Fatalf("handle %#x was allocated more than once", value)
		}
		seen[value] = struct{}{}
	}

	align := uintptr(unsafe.Alignof(uintptr(0)))
	for name, pointer := range map[string]unsafe.Pointer{
		"first":  firstPointer,
		"second": secondPointer,
	} {
		if got := uintptr(pointer) % align; got != 0 {
			t.Fatalf("%s pointer handle alignment remainder = %d, want 0 (alignment %d)", name, got, align)
		}
	}
}

func TestInitRestartsOpaqueHandleArena(t *testing.T) {
	Init()
	_ = New()
	if got := cur; got != handleStride {
		t.Fatalf("cursor after first allocation = %d, want %d", got, handleStride)
	}
	Release()

	Init()
	t.Cleanup(Release)
	if got := cur; got != 0 {
		t.Fatalf("cursor after reinitialization = %d, want 0", got)
	}
	_ = New()
	if got := cur; got != handleStride {
		t.Fatalf("cursor after reinitialized allocation = %d, want %d", got, handleStride)
	}
}
