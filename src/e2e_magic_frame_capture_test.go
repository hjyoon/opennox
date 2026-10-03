package opennox

import (
	"bytes"
	"image"
	"image/color"
	"image/png"
	"os"
	"path/filepath"
	"testing"
)

func TestE2EMagicFrameCaptureReadOnlyAndPrivate(t *testing.T) {
	dir := t.TempDir()
	img := image.NewRGBA(image.Rect(0, 0, 8, 6))
	img.SetRGBA(3, 2, color.RGBA{R: 250, G: 70, B: 100, A: 255})
	before := append([]byte(nil), img.Pix...)
	first, err := e2eWriteMagicFrame(dir, img)
	if err != nil {
		t.Fatal(err)
	}
	second, err := e2eWriteMagicFrame(dir, img)
	if err != nil {
		t.Fatal(err)
	}
	if first == second || filepath.Dir(first) != dir || filepath.Dir(second) != dir {
		t.Fatalf("capture did not use fresh private paths: %q, %q", first, second)
	}
	if !bytes.Equal(before, img.Pix) {
		t.Fatal("capture changed the observed pixels")
	}
	for _, path := range []string{first, second} {
		raw, err := os.ReadFile(path)
		if err != nil {
			t.Fatal(err)
		}
		got, err := png.Decode(bytes.NewReader(raw))
		if err != nil {
			t.Fatal(err)
		}
		if got.Bounds() != img.Bounds() {
			t.Fatalf("capture bounds = %v, want %v", got.Bounds(), img.Bounds())
		}
		for y := 0; y < 6; y++ {
			for x := 0; x < 8; x++ {
				if color.NRGBAModel.Convert(got.At(x, y)) != color.NRGBAModel.Convert(img.At(x, y)) {
					t.Fatalf("capture pixel (%d, %d) changed", x, y)
				}
			}
		}
	}
}

func TestE2EMagicFrameCaptureRejectsInvalidDirectory(t *testing.T) {
	dir := filepath.Join(t.TempDir(), "missing", "directory")
	if path, err := e2eWriteMagicFrame(dir, image.NewRGBA(image.Rect(0, 0, 1, 1))); err == nil || path != "" {
		t.Fatalf("invalid directory accepted: path=%q err=%v", path, err)
	}
}

func TestE2EMagicFrameCaptureSchedule(t *testing.T) {
	var sc e2eScenario
	sc.CaptureMagicFrame("diagnostic")
	if len(sc.steps) != 1 || sc.steps[0].name != "diagnostic" || sc.steps[0].fnc == nil || sc.steps[0].ready != nil {
		t.Fatal("capture must schedule one diagnostic step without changing readiness")
	}
}
