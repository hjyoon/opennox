package legacy

import (
	"slices"
	"testing"
	"unsafe"

	"github.com/opennox/opennox/v1/legacy/common/alloc"
)

func TestLegacyCStringCopyRetainsCodeUnitsAndBounds(t *testing.T) {
	wide := [...]uint16{0x0080, 0x0100, 0xd83d, 0x00e9, 0}
	narrow := [...]byte{0xdd, 0xdd, 0xdd, 0xdd, 0xdd, 0xdd}
	copyLegacyCString(narrow[:5], &wide[0])
	if want := [6]byte{0x80, 0, 0x3d, 0xe9, 0, 0xdd}; narrow != want {
		t.Fatalf("wide-to-narrow code units = %x, want %x", narrow, want)
	}
	narrowSource := [...]byte{0x80, 0xff, 0x41}
	wideDestination := [...]uint16{0xdddd, 0xdddd, 0xdddd, 0xdddd}
	copyLegacyCString(wideDestination[:3], &narrowSource[0])
	if want := [4]uint16{0x80, 0xff, 0, 0xdddd}; wideDestination != want {
		t.Fatalf("bounded narrow-to-wide code units = %x, want %x", wideDestination, want)
	}
	null := make([]byte, len("(null)")+1)
	copyLegacyCString(null, (*uint16)(nil))
	if got := string(null); got != "(null)\x00" {
		t.Fatalf("nil wide string = %q, want a terminated (null)", got)
	}
}

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

func TestLegacyWideDecimalMatchesRetainedC(t *testing.T) {
	for _, input := range []string{
		"", "+", " - 1", " \t\n\v\f\r+17tail", "0xff", "42\x00-1",
		"１２", "12９34", "2147483648", "-2147483649", "9223372036854775807",
		"9223372036854775808suffix", "-9223372036854775808", "-9223372036854775809",
		"18446744073709551616", "-18446744073709551616",
	} {
		t.Run(input, func(t *testing.T) {
			words, free := alloc.CString16High(input)
			defer free()
			text := (*wchar2_t)(unsafe.Pointer(words))
			if got, want := parseLegacyWideDecimal(text), nox_wcstol(text, nil, 10); got != want {
				t.Errorf("decimal prefix result = %d, want %d", got, want)
			}
		})
	}
}
