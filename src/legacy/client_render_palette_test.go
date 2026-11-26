package legacy

import (
	"runtime"
	"testing"

	"github.com/opennox/opennox/v1/legacy/internal/cfixture"
)

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
