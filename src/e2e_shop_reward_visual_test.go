package opennox

import (
	"image"
	"testing"

	"github.com/opennox/libs/noximage"
)

func TestE2EItemIconPixelsRequireRealSilhouette(t *testing.T) {
	reference := noximage.NewImage16(image.Rect(-2, -3, 4, 5))
	actual := noximage.NewImage16(reference.Rect)
	if pixels, matching := e2eItemIconPixelCount(actual, reference); pixels != 0 || matching != 0 {
		t.Fatalf("empty reference: %d/%d", pixels, matching)
	}
	reference.Pix[reference.PixOffset(1, 2)] = 0x1234
	reference.Pix[reference.PixOffset(2, 3)] = 0x5678
	if pixels, matching := e2eItemIconPixelCount(actual, reference); pixels != 2 || matching != 0 {
		t.Fatalf("background cannot prove icon: %d/%d", pixels, matching)
	}
	actual.Pix[actual.PixOffset(1, 2)] = 0x1234
	if pixels, matching := e2eItemIconPixelCount(actual, reference); pixels != 2 || matching != 1 {
		t.Fatalf("missing icon pixel: %d/%d", pixels, matching)
	}
	actual.Pix[actual.PixOffset(2, 3)] = 0x5678
	if pixels, matching := e2eItemIconPixelCount(actual, reference); pixels != 2 || matching != 2 {
		t.Fatalf("real matching silhouette: %d/%d", pixels, matching)
	}
	actual.Pix[actual.PixOffset(0, 0)] = 0x7fff
	if pixels, matching := e2eItemIconPixelCount(actual, reference); pixels != 2 || matching != 2 {
		t.Fatalf("ignore non-icon background: %d/%d", pixels, matching)
	}
	clipped := noximage.NewImage16(image.Rect(0, 0, 2, 3))
	clipped.Pix[clipped.PixOffset(1, 2)] = 0x1234
	if pixels, matching := e2eItemIconPixelCount(clipped, reference); pixels != 2 || matching != 1 {
		t.Fatalf("offscreen icon pixel is not a match: %d/%d", pixels, matching)
	}
}
