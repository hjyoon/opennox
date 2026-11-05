package noxrender

import (
	"bytes"
	"crypto/sha256"
	_ "embed"
	"encoding/binary"
	"encoding/hex"
	"encoding/json"
	"image"
	"image/png"
	"log/slog"
	"testing"
)

type pngPixelGolden struct {
	Width        int    `json:"width"`
	Height       int    `json:"height"`
	RGBA16SHA256 string `json:"rgba16be_sha256"`
}

// These pixels come from PNGs that first matched every pre-existing MD5.
// See testdata/png_golden_pixels.md; compressed bytes vary with Go's encoder.
//
//go:embed testdata/png_golden_pixels.json
var pngPixelGoldenJSON []byte

var pngPixelGoldens = func() map[string]pngPixelGolden {
	var refs map[string]pngPixelGolden
	if err := json.Unmarshal(pngPixelGoldenJSON, &refs); err != nil {
		panic(err)
	}
	return refs
}()

func decodedPNGPixelGolden(encoded []byte) (pngPixelGolden, error) {
	img, err := png.Decode(bytes.NewReader(encoded))
	if err != nil {
		return pngPixelGolden{}, err
	}
	bounds := img.Bounds()
	h := sha256.New()
	var pixel [8]byte
	for y := bounds.Min.Y; y < bounds.Max.Y; y++ {
		for x := bounds.Min.X; x < bounds.Max.X; x++ {
			r, g, b, a := img.At(x, y).RGBA()
			for i, value := range [4]uint32{r, g, b, a} {
				binary.BigEndian.PutUint16(pixel[2*i:], uint16(value))
			}
			_, _ = h.Write(pixel[:])
		}
	}
	return pngPixelGolden{bounds.Dx(), bounds.Dy(), hex.EncodeToString(h.Sum(nil))}, nil
}

func requireRasterPNGGolden(t *testing.T, originalMD5 string, encoded []byte) {
	t.Helper()
	want, ok := pngPixelGoldens[originalMD5]
	if !ok {
		t.Fatalf("missing preserved PNG golden %q", originalMD5)
	}
	got, err := decodedPNGPixelGolden(encoded)
	if err != nil {
		t.Fatal(err)
	}
	if got != want {
		t.Fatalf("pixels differ from preserved PNG %s: got %+v, want %+v", originalMD5, got, want)
	}
}

func TestRasterPNGGoldenDetectsNativePixelChange(t *testing.T) {
	c := particleCases[0]
	img := genParticle(c.opt)
	off, size, _ := img.Meta()
	canvas := image.Pt(2*(size.X+off.X), 2*(size.Y+off.Y))
	pix := newBlack16(canvas.X, canvas.Y)
	r := NewRender(slog.Default(), nil)
	r.SetPixBuffer(pix)
	r.SetData(newRenderData(canvas.X, canvas.Y))
	r.DrawImage16(img, image.Pt(size.X/2, size.Y/2))
	var encoded bytes.Buffer
	if err := png.Encode(&encoded, pix); err != nil {
		t.Fatal(err)
	}
	requireRasterPNGGolden(t, c.exp, encoded.Bytes())
	pix.Pix[0] ^= 1
	encoded.Reset()
	if err := png.Encode(&encoded, pix); err != nil {
		t.Fatal(err)
	}
	changed, err := decodedPNGPixelGolden(encoded.Bytes())
	if err != nil {
		t.Fatal(err)
	}
	if changed == pngPixelGoldens[c.exp] {
		t.Fatal("one native pixel bit change escaped the preserved golden")
	}
}
