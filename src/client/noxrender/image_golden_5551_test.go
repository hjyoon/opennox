package noxrender

import (
	"bytes"
	"image"
	"image/color"
	"image/png"
	"slices"
	"testing"

	"github.com/opennox/libs/noximage"
)

// rasterGoldenImage preserves the 2022 golden's left-aligned RGB5551 color
// model (31 -> 248). Presentation's 31 -> 255 conversion is not raster math.
// This view only reads the native pixels; it must not call the raster's
// SplitColor helpers or change production's display conversion.
type rasterGoldenImage struct{ *noximage.Image16 }

func (v rasterGoldenImage) At(x, y int) color.Color {
	word := uint16(v.RGBA5551At(x, y))
	a := byte(255)
	if word&0x8000 != 0 {
		a = 0
	}
	return color.NRGBA{
		R: byte((word >> 7) & 0xf8),
		G: byte((word >> 2) & 0xf8),
		B: byte((word << 3) & 0xf8),
		A: a,
	}
}

func TestRasterGoldenImageColorFields(t *testing.T) {
	pix := noximage.NewImage16(image.Rect(7, 11, 8, 12))
	view := rasterGoldenImage{pix}
	for word := 0; word <= 0xffff; word++ {
		pix.Pix[0] = uint16(word)
		want := color.NRGBA{
			R: byte((word>>10)&31) * 8,
			G: byte((word>>5)&31) * 8,
			B: byte(word&31) * 8,
			A: 255,
		}
		if word >= 0x8000 {
			want.A = 0
		}
		if got := view.At(7, 11); got != want || pix.Pix[0] != uint16(word) {
			t.Fatalf("word=%04x: got %+v, want %+v, source=%04x", word, got, want, pix.Pix[0])
		}
	}
}

func TestRasterGoldenImagePNGRoundTrip(t *testing.T) {
	pix := noximage.NewImage16(image.Rect(7, 11, 263, 139))
	for word := range pix.Pix {
		pix.Pix[word] = uint16(word)
	}
	before := slices.Clone(pix.Pix)
	var encoded bytes.Buffer
	if err := png.Encode(&encoded, rasterGoldenImage{pix}); err != nil {
		t.Fatal(err)
	}
	if !slices.Equal(before, pix.Pix) {
		t.Fatal("golden encoding changed native pixels")
	}
	decoded, err := png.Decode(bytes.NewReader(encoded.Bytes()))
	if err != nil {
		t.Fatal(err)
	}
	if decoded.Bounds() != image.Rect(0, 0, 256, 128) {
		t.Fatalf("decoded bounds=%v", decoded.Bounds())
	}
	for word := range pix.Pix {
		r, g, b, a := decoded.At(word%256, word/256).RGBA()
		wantR := uint32((word>>10)&31) * 8 * 257
		wantG := uint32((word>>5)&31) * 8 * 257
		wantB := uint32(word&31) * 8 * 257
		if r != wantR || g != wantG || b != wantB || a != 65535 {
			t.Fatalf("word=%04x: decoded=%d/%d/%d/%d, want=%d/%d/%d/65535", word, r, g, b, a, wantR, wantG, wantB)
		}
	}
	// The canonical view must still detect an actual one-pixel difference.
	pix.Pix[0] ^= 1
	var changed bytes.Buffer
	if err := png.Encode(&changed, rasterGoldenImage{pix}); err != nil {
		t.Fatal(err)
	}
	if bytes.Equal(encoded.Bytes(), changed.Bytes()) {
		t.Fatal("golden encoding ignored a native pixel change")
	}
}
