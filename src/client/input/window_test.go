package input

import (
	"image"
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
