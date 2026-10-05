package input

import (
	"image"
	"log/slog"
	"strconv"
	"testing"

	"github.com/opennox/libs/client/keybind"
	seatheadless "github.com/opennox/opennox/v1/client/seat/headless"
)

func TestTextLanguageAccessor(t *testing.T) {
	for _, code := range []int{-1, 0, 1, 2, 3, 4, 5, 6, 7, 8, 9, 255} {
		t.Run(strconv.Itoa(code), func(t *testing.T) {
			h := &Handler{textHandler: textHandler{language: code}}
			if got := h.Language(); got != code {
				t.Fatalf("language = %d, want %d", got, code)
			}
		})
	}
}

func TestTextLanguageSelectedByConstructor(t *testing.T) {
	for code := 0; code <= 9; code++ {
		t.Run(strconv.Itoa(code), func(t *testing.T) {
			st := seatheadless.New(image.Pt(640, 480))
			t.Cleanup(func() { _ = st.Close() })
			h := New(slog.Default(), st, false, code)
			if got := h.Language(); got != code {
				t.Fatalf("constructor language = %d, want %d", got, code)
			}
			if got := h.KeyToWChar(keybind.KeySpace); got != ' ' {
				t.Fatalf("generic scan-code map changed: %#04x", got)
			}
		})
	}
}

func TestTextLanguageTracksLiveSelection(t *testing.T) {
	st := seatheadless.New(image.Pt(640, 480))
	t.Cleanup(func() { _ = st.Close() })
	h := New(slog.Default(), st, false, 0)
	for _, code := range []int{6, 8, 0, 1, 2, 3, 4, 5, 7, 9, 6, 0} {
		t.Run(strconv.Itoa(code), func(t *testing.T) {
			h.SetLanguage(code)
			if got := h.Language(); got != code {
				t.Fatalf("selected language = %d, want %d", got, code)
			}
			if got := h.KeyToWChar(keybind.KeySpace); got != ' ' {
				t.Fatalf("generic scan-code map changed after selection: %#04x", got)
			}
		})
	}
}
