package headless_test

import (
	"image"
	"testing"

	libseat "github.com/opennox/libs/client/seat"
	"github.com/opennox/libs/noximage"

	"github.com/opennox/opennox/v1/client/render"
	headless "github.com/opennox/opennox/v1/client/seat/headless"
)

func TestSeatRendersAndCapturesWithoutWindow(t *testing.T) {
	st := headless.New(image.Pt(640, 480))
	r, err := render.New(st)
	if err != nil {
		t.Fatal(err)
	}

	frame := noximage.NewImage16(image.Rect(0, 0, 320, 240))
	frame.Pix[0] = 0x1234
	frame.Pix[len(frame.Pix)-1] = 0x5678
	r.CopyBuffer(frame)

	got, view := st.Snapshot()
	if got == nil {
		t.Fatal("headless seat did not retain the rendered frame")
	}
	if got.Rect != frame.Rect || got.Pix[0] != 0x1234 || got.Pix[len(got.Pix)-1] != 0x5678 {
		t.Fatalf("captured frame differs: rect=%v first=%#x last=%#x", got.Rect, got.Pix[0], got.Pix[len(got.Pix)-1])
	}
	if view != image.Rect(0, 0, 640, 480) {
		t.Fatalf("viewport = %v, want 640x480", view)
	}
	if st.PresentCount() != 1 {
		t.Fatalf("present count = %d, want 1", st.PresentCount())
	}

	// The snapshot must not alias the renderer's source or the stored surface.
	got.Pix[0] = 0
	again, _ := st.Snapshot()
	if again.Pix[0] != 0x1234 {
		t.Fatal("snapshot aliases the in-memory surface")
	}
}

func TestSeatDeliversOnlySyntheticInput(t *testing.T) {
	st := headless.New(image.Pt(640, 480))
	var got []libseat.InputEvent
	st.OnInput(func(ev libseat.InputEvent) {
		got = append(got, ev)
	})
	st.InputTick()
	if len(got) != 0 {
		t.Fatalf("unexpected physical input: %v", got)
	}

	want := &libseat.MouseMoveEvent{Pos: image.Pt(123, 45)}
	st.QueueInput(want, &libseat.MouseButtonEvent{Button: libseat.MouseButtonLeft, Pressed: true})
	st.InputTick()
	if len(got) != 2 || got[0] != want {
		t.Fatalf("synthetic input = %#v, want move then click", got)
	}
	st.InputTick()
	if len(got) != 2 {
		t.Fatal("synthetic input was delivered more than once")
	}
}
