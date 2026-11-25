package legacy

/*
#include "noxstring.h"
#include <ctype.h>
#include <errno.h>
#include <stdlib.h>
#include <wctype.h>

typedef const wchar2_t nox_const_string_unit_t;
typedef const char nox_const_string_byte_t;

struct nox_strtol_dialect {
	int auto_binary;
	int base2_binary;
	int auto_reject_bare_hex;
	int base16_reject_bare_hex;
	int missing_digits_errno;
};

static struct nox_strtol_dialect nox_strtol_dialect(void) {
	const char binary[] = "0b1";
	const char bare_hex[] = "0x";
	struct nox_strtol_dialect result;
	char* end;
	int saved = errno;
	strtol(binary, &end, 0);
	result.auto_binary = end == binary + 3;
	strtol(binary, &end, 2);
	result.base2_binary = end == binary + 3;
	strtol(bare_hex, &end, 0);
	result.auto_reject_bare_hex = end == bare_hex;
	strtol(bare_hex, &end, 16);
	result.base16_reject_bare_hex = end == bare_hex;
	errno = 0;
	strtol("!", &end, 10);
	result.missing_digits_errno = errno;
	errno = saved;
	return result;
}

static void nox_strtol_set_errno(int value) {
	errno = value;
}

// Invalid bases can invoke the host CRT's invalid-parameter handler.
static void nox_strtol_invalid_base(int base) {
	strtol("", NULL, base);
}
*/
import "C"

import (
	"strconv"
	"unsafe"

	"github.com/opennox/opennox/v1/server"
)

var legacyLongDialect = C.nox_strtol_dialect()

// Keep caller-owned UTF-16 buffers within Go when the caller is native Go.
func copyWideString(dst, src *wchar2_t) *wchar2_t {
	return nox_wcscpy(dst, (*C.nox_const_string_unit_t)(src))
}

func appendWideString(dst, src *wchar2_t) {
	nox_wcscat(dst, (*C.nox_const_string_unit_t)(src))
}

func parseLegacyDecimal32(text *C.char) C.int {
	value, _ := server.ParseCRTDecimal32(GoString(text))
	return C.int(value)
}

func parseLegacyWideDecimal(text *wchar2_t) C.long {
	value, _ := scanLegacyLongDecimal((*uint16)(unsafe.Pointer(text)))
	return value
}

// scanLegacyLongDecimal reads an ASCII decimal prefix using the target's
// signed C long range. It returns the number of code units consumed.
func scanLegacyLongDecimal[T formatCodeUnit](text *T) (C.long, uintptr) {
	value, consumed, _ := scanLegacyLong(text, 10, legacyLongPrefixes{})
	return value, consumed
}

func legacyLongDigit[T formatCodeUnit](ch T) uint64 {
	if ch >= '0' && ch <= '9' {
		return uint64(ch - '0')
	}
	ch |= 'a' - 'A'
	if ch >= 'a' && ch <= 'z' {
		return uint64(ch-'a') + 10
	}
	return 36
}

type legacyLongPrefixes struct {
	binary, rejectBareHex bool
}

func scanLegacyLong[T formatCodeUnit](text *T, base int, prefixes legacyLongPrefixes) (C.long, uintptr, bool) {
	unitAt := func(index uintptr) T {
		return *(*T)(unsafe.Add(unsafe.Pointer(text), index*unsafe.Sizeof(*text)))
	}
	var offset uintptr
	ch := unitAt(offset)
	for ch == ' ' || ch >= '\t' && ch <= '\r' {
		offset++
		ch = unitAt(offset)
	}
	negative := ch == '-'
	if ch == '+' || negative {
		offset++
		ch = unitAt(offset)
	}
	if ch == '0' {
		prefix := unitAt(offset+1) | ('a' - 'A')
		if (base == 0 || base == 16) && prefix == 'x' {
			if legacyLongDigit(unitAt(offset+2)) < 16 {
				base, offset = 16, offset+2
			} else if prefixes.rejectBareHex {
				return 0, 0, false
			}
		}
		if prefixes.binary && (base == 0 || base == 2) && prefix == 'b' && legacyLongDigit(unitAt(offset+2)) < 2 {
			base, offset = 2, offset+2
		}
		if base == 0 {
			base = 8
		}
	}
	if base == 0 {
		base = 10
	}
	firstDigit := offset
	const positiveLimit = (uint64(1) << (unsafe.Sizeof(C.long(0))*8 - 1)) - 1
	limit := positiveLimit
	if negative {
		limit++
	}
	radix := uint64(base)
	cutoff, lastDigit := limit/radix, limit%radix
	var value uint64
	var overflow bool
	for {
		digit := legacyLongDigit(unitAt(offset))
		if digit >= radix {
			break
		}
		if value > cutoff || value == cutoff && digit > lastDigit {
			value = limit
			overflow = true
		} else {
			value = radix*value + digit
		}
		offset++
	}
	if offset == firstDigit {
		return 0, 0, false
	}
	if negative {
		return -C.long(value), offset, overflow
	}
	return C.long(value), offset, overflow
}

