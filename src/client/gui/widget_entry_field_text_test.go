package gui

import (
	"image"
	"log/slog"
	"strconv"
	"testing"
	"unicode/utf16"

	"github.com/opennox/libs/client/keybind"
	"github.com/opennox/libs/client/seat"
	"github.com/opennox/opennox/v1/client/input"
	seatheadless "github.com/opennox/opennox/v1/client/seat/headless"
	"github.com/opennox/opennox/v1/legacy/common/alloc"
)

type entryTextFixture struct {
	g   *GUI
	st  *seatheadless.Seat
	inp *input.Handler
	win *Window
}

func newEntryTextFixture(t *testing.T, language int, opts EntryFieldData) entryTextFixture {
	t.Helper()
	g := New(nil)
	t.Cleanup(g.DestroyAll)
	st := seatheadless.New(image.Pt(640, 480))
	t.Cleanup(func() { _ = st.Close() })
	inp := input.New(slog.Default(), st, false, language)
	g.SetInput(inp)
	draw := WindowData{Style: StyleEntryField}
	win := NewEntryFieldRaw(g, nil, StatusEnabled, 0, 0, 240, 20, &draw, &opts)
	if win == nil {
		t.Fatal("native entry allocation failed")
	}
	inp.OnInputString(func(text string) {
		if lang := inp.Language(); lang != 6 && lang != 8 {
			return
		}
		for _, ch := range utf16.Encode([]rune(text)) {
			EntryFieldOnChar(g.Focused(), ch)
		}
	})
	g.Focus(win)
	return entryTextFixture{g: g, st: st, inp: inp, win: win}
}

func (f entryTextFixture) poll(events ...seat.InputEvent) {
	f.st.QueueInput(events...)
	f.inp.Tick()
	f.g.ProcessKeys(f.inp)
}

func (f entryTextFixture) text() string {
	return alloc.GoString16(&entryFieldData(f.win).Text[0])
}

func TestEntryFieldCommittedTextOwnsPrintableKeys(t *testing.T) {
	for _, language := range []int{6, 8} {
		t.Run(strconv.Itoa(language), func(t *testing.T) {
			for _, tc := range []struct {
				name string
				opts EntryFieldData
				key  keybind.Key
				text string
				want string
			}{
				{"plain", EntryFieldData{Field_1040: 64}, keybind.KeyA, "한 a", "한 a"},
				{"numeric", EntryFieldData{Field_1040: 64, Field_1028: 1}, keybind.Key1, "1a2!", "12"},
				{"alphanumeric", EntryFieldData{Field_1040: 64, Field_1032: 1}, keybind.KeyA, "A!한 1", "A한1"},
				{"password", EntryFieldData{Field_1040: 64, Field_1024: 1}, keybind.KeyA, "a한1!", "a한1!"},
				{"length-limit", EntryFieldData{Field_1040: 5}, keybind.KeyA, "abcdef", "abcd"},
			} {
				t.Run(tc.name, func(t *testing.T) {
					f := newEntryTextFixture(t, language, tc.opts)
					f.poll(&seat.KeyboardEvent{Key: tc.key, Pressed: true}, &seat.TextInputEvent{Text: tc.text})
					f.poll(&seat.KeyboardEvent{Key: tc.key, Pressed: false})
					if got := f.text(); got != tc.want {
						t.Fatalf("committed text = %q, want %q", got, tc.want)
					}
					for _, key := range []keybind.Key{keybind.KeyBackspace, keybind.KeyDel} {
						f.poll(&seat.KeyboardEvent{Key: key, Pressed: true})
						f.poll(&seat.KeyboardEvent{Key: key, Pressed: false})
					}
					want := string([]rune(tc.want)[:len([]rune(tc.want))-2])
					if got := f.text(); got != want {
						t.Fatalf("Backspace/Delete changed: text = %q, want %q", got, want)
					}
				})
			}
		})
	}
}

func TestEntryFieldLegacyKeyboardTextIsPreserved(t *testing.T) {
	for _, language := range []int{0, 1, 2, 3, 4, 5, 7, 9} {
		t.Run(strconv.Itoa(language), func(t *testing.T) {
			f := newEntryTextFixture(t, language, EntryFieldData{Field_1040: 64})
			want := string(rune(f.inp.KeyToWChar(keybind.KeyA)))
			f.poll(&seat.KeyboardEvent{Key: keybind.KeyA, Pressed: true}, &seat.TextInputEvent{Text: want})
			f.poll(&seat.KeyboardEvent{Key: keybind.KeyA, Pressed: false})
			if got := f.text(); got != want {
				t.Fatalf("legacy text = %q, want %q", got, want)
			}
		})
	}
}
