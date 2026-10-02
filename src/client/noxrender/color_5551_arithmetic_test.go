package noxrender

import (
	"encoding/binary"
	"log/slog"
	"testing"

	noxcolor "github.com/opennox/libs/color"
)

func TestSplitColor16KeepsFiveBitArithmetic(t *testing.T) {
	// Raster arithmetic expands each five-bit field by shifting, not by
	// saturating 31 to 255 as the display color model does.
	for word := 0; word <= 0xffff; word++ {
		want := Color16{
			R: uint16((word>>10)&31) * 8,
			G: uint16((word>>5)&31) * 8,
			B: uint16(word&31) * 8,
		}
		if got := SplitColor16(uint16(word)); got != want {
			t.Fatalf("word=%04x: got %+v, want %+v", word, got, want)
		}
	}
	// This fix must not change the library's presentation-only white value.
	if got := noxcolor.RGBA5551(0x7fff).ColorNRGBA(); got.R != 255 || got.G != 255 || got.B != 255 {
		t.Fatalf("display white changed: %+v", got)
	}
}

func TestRGB5551MultiplyKeepsFiveBitArithmetic(t *testing.T) {
	r := NewRender(slog.Default(), nil)
	r.SetData(newRenderData(1, 1))
	for _, tc := range []struct {
		name string
		src  uint16
		want uint16
	}{
		{name: "white", src: 0x7fff, want: 0x7bde},
		{name: "red", src: 0x7c00, want: 0x7800},
		{name: "green", src: 0x03e0, want: 0x03c0},
		{name: "blue", src: 0x001f, want: 0x001e},
		{name: "inverted alpha", src: 0xffff, want: 0x7bde},
	} {
		t.Run(tc.name, func(t *testing.T) {
			src := make([]byte, 2)
			binary.LittleEndian.PutUint16(src, tc.src)
			dst := []uint16{0x1234}
			drest, srest := r.pixOpSrcMultiply(dst, src, 1)
			if dst[0] != tc.want || len(drest) != 0 || len(srest) != 0 {
				t.Fatalf("pixel=%04x, tails=%d/%d: want %04x and empty tails", dst[0], len(drest), len(srest), tc.want)
			}
			if binary.LittleEndian.Uint16(src) != tc.src {
				t.Fatal("source pixel changed")
			}
		})
	}
}