// copyLegacyCString applies the legacy formatter's code-unit conversion to
// one string, with a terminator within the destination.
func copyLegacyCString[D, S formatCodeUnit](dst []D, src *S) {
	if len(dst) == 0 {
		return
	}
	limit := len(dst) - 1
	if src == nil {
		const nullText = "(null)"
		n := min(limit, len(nullText))
		for i := range n {
			dst[i] = D(nullText[i])
		}
		dst[n] = 0
		return
	}
	stride := unsafe.Sizeof(*src)
	var n int
	for n < limit {
		unit := *(*S)(unsafe.Add(unsafe.Pointer(src), uintptr(n)*stride))
		if unit == 0 {
			break
		}
		dst[n] = D(unit)
		n++
	}
	dst[n] = 0
}

func wideStringUnit(text *wchar2_t, index uintptr) *wchar2_t {
	return (*wchar2_t)(unsafe.Add(unsafe.Pointer(text), index*2))
}

//export nox_itow
func nox_itow(value C.int, dst *C.wchar2_t, radix C.int) *C.wchar2_t {
	text := strconv.FormatInt(int64(value), int(radix))
	units := unsafe.Slice(dst, len(text)+1)
	for i := range len(text) {
		units[i] = C.wchar2_t(text[i])
	}
	units[len(text)] = 0
	return dst
}

//export nox_wcscat
func nox_wcscat(dst *C.wchar2_t, src *C.nox_const_string_unit_t) *C.wchar2_t {
	if src != nil {
		end := wideStringUnit(dst, uintptr(nox_wcslen(dst)))
		nox_wcscpy(end, src)
	}
	return dst
}

//export nox_wcschr
func nox_wcschr(text *C.nox_const_string_unit_t, unit C.wchar2_t) *C.wchar2_t {
	for ptr := text; *ptr != 0; ptr = wideStringUnit(ptr, 1) {
		if *ptr == unit {
			return ptr
		}
	}
	return nil
}

//export nox_wcscmp
func nox_wcscmp(a, b *C.nox_const_string_unit_t) C.int {
	switch {
	case a == b:
		return 0
	case a == nil:
		return -1
	case b == nil:
		return 1
	}
	for i := uintptr(0); ; i++ {
		left, right := *wideStringUnit(a, i), *wideStringUnit(b, i)
		if left == 0 || left != right {
			return C.int(left) - C.int(right)
		}
	}
}

//export nox_wcscpy
func nox_wcscpy(dst *C.wchar2_t, src *C.nox_const_string_unit_t) *C.wchar2_t {
	for i := uintptr(0); ; i++ {
		word := *wideStringUnit(src, i)
		*wideStringUnit(dst, i) = word
		if word == 0 {
			return dst
		}
	}
}

//export nox_wcslen
func nox_wcslen(text *C.nox_const_string_unit_t) C.size_t {
	for i := uintptr(0); ; i++ {
		if *wideStringUnit(text, i) == 0 {
			return C.size_t(i)
		}
	}
}

