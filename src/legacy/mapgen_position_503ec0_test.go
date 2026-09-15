package legacy

import (
	"math"
	"runtime"
	"testing"
	"unsafe"
)

func TestMapgenPosition503EC0UsesNativeObjectPosition(t *testing.T) {
	got := mapgenPositionFixture503EC0(3000, 3000, 3000, 3000)
	if !got.valid {
		t.Fatal("position helper rejected an active generated map")
	}
	if got.outputBits != [2]uint32{} {
		t.Fatalf("relative position bits = %08x, want two positive zeroes", got.outputBits)
	}
	if want := [2]uint32{math.Float32bits(3000), math.Float32bits(3000)}; got.objectBits != want {
		t.Fatalf("live object position bits = %08x, want %08x", got.objectBits, want)
	}
	if unsafe.Sizeof(uintptr(0)) == 8 && runtime.GOOS != "windows" && got.objectAddress <= 0xffffffff {
		t.Fatalf("fixture object address = %#x, want above 4 GiB", got.objectAddress)
	}
	if got.gated {
		t.Fatal("position helper accepted an already placed generated map")
	}
	if got.nullGated {
		t.Fatal("position helper accepted nil pointers behind an already-placed gate")
	}
	if want := [2]uint32{math.Float32bits(123.25), math.Float32bits(-456.5)}; got.gatedOutputBits != want {
		t.Fatalf("gated output bits = %08x, want untouched %08x", got.gatedOutputBits, want)
	}
	if want := [2]uint32{math.Float32bits(-999), math.Float32bits(9999)}; got.gatedObjectBits != want {
		t.Fatalf("gated object bits = %08x, want untouched %08x", got.gatedObjectBits, want)
	}
}

func TestMapgenPosition503EC0ClampsLiveObjectCoordinates(t *testing.T) {
	got := mapgenPositionFixture503EC0(1, 6000, 3000, 3000)
	if !got.valid {
		t.Fatal("position helper rejected an active generated map")
	}
	if want := [2]uint32{math.Float32bits(82.5), math.Float32bits(5852.5)}; got.objectBits != want {
		t.Fatalf("clamped live object position bits = %08x, want %08x", got.objectBits, want)
	}
}
