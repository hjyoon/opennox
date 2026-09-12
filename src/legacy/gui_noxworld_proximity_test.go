package legacy

import (
	"image"
	"testing"
	"unsafe"

	"github.com/opennox/opennox/v1/common/memmap"
)

func TestWOLServerProximityKeepsNativePointer(t *testing.T) {
	// Standalone legacy tests do not load the PE32 read-only data blob.
	radius := memmap.PtrFloat64(0x581450, 9720)
	oldRadius := *radius
	*radius = 20
	defer func() { *radius = oldRadius }()

	server := &Nox_gui_server_ent_t{Field_11_0: 216, Field_11_2: 27}
	if unsafe.Sizeof(uintptr(0)) > 4 && uintptr(unsafe.Pointer(server)) <= 0xffffffff {
		t.Skip("test requires a server record above the 32-bit address range")
	}
	if !server.NearWOLPoint(image.Pt(216, 27)) {
		t.Fatal("server at the cursor position must be nearby")
	}
	for _, cursor := range []image.Point{image.Pt(206, 27), image.Pt(226, 27), image.Pt(216, 17), image.Pt(216, 37)} {
		if !server.NearWOLPoint(cursor) {
			t.Fatalf("server ten pixels from %v must be nearby", cursor)
		}
	}
	if server.NearWOLPoint(image.Pt(300, 27)) {
		t.Fatal("server 84 pixels from the cursor must not be nearby")
	}
	if (*Nox_gui_server_ent_t)(nil).NearWOLPoint(image.Pt(216, 27)) {
		t.Fatal("nil server cannot be nearby")
	}
}