//export nox_wcsncpy
func nox_wcsncpy(dst *C.wchar2_t, src *C.nox_const_string_unit_t, count C.size_t) *C.wchar2_t {
	words := unsafe.Slice(dst, int(count))
	for i := range words {
		word := *wideStringUnit(src, uintptr(i))
		if word == 0 {
			clear(words[i:])
			break
		}
		words[i] = word
	}
	return dst
}

//export nox_wcsspn
func nox_wcsspn(text, accept *C.nox_const_string_unit_t) C.size_t {
	for i := uintptr(0); ; i++ {
		word := *wideStringUnit(text, i)
		if word == 0 || nox_wcschr(accept, word) == nil {
			return C.size_t(i)
		}
	}
}

var wideStringTokenNext *wchar2_t

//export nox_wcstok
func nox_wcstok(text *C.wchar2_t, delimiters *C.nox_const_string_unit_t) *C.wchar2_t {
	if text == nil {
		text = wideStringTokenNext
	}
	if text == nil {
		return nil
	}
	text = wideStringUnit(text, uintptr(nox_wcsspn(text, delimiters)))
	if *text == 0 {
		wideStringTokenNext = nil
		return nil
	}
	for pos := text; ; pos = wideStringUnit(pos, 1) {
		if *pos == 0 {
			wideStringTokenNext = pos
			return text
		}
		if nox_wcschr(delimiters, *pos) != nil {
			*pos = 0
			wideStringTokenNext = wideStringUnit(pos, 1)
			return text
		}
	}
}

//export _nox_wcsicmp
func _nox_wcsicmp(a, b *C.nox_const_string_unit_t) C.int {
	for i := uintptr(0); ; i++ {
		left, right := *wideStringUnit(a, i), *wideStringUnit(b, i)
		if left != right {
			diff := C.int(C.towlower(C.wint_t(left))) - C.int(C.towlower(C.wint_t(right)))
			if diff != 0 {
				return diff
			}
		}
		if left == 0 {
			return 0
		}
	}
}

func compareLegacyStringBytes(a, b unsafe.Pointer, count int) C.int {
	if count == 0 {
		return 0
	}
	for i := 0; ; i++ {
		left := *(*uint8)(unsafe.Add(a, i))
		right := *(*uint8)(unsafe.Add(b, i))
		if left != right {
			lowerLeft, lowerRight := C.tolower(C.int(left)), C.tolower(C.int(right))
			switch {
			case lowerLeft < lowerRight:
				return -1
			case lowerLeft > lowerRight:
				return 1
			}
		}
		if left == 0 || count > 0 && i+1 == count {
			return 0
		}
	}
}

//export nox_strcmpi
func nox_strcmpi(a, b *C.nox_const_string_byte_t) C.int {
	return compareLegacyStringBytes(unsafe.Pointer(a), unsafe.Pointer(b), -1)
}

//export nox_strnicmp
func nox_strnicmp(a, b *C.nox_const_string_byte_t, count C.int) C.int {
	return compareLegacyStringBytes(unsafe.Pointer(a), unsafe.Pointer(b), int(count))
}

//export nox_wcstol
func nox_wcstol(text *C.nox_const_string_unit_t, end **C.wchar2_t, base C.int) C.long {
	var value C.long
	var consumed uintptr
	if base != 0 && (base < 2 || base > 36) {
		C.nox_strtol_invalid_base(base)
	} else {
		prefixes := legacyLongPrefixes{
			binary: base == 0 && legacyLongDialect.auto_binary != 0 ||
				base == 2 && legacyLongDialect.base2_binary != 0,
			rejectBareHex: base == 0 && legacyLongDialect.auto_reject_bare_hex != 0 ||
				base == 16 && legacyLongDialect.base16_reject_bare_hex != 0,
		}
		var overflow bool
		value, consumed, overflow = scanLegacyLong((*uint16)(unsafe.Pointer(text)), int(base), prefixes)
		if overflow {
			C.nox_strtol_set_errno(C.ERANGE)
		} else if consumed == 0 && legacyLongDialect.missing_digits_errno != 0 {
			C.nox_strtol_set_errno(legacyLongDialect.missing_digits_errno)
		}
	}
	if end != nil {
		*end = wideStringUnit(text, consumed)
	}
	return value
}
