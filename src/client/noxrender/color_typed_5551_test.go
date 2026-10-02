package noxrender

import (
	"image"
	"log/slog"
	"testing"

	noxcolor "github.com/opennox/libs/color"
	"github.com/opennox/libs/noximage"
)

func TestSplitTypedColorKeepsFiveBitArithmetic(t *testing.T) {
	for word := 0; word <= 0xffff; word++ {
		want := Color16{
			R: uint16((word>>10)&31) * 8,
			G: uint16((word>>5)&31) * 8,
			B: uint16(word&31) * 8,
		}
		if got := SplitColor(noxcolor.RGBA5551(word)); got != want {
			t.Fatalf("word=%04x: got %+v, want %+v", word, got, want)
		}
	}
}

func TestRGB5551AlphaLineKeepsFiveBitArithmetic(t *testing.T) {
	for _, tc := range []struct {
		name string
		src  uint16
		want uint16
	}{
		{name: "white", src: 0x7fff, want: 0x7bde},
		{name: "red", src: 0x7c00, want: 0x7800},
		{name: "green", src: 0x03e0, want: 0x03c0},
		{name: "blue", src: 0x001f, want: 0x001e},
	} {
		t.Run(tc.name, func(t *testing.T) {
			pix := noximage.NewImage16(image.Rect(0, 0, 2, 1))
			d := newRenderData(2, 1)
			d.SetAlpha(255)
			r := NewRender(slog.Default(), nil)
			r.SetData(d)
			r.SetPixBuffer(pix)
			r.DrawLineAlpha(image.Pt(0, 0), image.Pt(1, 0), noxcolor.RGBA5551(tc.src))
			if pix.Pix[0] != tc.want || pix.Pix[1] != 0 {
				t.Fatalf("pixels=%04x/%04x: want %04x/0000", pix.Pix[0], pix.Pix[1], tc.want)
			}
		})
	}
}
