package opennox

import (
	"testing"
	"unsafe"

	"github.com/opennox/opennox/v1/client/noxrender"
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
	var imageByte byte
	images := [...]noxrender.ImageHandle{noxrender.ImageHandle(unsafe.Pointer(&imageByte))}
	data.images = &images[0]
	data.count = 7
	if got := staticRandomDrawFrameCount(unsafe.Pointer(&data)); got != 7 {
		t.Fatalf("static-random frame count = %d, want 7", got)
	}
}

func TestStaticRandomDrawImageUsesNativePointerArray(t *testing.T) {
	var imageBytes [2]byte
	images := [...]noxrender.ImageHandle{
		noxrender.ImageHandle(unsafe.Pointer(&imageBytes[0])),
		noxrender.ImageHandle(unsafe.Pointer(&imageBytes[1])),
	}
	data := nativeStaticRandomDrawData{images: &images[0], count: uint8(len(images))}

	if got, ok := staticRandomDrawImage(unsafe.Pointer(&data), 1); !ok || got != images[1] {
		t.Fatalf("image[1] = (%p, %t), want (%p, true)", got, ok, images[1])
	}
	for _, index := range []int{-1, len(images)} {
		if got, ok := staticRandomDrawImage(unsafe.Pointer(&data), index); ok || got != nil {
			t.Fatalf("image[%d] = (%p, %t), want (nil, false)", index, got, ok)
		}
	}
	if got, ok := staticRandomDrawImage(nil, 0); ok || got != nil {
		t.Fatalf("nil data image = (%p, %t), want (nil, false)", got, ok)
	}
}
