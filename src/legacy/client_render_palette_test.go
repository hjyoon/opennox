package legacy

import (
	"runtime"
	"testing"
	"unsafe"

	"github.com/opennox/opennox/v1/legacy/internal/cfixture"
)

func TestDisplayPaletteExpansion435120(t *testing.T) {
	var packed [256][3]byte
	for i := range packed {
		packed[i] = [3]byte{byte(i), byte(17*i + 3), byte(255 - i)}
	}
	original := packed
	var expanded [256][4]byte
	Sub_435120(unsafe.Pointer(&expanded), unsafe.Pointer(&packed))
	for i, rgb := range original {
		want := [4]byte{rgb[0], rgb[1], rgb[2], 4}
		if expanded[i] != want {
			t.Fatalf("color %d = %v, want %v", i, expanded[i], want)
		}
	}
	if packed != original {
		t.Fatal("expanded palette modified the packed source")
	}
	var restored [256][3]byte
	Sub_435150(unsafe.Pointer(&restored), unsafe.Pointer(&expanded))
	if restored != original {
		t.Fatal("palette round trip changed the packed RGB values")
	}
}

func TestLegacyRenderPaletteNativeStorage(t *testing.T) {
	slots := [...]struct {
		index int
		addr  *uint32
	}{
		{0, Get_nox_color_black_2650656_ptr()},
		{4, Get_nox_color_white_2523948_ptr()},
		{5, Get_nox_color_violet_2598268_ptr()},
		{7, Get_nox_color_red_2589776_ptr()},
		{10, Get_nox_color_green_2614268_ptr()},
		{12, Get_nox_color_cyan_2649820_ptr()},
		{13, Get_nox_color_blue_2650684_ptr()},
		{14, Get_nox_color_orange_2614256_ptr()},
		{15, Get_nox_color_yellow_2589772_ptr()},
	}
	var previous, want [len(slots)]uint32
	for i, slot := range slots {
		previous[i] = *slot.addr
		want[i] = 0xa55a0000 | uint32(slot.index)
	}
	defer func() {
		for i, slot := range slots {
			*slot.addr = previous[i]
		}
	}()
	for i, slot := range slots {
		*slot.addr = want[i]
	}
	runtime.GC()
	for i, slot := range slots {
		if got := cfixture.Game2ReferenceColor(slot.index); got != want[i] {
			t.Errorf("C color %d = %#x, want %#x", slot.index, got, want[i])
		}
		if got := g2ServerPlayerColor(slot.index); got != want[i] {
			t.Errorf("Go color %d = %#x, want %#x", slot.index, got, want[i])
		}
	}
}
