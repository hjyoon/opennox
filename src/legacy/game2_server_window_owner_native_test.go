package legacy

import (
	"bytes"
	"runtime"
	"testing"
	"unsafe"
)

func TestGame2ServerNativeOptionsWindowGetter(t *testing.T) {
	// Fixture field8 is the actual C-owned nox_window* global. The getter
	// must return that pointer, rather than reading its low DWORD as an int.
	old := game2ServerFixtureOwner(8)
	var allocations [2]unsafe.Pointer
	defer func() {
		game2ServerFixtureSetOwner(8, old)
		for _, p := range allocations {
			if p != nil {
				game2ServerFixtureFree(p)
			}
		}
	}()

	windowBytes := game2ServerFixtureLayout(0)
	var windows [2]unsafe.Pointer
	var before [2][]byte
	for i := range allocations {
		p := game2ServerFixtureAlloc(windowBytes + 32)
		if p == nil {
			t.Fatal("C window allocation")
		}
		allocations[i] = p
		windows[i] = unsafe.Add(p, 16)
		if unsafe.Sizeof(uintptr(0)) == 8 && uintptr(windows[i]) <= 0xFFFFFFFF {
			t.Fatal("C window must exercise address bits above the low DWORD")
		}
		storage := unsafe.Slice((*byte)(p), int(windowBytes)+32)
		for j := range storage {
			storage[j] = byte(0x39 + i*0x21 + j*7)
		}
		before[i] = append([]byte(nil), storage...)
	}
	if windows[0] == windows[1] {
		t.Fatal("independent C window allocations alias")
	}

	for _, want := range []unsafe.Pointer{nil, windows[0], windows[1], windows[0], nil} {
		game2ServerFixtureSetOwner(8, want)
		for gc := 0; gc < 3; gc++ {
			runtime.GC()
			var got unsafe.Pointer = Get_dword_5d4594_1046492()
			if got != want || game2ServerFixtureOwner(8) != want {
				t.Fatalf("public getter lost live C window identity: got %p want %p", got, want)
			}
			for i, p := range allocations {
				if !bytes.Equal(unsafe.Slice((*byte)(p), int(windowBytes)+32), before[i]) {
					t.Fatal("getter changed a complete C window or its surrounding canaries")
				}
			}
		}
	}
}
