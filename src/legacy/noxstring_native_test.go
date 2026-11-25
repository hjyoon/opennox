package legacy

import (
	"slices"
	"testing"
	"unsafe"

	"github.com/opennox/opennox/v1/legacy/common/alloc"
)

func TestLegacyWideStringKeepsCodeUnitsAndPointerOffsets(t *testing.T) {
	previous := wideStringTokenNext
	t.Cleanup(func() { wideStringTokenNext = previous })
	words, free := alloc.CloneSliceHigh([]uint16{' ', ' ', 0xd800, ':', 'Q', 0})
	defer free()
	delimiters, freeDelimiters := CWString(" :")
	defer freeDelimiters()
	text := (*wchar2_t)(unsafe.Pointer(&words[0]))
	if got := nox_wcslen(text); got != 5 {
		t.Fatalf("wide length = %d, want 5", got)
	}
	if got := nox_wcschr(text, ':'); got != wideStringUnit(text, 3) {
		t.Fatalf("wide search = %p, want %p", got, wideStringUnit(text, 3))
	}
	if nox_wcschr(text, 0) != nil {
		t.Fatal("wide search included the terminator")
	}
	if first := nox_wcstok(text, delimiters); first != wideStringUnit(text, 2) || *first != 0xd800 {
		t.Fatalf("first token = %p, want original raw UTF-16 at %p", first, wideStringUnit(text, 2))
	}
	if second := nox_wcstok(nil, delimiters); second != wideStringUnit(text, 4) || *second != 'Q' {
		t.Fatalf("second token = %p, want %p", second, wideStringUnit(text, 4))
	}
	if nox_wcstok(nil, delimiters) != nil {
		t.Fatal("wide tokenizer continued past the end")
	}
	if !slices.Equal(words, []uint16{' ', ' ', 0xd800, 0, 'Q', 0}) {
		t.Fatalf("tokenizer wrote unexpected UTF-16 units: %x", words)
	}
}

func TestLegacyWideStringNumberLeavesEndAtFirstNonASCIIUnit(t *testing.T) {
	words, free := alloc.CloneSliceHigh([]uint16{' ', '-', '0', 'x', '2', 'A', 0xd800, 0})
	defer free()
	text := (*wchar2_t)(unsafe.Pointer(&words[0]))
	var end *wchar2_t
	if got := nox_wcstol(text, &end, 0); got != -42 || end != wideStringUnit(text, 6) {
		t.Fatalf("parse result = %d, %p, want -42, %p", got, end, wideStringUnit(text, 6))
	}
}
