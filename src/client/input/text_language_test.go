package input

import (
	"strconv"
	"testing"
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
