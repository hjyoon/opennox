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
	// Keep the renderer's existing float32 aspect-ratio rounding.
	want := image.Rect(239, 0, 1679, 1080)
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
	if want := image.Rect(239, 0, 1679, 1080); view != want {
		t.Fatalf("initial resized viewport = %v, want %v", view, want)
	}
}
