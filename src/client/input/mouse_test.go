package input

import (
	"image"
	"log/slog"
	"testing"

	"github.com/opennox/libs/client/seat"
	"github.com/stretchr/testify/require"

	seatheadless "github.com/opennox/opennox/v1/client/seat/headless"
)

func TestMouseButtonCountMatchesNativeInput(t *testing.T) {
	buttons := []seat.MouseButton{seat.MouseButtonLeft, seat.MouseButtonRight, seat.MouseButtonMiddle}
	if MouseButtonCount != len(buttons) {
		t.Fatalf("logical capacity=%d, supported buttons=%v", MouseButtonCount, buttons)
	}
	for _, button := range buttons {
		t.Run(toMouseBtn(button).String(), func(t *testing.T) {
			st := seatheadless.New(image.Pt(640, 480))
			t.Cleanup(func() { _ = st.Close() })
			h := New(slog.Default(), st, false, 0)
			h.SetDrawWinSize(image.Pt(640, 480))
			st.QueueInput(&seat.MouseMoveEvent{Pos: image.Pt(100, 100)})
			h.Tick()
			for _, pressed := range []bool{true, false} {
				st.QueueInput(&seat.MouseButtonEvent{Button: button, Pressed: pressed})
				h.Tick()
				want := NOX_MOUSE_UP
				if pressed {
					want = NOX_MOUSE_DOWN
				}
				if got := h.GetMouseState(button); got != ToMouseState(toMouseBtn(button), want) {
					t.Fatalf("pressed=%t: state=%d, want %d", pressed, got, ToMouseState(toMouseBtn(button), want))
				}
				for _, other := range buttons {
					if got := h.IsMousePressed(other); got != (pressed && other == button) {
						t.Fatalf("pressed=%t: button %v pressed=%t", pressed, other, got)
					}
				}
			}
			st.QueueInput(&seat.MouseWheelEvent{Wheel: 19})
			h.Tick()
			if h.GetMouseWheel() != 19 || MouseButtonCount != len(h.m.cur.btn) {
				t.Fatalf("wheel=%d, logical capacity=%d", h.GetMouseWheel(), MouseButtonCount)
			}
		})
	}
}

func TestMouseStateCode(t *testing.T) {
	states := []MouseStateCode{
		NOX_MOUSE_LEFT_DOWN,
		NOX_MOUSE_LEFT_DRAG_END,
		NOX_MOUSE_LEFT_UP,
		NOX_MOUSE_LEFT_PRESSED,
		NOX_MOUSE_RIGHT_DOWN,
		NOX_MOUSE_RIGHT_DRAG_END,
		NOX_MOUSE_RIGHT_UP,
		NOX_MOUSE_RIGHT_PRESSED,
		NOX_MOUSE_MIDDLE_DOWN,
		NOX_MOUSE_MIDDLE_DRAG_END,
		NOX_MOUSE_MIDDLE_UP,
		NOX_MOUSE_MIDDLE_PRESSED,
	}
	for i, btn := range []MouseButton{
		-1, NOX_MOUSE_LEFT, NOX_MOUSE_RIGHT, NOX_MOUSE_MIDDLE,
	} {
		t.Run(btn.String(), func(t *testing.T) {
			for j, st := range []MouseState{
				NOX_MOUSE_DOWN, NOX_MOUSE_DRAG_END, NOX_MOUSE_UP, NOX_MOUSE_PRESSED,
			} {
				t.Run(st.String(), func(t *testing.T) {
					bst := ToMouseState(btn, st)
					t.Log(bst)
					if i != 0 {
						require.Equal(t, states[4*(i-1)+j], bst)
					}
					btn2, st2 := bst.Split()
					require.Equal(t, btn, btn2)
					require.Equal(t, st, st2)
				})
			}
		})
	}
}
