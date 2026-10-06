package input

import (
	"image"
	"log/slog"
	"testing"

	"github.com/opennox/libs/client/seat"

	seatheadless "github.com/opennox/opennox/v1/client/seat/headless"
)

func TestIdleSamePositionMotionUsesNormalHeadlessInput(t *testing.T) {
	st := seatheadless.New(image.Pt(640, 480))
	t.Cleanup(func() { _ = st.Close() })
	h := New(slog.Default(), st, false, 0)
	h.SetWinSize(image.Rect(0, 0, 640, 480))
	h.SetDrawWinSize(image.Pt(640, 480))
	st.QueueInput(&seat.MouseMoveEvent{Pos: image.Pt(100, 120)})
	h.Tick()
	position := h.GetMousePos()
	for i := 0; i < 2701; i++ {
		h.Tick()
	}
	if h.SeqDelay() <= 2700 {
		t.Fatalf("no-input delay=%d, want original idle threshold exceeded", h.SeqDelay())
	}
	sequence := h.CurrentSeq()
	// Same canvas cursor position; no button, text, key, reset or direct
	// sequence write. This is the ordinary event used by the cloud setup.
	st.QueueInput(&seat.MouseMoveEvent{Pos: h.DrawPosToWindow(position)})
	h.Tick()
	if h.SeqDelay() != 1 || h.CurrentSeq() != sequence+1 || h.GetMousePos() != position || h.GetMouseRel() != (image.Point{}) {
		t.Fatalf("delay=%d sequence=%d/%d cursor=%v/%v relative=%v", h.SeqDelay(), h.CurrentSeq(), sequence+1, h.GetMousePos(), position, h.GetMouseRel())
	}
	for _, button := range []seat.MouseButton{seat.MouseButtonLeft, seat.MouseButtonRight, seat.MouseButtonMiddle} {
		if h.IsMousePressed(button) {
			t.Fatalf("motion pressed %v", button)
		}
	}
	if event := h.k.nextKeyEvent(); event != nil {
		t.Fatalf("motion produced keyboard input: %+v", event)
	}
	h.Tick()
	if h.SeqDelay() != 2 {
		t.Fatalf("idle delay did not resume normally: %d", h.SeqDelay())
	}
}
