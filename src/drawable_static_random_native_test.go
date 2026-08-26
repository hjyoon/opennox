package opennox

import (
	"testing"
	"unsafe"
)

func TestStaticRandomDrawDataUsesNativePointerLayout(t *testing.T) {
	var data nativeStaticRandomDrawData
	pointerSize := unsafe.Sizeof(uintptr(0))
	if got, want := unsafe.Offsetof(data.count), 2*pointerSize; got != want {
		t.Fatalf("static-random frame count offset = %d, want %d", got, want)
	}
	if got, want := unsafe.Sizeof(data), 3*pointerSize; got != want {
		t.Fatalf("static-random draw data size = %d, want %d", got, want)
	}
	data.images = ^uintptr(0)
	data.count = 7
	if got := staticRandomDrawFrameCount(unsafe.Pointer(&data)); got != 7 {
		t.Fatalf("static-random frame count = %d, want 7", got)
	}
}
