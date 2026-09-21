package legacy

import (
	"math"
	"testing"
	"unsafe"
)

func TestPolygonEditorNearestAngleUsesNativePointer420E80(t *testing.T) {
	Sub_421B10()
	t.Cleanup(Sub_421B10)

	want := Nox_xxx_polygonSetAngle_420D40(123.5, -47.25, 1, 0)
	if want == nil {
		t.Fatal("could not allocate polygon angle")
	}
	got := Sub_420E80(123.5, -47.25, 900)
	if got != want {
		t.Fatalf("nearest polygon angle = %p, want %p", got, want)
	}
	if got.Index != 1 || got.X != 123.5 || got.Y != -47.25 || got.Active == 0 {
		t.Fatalf("polygon angle = %+v, want index=1 position=(123.5,-47.25) active", *got)
	}
	if unsafe.Sizeof(uintptr(0)) == 8 && uintptr(unsafe.Pointer(got)) <= math.MaxUint32 {
		t.Fatalf("polygon angle address = %p, want a native address above 4 GiB", got)
	}
}
