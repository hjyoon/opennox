package noxrender

import (
	"testing"

	noxcolor "github.com/opennox/libs/color"
)

func TestRenderMaterialKeepsFiveBitArithmetic(t *testing.T) {
	d := newRenderData(1, 1)
	for word := 0; word <= 0xffff; word++ {
		d.SetMaterial(0, noxcolor.RGBA5551(word))
		want := RGB{
			R: ((word >> 10) & 31) * 8,
			G: ((word >> 5) & 31) * 8,
			B: (word & 31) * 8,
		}
		wantPacked := uint32(word) | uint32(word)<<16
		if word == 0x8000 {
			wantPacked = 0x80000000
		}
		m := d.material(0)
		if m.Color != want || m.Color32 != wantPacked {
			t.Fatalf("word=%04x: got %+v/%08x, want %+v/%08x", word, m.Color, m.Color32, want, wantPacked)
		}
		before := *m
		d.SetMaterial(0, noxcolor.RGBA5551(word))
		if *m != before {
			t.Fatalf("word=%04x: repeated packed color changed material", word)
		}
	}
	before := d.materials
	for _, index := range []int{-1, 16} {
		d.SetMaterial(index, noxcolor.RGBA5551(0x7fff))
		if d.materials != before {
			t.Fatalf("out-of-range material %d changed the palette", index)
		}
	}
	// Equipment's byte-RGB API intentionally preserves unquantized inputs.
	d.SetMaterialRGB(15, 255, 255, 255)
	if m := d.material(15); m.Color != (RGB{255, 255, 255}) || m.Color32 != 0x7fff7fff {
		t.Fatalf("byte-RGB material changed: %+v", m)
	}
}
