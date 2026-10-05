package opennox

import (
	"fmt"
	"image"
	"log/slog"
	"testing"

	"github.com/opennox/libs/client/keybind"
	"github.com/opennox/libs/client/seat"
	"github.com/opennox/opennox/v1/client/input"
	seatheadless "github.com/opennox/opennox/v1/client/seat/headless"
)

func TestE2EConsoleKeyboardDefaultLayout(t *testing.T) {
	for ch := rune(' '); ch <= '~'; ch++ {
		t.Run(fmt.Sprintf("U+%04X", ch), func(t *testing.T) {
			st := seatheadless.New(image.Pt(640, 480))
			t.Cleanup(func() { _ = st.Close() })
			inp := input.New(slog.Default(), st, false, 0)
			key, shift := e2eConsoleKey(ch)
			if shift {
				st.QueueInput(&seat.KeyboardEvent{Key: keybind.KeyLShift, Pressed: true})
			}
			st.QueueInput(&seat.KeyboardEvent{Key: key, Pressed: true})
			inp.Tick()
			if got := inp.KeyToWChar(key); got != uint16(ch) {
				t.Fatalf("key %s shift=%t produced U+%04X, want U+%04X", key, shift, got, ch)
			}
		})
	}
}

func TestE2EConsoleKeyboardPairsReachNativeInput(t *testing.T) {
	for _, language := range []int{0, 6, 8} {
		t.Run(fmt.Sprintf("language-%d", language), func(t *testing.T) {
			f := newConsoleKeyboardTextFixture(t, language)
			oldInput := e2e.input
			e2e.input = nil
			t.Cleanup(func() { e2e.input = oldInput })
			text := `SET NAME "aaa AAA 111 !_+"`
			if language != 0 {
				text = `SET NAME "aaa 한글 中 !_+"`
			}
			var sc e2eScenario
			sc.consoleTypeKeyboardText(text)
			for _, step := range sc.steps {
				step.fnc()
				var events []seat.InputEvent
				for _, queued := range e2e.input {
					events = append(events, queued.event)
				}
				e2e.input = nil
				f.poll(events...)
			}
			if got := f.text(); got != text {
				t.Fatalf("paired keyboard input became %q, want %q", got, text)
			}
		})
	}
}
