package input

import (
	"image"
	"math"
	"testing"
)

func TestScaleViewport(t *testing.T) {
	tests := []struct {
		name     string
		view     image.Rectangle
		drawable image.Point
		logical  image.Point
		want     image.Rectangle
	}{
		{
			name:     "same coordinate space",
			view:     image.Rect(80, 0, 1360, 960),
			drawable: image.Pt(1440, 960),
			logical:  image.Pt(1440, 960),
			want:     image.Rect(80, 0, 1360, 960),
		},
		{
			name:     "macOS Retina two times",
			view:     image.Rect(196, 0, 2744, 1912),
			drawable: image.Pt(2940, 1912),
			logical:  image.Pt(1470, 956),
			want:     image.Rect(98, 0, 1372, 956),
		},
		{
			name:     "fractional scale",
			view:     image.Rect(240, 0, 1680, 1080),
			drawable: image.Pt(1920, 1080),
			logical:  image.Pt(1280, 720),
			want:     image.Rect(160, 0, 1120, 720),
		},
		{
			name:     "invalid dimensions",
			view:     image.Rect(10, 20, 30, 40),
			drawable: image.Point{},
			logical:  image.Pt(640, 480),
			want:     image.Rect(10, 20, 30, 40),
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := ScaleViewport(tt.view, tt.drawable, tt.logical); got != tt.want {
				t.Fatalf("ScaleViewport(%v, %v, %v) = %v, want %v", tt.view, tt.drawable, tt.logical, got, tt.want)
			}
		})
	}
}

func TestRetinaMousePositionMapsToMainMenuButton(t *testing.T) {
	var win window
	win.init(image.Pt(640, 480))
	view := ScaleViewport(
		image.Rect(196, 0, 2744, 1912),
		image.Pt(2940, 1912),
		image.Pt(1470, 956),
	)
	win.SetWinSize(view)

	// The center of the first MainMenu.wnd button is displayed near this
	// logical window point after the 4:3 viewport is letterboxed.
	got := win.toDrawSpace(image.Pt(576, 215))
	if got.X < 157 || got.X > 323 || got.Y < 91 || got.Y > 126 {
		t.Fatalf("Retina mouse point mapped to %v, outside first menu button", got)
	}
}

func TestDrawPosToWindow(t *testing.T) {
	for _, tc := range []struct {
		name string
		draw image.Point
		view image.Rectangle
	}{
		{name: "unscaled", draw: image.Pt(640, 480), view: image.Rect(0, 0, 640, 480)},
		{name: "Retina menu", draw: image.Pt(640, 480), view: image.Rect(98, 0, 1372, 956)},
		{name: "Retina game", draw: image.Pt(1024, 768), view: image.Rect(98, 0, 1372, 956)},
		{name: "fractional offset", draw: image.Pt(640, 480), view: image.Rect(160, 12, 1120, 732)},
		{name: "downscaled", draw: image.Pt(1024, 768), view: image.Rect(0, 30, 512, 414)},
	} {
		t.Run(tc.name, func(t *testing.T) {
			h := Handler{m: &mouseHandler{}}
			h.m.win.init(tc.draw)
			h.SetWinSize(tc.view)
			maxError := image.Pt(
				max(0, int(math.Ceil(float64(tc.draw.X)/float64(tc.view.Dx())))-1),
				max(0, int(math.Ceil(float64(tc.draw.Y)/float64(tc.view.Dy())))-1),
			)
			for y := 0; y < tc.draw.Y; y++ {
				for x := 0; x < tc.draw.X; x++ {
					p := image.Pt(x, y)
					logical := h.DrawPosToWindow(p)
					if !logical.In(tc.view) {
						t.Fatalf("canvas point %v maps outside viewport: %v", p, logical)
					}
					got := h.m.win.toDrawSpace(logical)
					if math.Abs(float64(got.X-p.X)) > float64(maxError.X) || math.Abs(float64(got.Y-p.Y)) > float64(maxError.Y) {
						t.Fatalf("canvas %v -> logical %v -> canvas %v, maximum rounding error %v", p, logical, got, maxError)
					}
				}
			}
		})
	}
}

func TestDrawPosToWindowBeforeInitialization(t *testing.T) {
	h := Handler{m: &mouseHandler{}}
	p := image.Pt(12, 34)
	if got := h.DrawPosToWindow(p); got != p {
		t.Fatalf("uninitialized mapping = %v, want %v", got, p)
	}
}
