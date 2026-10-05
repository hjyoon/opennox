package opennox

import (
	"log/slog"
	"testing"
	"unicode/utf16"

	"github.com/opennox/libs/client/keybind"
	"github.com/opennox/libs/client/seat"
	"github.com/opennox/opennox/v1/client"
	"github.com/opennox/opennox/v1/client/gui"
	"github.com/opennox/opennox/v1/client/input"
)

type consoleKeyboardTextFixture struct {
	g   *gui.GUI
	inp *input.Handler
	win *gui.Window
}

func newConsoleKeyboardTextFixture(t *testing.T, language int) consoleKeyboardTextFixture {
	t.Helper()
	consoleCommandTestFlags(t)
	oldClient := noxClient
	t.Cleanup(func() { noxClient = oldClient })
	g := gui.New(nil)
	t.Cleanup(g.DestroyAll)
	inp := input.New(slog.Default(), &consoleEntryTestInput{}, false, language)
	noxClient = &Client{Client: &client.Client{Inp: inp, GUI: g}, ctrl: new(CtrlEventHandler)}
	g.SetInput(inp)
	draw := gui.WindowData{Style: gui.StyleEntryField}
	win := gui.NewEntryFieldRaw(g, nil, gui.StatusEnabled, 0, 0, 240, 20, &draw, &gui.EntryFieldData{Field_1040: 128})
	if win == nil {
		t.Fatal("native console entry allocation failed")
	}
	con := &guiConsole{input: win}
	win.SetFunc93(con.inputProc)
	// Match NoxInputOnChar's existing language gate. For other languages the
	// stock scan-code path is still the sole character source.
	inp.OnInputString(func(text string) {
		if lang := inp.Language(); lang != 6 && lang != 8 {
			return
		}
		for _, ch := range utf16.Encode([]rune(text)) {
			gui.EntryFieldOnChar(win, ch)
		}
	})
	g.Focus(win)
	return consoleKeyboardTextFixture{g: g, inp: inp, win: win}
}

func (f consoleKeyboardTextFixture) poll(events ...seat.InputEvent) {
	for _, ev := range events {
		f.inp.InputEvent(ev)
	}
	f.inp.Tick()
	f.g.ProcessKeys(f.inp)
}

func (f consoleKeyboardTextFixture) text() string {
	return eventRespStr(f.win.Func94(gui.AsWindowEvent(0x401d, 0, 0)))
}

func TestConsoleKeyboardTextIsInsertedOnce(t *testing.T) {
	for _, language := range []int{6, 8} {
		name := map[int]string{6: "Korean", 8: "Chinese"}[language]
		t.Run(name, func(t *testing.T) {
			for _, tc := range []struct {
				name  string
				key   keybind.Key
				text  string
				order int
			}{
				{"key-then-text", keybind.KeyA, "a", 0},
				{"text-then-key", keybind.KeyA, "a", 1},
				{"key-before-next-poll-text", keybind.KeyA, "a", 2},
				{"text-before-next-poll-key", keybind.KeyA, "a", 3},
				{"shifted-symbol", keybind.Key1, "!", 0},
				{"space", keybind.KeySpace, " ", 0},
				{"composition-commit", keybind.KeyG, "한", 0},
				{"chinese-commit", keybind.KeyA, "中", 0},
				{"multi-character-commit", keybind.KeyA, "aaa 한글", 0},
				{"pending-composition", keybind.KeyA, "", 2},
			} {
				t.Run(tc.name, func(t *testing.T) {
					f := newConsoleKeyboardTextFixture(t, language)
					key := &seat.KeyboardEvent{Key: tc.key, Pressed: true}
					text := &seat.TextInputEvent{Text: tc.text}
					if tc.name == "composition-commit" || tc.name == "pending-composition" {
						f.inp.InputEvent(&seat.TextEditEvent{Text: "ㅎ"})
					}
					switch tc.order {
					case 0:
						f.poll(key, text)
					case 1:
						f.poll(text, key)
					case 2:
						f.poll(key)
						if got := f.text(); got != "" {
							t.Fatalf("scan-code text leaked before SDL commit: %q", got)
						}
						f.poll(text)
					case 3:
						f.poll(text)
						f.poll(key)
					}
					f.poll(&seat.KeyboardEvent{Key: tc.key, Pressed: false})
					if got := f.text(); got != tc.text {
						t.Fatalf("one committed text event became %q, want %q", got, tc.text)
					}
				})
			}
		})
	}
}

func TestConsoleKeyboardTextPreservesRepeatPasteAndBackspace(t *testing.T) {
	for _, language := range []int{6, 8} {
		name := map[int]string{6: "Korean", 8: "Chinese"}[language]
		t.Run(name, func(t *testing.T) {
			f := newConsoleKeyboardTextFixture(t, language)
			for i := 0; i < 3; i++ {
				f.poll(&seat.KeyboardEvent{Key: keybind.KeyA, Pressed: true}, &seat.TextInputEvent{Text: "a"})
			}
			if got := f.text(); got != "aaa" {
				t.Fatalf("legitimate repeated text changed: %q", got)
			}
			f.poll(&seat.KeyboardEvent{Key: keybind.KeyA, Pressed: false}, &seat.TextInputEvent{Text: " 이름"})
			if got := f.text(); got != "aaa 이름" {
				t.Fatalf("committed/pasted text changed: %q", got)
			}
			f.poll(&seat.KeyboardEvent{Key: keybind.KeyBackspace, Pressed: true})
			f.poll(&seat.KeyboardEvent{Key: keybind.KeyBackspace, Pressed: false})
			if got := f.text(); got != "aaa 이" {
				t.Fatalf("Backspace did not delete exactly one character: %q", got)
			}
		})
	}
}

func TestConsoleKeyboardTextPreservesLegacyLanguage(t *testing.T) {
	f := newConsoleKeyboardTextFixture(t, 0)
	f.poll(&seat.KeyboardEvent{Key: keybind.KeyA, Pressed: true}, &seat.TextInputEvent{Text: "a"})
	f.poll(&seat.KeyboardEvent{Key: keybind.KeyA, Pressed: false})
	if got := f.text(); got != "a" {
		t.Fatalf("legacy non-IME character path changed: %q", got)
	}
}

func TestConsoleKeyboardTextTracksLanguageSelection(t *testing.T) {
	f := newConsoleKeyboardTextFixture(t, 0)
	for _, tc := range []struct {
		language int
		key      keybind.Key
		text     string
		want     string
	}{
		{0, keybind.KeyA, "a", "a"},
		{6, keybind.KeyG, "한", "a한"},
		{8, keybind.KeyA, "中", "a한中"},
		{0, keybind.KeyB, "b", "a한中b"},
	} {
		f.inp.SetLanguage(tc.language)
		f.poll(&seat.KeyboardEvent{Key: tc.key, Pressed: true}, &seat.TextInputEvent{Text: tc.text})
		f.poll(&seat.KeyboardEvent{Key: tc.key, Pressed: false})
		if got := f.text(); got != tc.want {
			t.Fatalf("language %d: input = %q, want %q", tc.language, got, tc.want)
		}
	}
}
