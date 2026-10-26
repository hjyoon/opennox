package render

import (
	"image"
	"testing"

	"github.com/opennox/libs/noximage"

	"github.com/opennox/opennox/v1/client/seat/headless"
)

func TestResizePreservesCanvasAspectRatio(t *testing.T) {
	sc := headless.New(image.Pt(640, 480))
	r, err := New(sc)
	if err != nil {
		t.Fatal(err)
	}
	img := noximage.NewImage16(image.Rect(0, 0, 640, 480))
	img.Pix[0] = 0x1234
	r.CopyBuffer(img)
	var gotView image.Rectangle
	r.OnViewResize(func(view image.Rectangle) { gotView = view })
	sc.ResizeScreen(image.Pt(1920, 1080))
	got, view := sc.Snapshot()
	// A 1440-pixel canvas has 240 pixels of padding on each side.
	want := image.Rect(240, 0, 1680, 1080)
	if view != want || gotView != want {
		t.Fatalf("resized viewport = %v, callback = %v; want %v", view, gotView, want)
	}
	if got.Rect != img.Rect || got.Pix[0] != 0x1234 {
		t.Fatal("resize changed the uploaded frame")
	}
}

func TestResizeBeforeFirstFrameUsesSurfaceAspectRatio(t *testing.T) {
	sc := headless.New(image.Pt(640, 480))
	_, err := New(sc)
	if err != nil {
		t.Fatal(err)
	}
	sc.ResizeScreen(image.Pt(1920, 1080))
	_, view := sc.Snapshot()
	if want := image.Rect(240, 0, 1680, 1080); view != want {
		t.Fatalf("initial resized viewport = %v, want %v", view, want)
	}
}

func TestViewportBinary32RoundingAcrossOrientations(t *testing.T) {
	for _, test := range []struct {
		name    string
		canvas  image.Point
		screen  image.Point
		rotate  bool
		stretch bool
		want    image.Rectangle
	}{
		{"landscape", image.Pt(640, 480), image.Pt(1920, 1080), false, false, image.Rect(240, 0, 1680, 1080)},
		{"portrait", image.Pt(640, 480), image.Pt(1080, 1920), false, false, image.Rect(0, 555, 1080, 1365)},
		{"rotated portrait", image.Pt(640, 480), image.Pt(1080, 1920), true, false, image.Rect(0, 240, 1080, 1680)},
		{"portrait canvas", image.Pt(480, 640), image.Pt(1920, 1080), false, false, image.Rect(555, 0, 1365, 1080)},
		{"odd screen", image.Pt(3, 2), image.Pt(17, 10), false, false, image.Rect(1, 0, 16, 10)},
		{"stretched", image.Pt(640, 480), image.Pt(1920, 1080), false, true, image.Rect(0, 0, 1920, 1080)},
	} {
		t.Run(test.name, func(t *testing.T) {
			sc := headless.New(test.canvas)
			r, err := New(sc)
			if err != nil {
				t.Fatal(err)
			}
			defer r.Close()
			r.rotate = test.rotate
			r.SetStretched(test.stretch)
			sc.ResizeScreen(test.screen)
			_, view := sc.Snapshot()
			if view != test.want {
				t.Fatalf("viewport=%v, want %v", view, test.want)
			}
		})
	}
}
