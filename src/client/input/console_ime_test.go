package input

import (
	"log/slog"
	"testing"

	"github.com/opennox/libs/client/keybind"
	"github.com/opennox/libs/client/seat"
)

type consoleIMESequencer uint

func (s *consoleIMESequencer) NextEventSeq() uint {
	seq := uint(*s)
	*s++
	return seq
}

func TestConsoleIMECommitReturnDoesNotBecomeCommandReturn(t *testing.T) {
	for _, key := range []keybind.Key{keybind.KeyEnter, keybind.KeyKpEnter} {
		t.Run(key.String(), func(t *testing.T) {
			var seq consoleIMESequencer
			h := newKeyboardHandler(slog.Default(), &seq, false)
			h.TextEdit(&seat.TextEditEvent{Text: "한"})
			h.Keyboard(&seat.KeyboardEvent{Key: key, Pressed: true})
			// SDL may commit text before GUI.ProcessKeys consumes the queue.
			// The empty buffer at dispatch must not reinterpret that same key.
			h.TextInput(&seat.TextInputEvent{Text: "한"})
			if h.GetInputEditBuffer() != "" {
				t.Fatal("composition did not complete")
			}
			if ev := h.nextKeyEvent(); ev != nil {
				t.Fatalf("composition-confirming Return leaked into command queue: %+v", ev)
			}
			h.Keyboard(&seat.KeyboardEvent{Key: key, Pressed: false})
			if ev := h.nextKeyEvent(); ev == nil || ev.Pressed || ev.Code != key {
				t.Fatal("ordinary key release was lost")
			}
			h.Keyboard(&seat.KeyboardEvent{Key: key, Pressed: true})
			if ev := h.nextKeyEvent(); ev == nil || !ev.Pressed || ev.Code != key {
				t.Fatal("the next command Return was lost")
			}
		})
	}
}

func TestConsoleIMEPreservesOtherKeysAndAltReturn(t *testing.T) {
	var seq consoleIMESequencer
	h := newKeyboardHandler(slog.Default(), &seq, false)
	h.TextEdit(&seat.TextEditEvent{Text: "한"})
	h.Keyboard(&seat.KeyboardEvent{Key: keybind.KeyA, Pressed: true})
	if ev := h.nextKeyEvent(); ev == nil || ev.Code != keybind.KeyA || !ev.Pressed {
		t.Fatal("IME composition suppressed an unrelated key")
	}
	h.nox_input_map_byKey[keybind.KeyLAlt] = noxKeyState{state: 2}
	fullScreen := 0
	h.OnToggleFullScreen(func() { fullScreen++ })
	h.Keyboard(&seat.KeyboardEvent{Key: keybind.KeyEnter, Pressed: true})
	if fullScreen != 1 || h.nextKeyEvent() != nil {
		t.Fatal("Alt+Return fullscreen behavior changed")
	}
}
