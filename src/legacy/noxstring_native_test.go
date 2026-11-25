package legacy

import (
	"math"
	"runtime"
	"slices"
	"strconv"
	"strings"
	"testing"
	"unsafe"

	"github.com/opennox/opennox/v1/legacy/common/alloc"
	"github.com/opennox/opennox/v1/legacy/internal/cfixture"
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

func TestLegacyWideIntegerMatchesRetainedC(t *testing.T) {
	compare := func(base int, input string) {
		words, free := alloc.CString16High(input)
		defer free()
		text := unsafe.Pointer(words)
		for _, withEnd := range []bool{true, false} {
			got, want := cfixture.CompareWideInteger(text, base, withEnd)
			if got != want {
				t.Errorf("base %d withEnd=%t input %q: got %+v, want %+v", base, withEnd, input, got, want)
			}
			if base == 10 && int64(parseLegacyWideDecimal((*wchar2_t)(text))) != want.Value {
				t.Errorf("decimal prefix for %q disagreed with %+v", input, want)
			}
		}
	}
	inputs := []string{
		"", "+", " - 1", " \t\n\v\f\r+17tail", "0", "08", "-0129!",
		"0xff", "-0XfF!", "0x", "0x+3", " -0X", "\t+0Xz", "0x-1", "0b", "+0b011", "0B011!", "-0b2",
		"Z.", "+xyz", "42\x00-1", "１２", "12９34", "\u00a012", "-3🙂4",
	}
	for base := 0; base <= 36; base++ {
		if base == 1 {
			continue
		}
		for _, input := range inputs {
			compare(base, input)
		}
		digitsBase := base
		if digitsBase == 0 {
			digitsBase = 10
		}
		for _, magnitude := range []uint64{math.MaxInt32, 1 << 31, 1<<31 + 1, math.MaxInt64, 1 << 63, 1<<63 + 1, math.MaxUint64} {
			digits := strconv.FormatUint(magnitude, digitsBase)
			compare(base, digits+"!")
			compare(base, "-"+digits+"!")
		}
		compare(base, strings.Repeat(strconv.FormatInt(int64(digitsBase-1), digitsBase), 160)+"!")
	}
	if runtime.GOOS != "windows" {
		for _, base := range []int{-1, 1, 37} {
			compare(base, "123")
		}
	}
}

func TestLegacyWideIntegerBinaryPrefixDialect(t *testing.T) {
	for _, base := range []int{0, 2} {
		for _, tc := range []struct {
			text     string
			value    int64
			consumed uintptr
		}{
			{"0b1012", 5, 5},
			{" -0B101!", -5, 7},
			{"0b", 0, 1},
			{"0B2", 0, 1},
		} {
			words, free := alloc.CString16High(tc.text)
			value, consumed, overflow := scanLegacyLong(words, base, legacyLongPrefixes{binary: true})
			free()
			if int64(value) != tc.value || consumed != tc.consumed || overflow {
				t.Fatalf("base %d input %q: got %d at %d (overflow=%t), want %d at %d", base, tc.text, value, consumed, overflow, tc.value, tc.consumed)
			}
		}
	}
}

func TestLegacyWideIntegerBareHexDialect(t *testing.T) {
	words, free := alloc.CString16High(" -0Xz")
	defer free()
	for _, base := range []int{0, 16} {
		for _, test := range []struct {
			reject   bool
			consumed uintptr
		}{{false, 3}, {true, 0}} {
			value, consumed, overflow := scanLegacyLong(words, base, legacyLongPrefixes{rejectBareHex: test.reject})
			if value != 0 || consumed != test.consumed || overflow {
				t.Fatalf("base %d bare prefix rejected=%t: got %d at %d (overflow=%t), want 0 at %d",
					base, test.reject, value, consumed, overflow, test.consumed)
			}
		}
	}
}

func TestThingIntegerReadsDecimalPrefixesAndAdvancesPastOverflow(t *testing.T) {
	const large = "18446744073709551616"
	text := CString(" \t-17+12\n" + large + " -9 no-number")
	defer StrFree(text)
	wide, free := CWString(large)
	defer free()
	_, reference := cfixture.CompareWideInteger(unsafe.Pointer(wide), 10, false)
	overflow := int32(reference.Value)
	arg := text
	for _, want := range []struct {
		value int32
		rest  string
	}{
		{-17, "+12\n" + large + " -9 no-number"},
		{12, "\n" + large + " -9 no-number"},
		{overflow, " -9 no-number"},
		{-9, " no-number"},
	} {
		got, next := scanThingInt(arg)
		if got != want.value || GoString(next) != want.rest {
			t.Fatalf("scan at %q = %d, %q, want %d, %q", GoString(arg), got, GoString(next), want.value, want.rest)
		}
		arg = next
	}
	for _, input := range []string{" no-number", " \v-", "\x00+12"} {
		invalid := CString(input)
		got, next := scanThingInt(invalid)
		StrFree(invalid)
		if got != 0 || next != nil {
			t.Errorf("scan at %q = %d, %p, want 0, nil", input, got, next)
		}
	}
}
