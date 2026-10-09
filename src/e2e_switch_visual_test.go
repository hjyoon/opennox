package opennox

import (
	"image"
	"testing"
)

func TestE2EMinimapWallSamplesOriginalInterior(t *testing.T) {
	// Independent literal zoom-100 projections: 640/6=106, top=187,
	// origin=(188,188), wall-grid(10,10) anchor=(42,229), length=23.
	for _, tc := range []struct {
		direction   byte
		first, last image.Point
	}{
		{0, image.Pt(46, 248), image.Pt(61, 233)},
		{1, image.Pt(46, 233), image.Pt(61, 248)},
	} {
		points, err := e2eMinimapWallSamples(image.Pt(10, 10), tc.direction, image.Pt(241, 241), image.Pt(640, 480))
		if err != nil || len(points) != 16 || points[0] != tc.first || points[15] != tc.last {
			t.Fatalf("dir%d projection=%v error=%v", tc.direction, points, err)
		}
	}
}

func TestE2EMinimapWallSamplesRejectsInvalidOrOffscreenInputs(t *testing.T) {
	for _, tc := range []struct {
		grid, center, size image.Point
		direction          byte
	}{
		{image.Pt(-2, 10), image.Pt(241, 241), image.Pt(640, 480), 0},
		{image.Pt(256, 10), image.Pt(241, 241), image.Pt(640, 480), 0},
		{image.Pt(10, 11), image.Pt(241, 241), image.Pt(640, 480), 0},
		{image.Pt(10, 10), image.Pt(241, 241), image.Pt(0, 480), 0},
		{image.Pt(10, 10), image.Pt(241, 241), image.Pt(640, 480), 2},
		{image.Pt(10, 10), image.Pt(800, 800), image.Pt(640, 480), 1},
	} {
		if points, err := e2eMinimapWallSamples(tc.grid, tc.direction, tc.center, tc.size); err == nil || points != nil {
			t.Fatalf("invalid/offscreen input accepted: %+v => %v/%v", tc, points, err)
		}
	}
}
