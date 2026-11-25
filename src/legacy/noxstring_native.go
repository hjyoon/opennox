package legacy

/*
#include "noxstring.h"
#include <ctype.h>
#include <stdlib.h>
#include <wctype.h>

typedef const wchar2_t nox_const_string_unit_t;
typedef const char nox_const_string_byte_t;
*/
import "C"

import (
	"strconv"
	"unsafe"

	"github.com/opennox/opennox/v1/legacy/common/alloc"
)

// Keep caller-owned UTF-16 buffers within Go when the caller is native Go.
func copyWideString(dst, src *wchar2_t) *wchar2_t {
	return nox_wcscpy(dst, (*C.nox_const_string_unit_t)(src))
}

func appendWideString(dst, src *wchar2_t) {
	nox_wcscat(dst, (*C.nox_const_string_unit_t)(src))
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
	n := int(nox_wcslen(text))
	raw, free := alloc.Calloc(n+1, 1)
	defer free()
	ascii := unsafe.Slice((*byte)(raw), n)
	for i := range ascii {
		word := *wideStringUnit(text, uintptr(i))
		if word < 0x80 {
			ascii[i] = byte(word)
		} else {
			ascii[i] = 0x7f
		}
	}
	first := (*C.char)(raw)
	last := first
	value := C.strtol(first, &last, base)
	if end != nil {
		*end = wideStringUnit(text, uintptr(unsafe.Pointer(last))-uintptr(raw))
	}
	return value
}
