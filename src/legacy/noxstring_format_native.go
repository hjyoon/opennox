package legacy

/*
#include "noxstring_format_abi.h"
*/
import "C"

import (
	"strconv"
	"unsafe"
)

type formatCodeUnit interface {
	byte | uint16
}

func legacyFormat[T formatCodeUnit](dst *T, capacity uintptr, format *T, args *C.nox_format_args) int32 {
	unitSize := unsafe.Sizeof(*new(T))
	wide := unitSize == 2
	var out int32
	emit := func(v uint16) {
		if dst != nil && uintptr(out) < capacity {
			*(*T)(unsafe.Add(unsafe.Pointer(dst), uintptr(out)*unitSize)) = T(v)
		}
		out++
	}
	var index uintptr
	next := func() uint16 {
		value := *(*T)(unsafe.Add(unsafe.Pointer(format), index*unitSize))
		index++
		return uint16(value)
	}
	for code := next(); code != 0; code = next() {
		if code != '%' {
			emit(code)
			continue
		}
		code = next()
		var flag uint16
		switch code {
		case '+', '-', '0':
			flag, code = code, next()
		}
		width, precision := int32(-1), int32(-1)
		if code >= '0' && code <= '9' {
			width = 0
			for code >= '0' && code <= '9' {
				width = width*10 + int32(code-'0')
				code = next()
			}
		}
		if code == '.' {
			precision, code = 0, next()
			for code >= '0' && code <= '9' {
				precision = precision*10 + int32(code-'0')
				code = next()
			}
		}
		switch code {
		case 'c':
			emit(uint16(C.nox_format_arg_int(args)))
		case 's', 'S':
			if (code == 's') == wide {
				ptr := C.nox_format_arg_wide(args)
				if ptr == nil {
					for _, ch := range "(null)" {
						emit(uint16(ch))
					}
					continue
				}
				for ; *ptr != 0; ptr = wideStringUnit(ptr, 1) {
					emit(uint16(*ptr))
				}
			} else {
				ptr := (*byte)(unsafe.Pointer(C.nox_format_arg_narrow(args)))
				if ptr == nil {
					for _, ch := range "(null)" {
						emit(uint16(ch))
					}
					continue
				}
				for ; *ptr != 0; ptr = (*byte)(unsafe.Add(unsafe.Pointer(ptr), 1)) {
					emit(uint16(*ptr))
				}
			}
		case 'd', 'i', 'x', 'X', 'o', 'u':
			var number int32
			base := 10
			if code == 'd' || code == 'i' {
				number = int32(C.nox_format_arg_int(args))
			} else {
				// The legacy integer service takes a signed 32-bit word for every base.
				number = int32(C.nox_format_arg_uint(args))
				switch code {
				case 'x', 'X':
					base = 16
				case 'o':
					base = 8
				}
			}
			var scratch [32]byte
			digits := strconv.AppendInt(scratch[:0], int64(number), base)
			digitCount := int32(len(digits))
			paddedWidth := digitCount
			if precision > 0 {
				paddedWidth = precision
			}
			pad := uint16(' ')
			if flag == '0' {
				pad = '0'
			}
			for i := int32(0); i < width-paddedWidth; i++ {
				emit(pad)
			}
			for i := int32(0); i < precision-digitCount; i++ {
				emit('0')
			}
			for _, digit := range digits {
				if code == 'X' && digit >= 'a' && digit <= 'f' {
					digit -= 'a' - 'A'
				}
				emit(uint16(digit))
			}
		case 'f':
			if !wide {
				C.nox_format_abort(C.int(code))
			}
			if precision <= 0 {
				precision = 5
			}
			var number [32]C.char
			// Use the host's floating-point formatting and retain the legacy
			// temporary buffer size before widening the formatted bytes.
			C.nox_format_arg_fixed(args, &number[0], C.size_t(len(number)), C.int(precision))
			for _, digit := range number {
				if digit == 0 {
					break
				}
				emit(uint16(byte(digit)))
			}
		case '%', '!':
			emit(code)
		default:
			C.nox_format_abort(C.int(code))
		}
	}
	emit(0)
	return out - 1
}

//export nox_format_narrow_go
func nox_format_narrow_go(dst *C.char, capacity C.size_t, format *C.nox_format_byte, args *C.nox_format_args) C.int {
	return C.int(legacyFormat((*byte)(unsafe.Pointer(dst)), uintptr(capacity), (*byte)(unsafe.Pointer(format)), args))
}

//export nox_format_wide_go
func nox_format_wide_go(dst *C.wchar2_t, capacity C.size_t, format *C.nox_format_word, args *C.nox_format_args) C.int {
	return C.int(legacyFormat((*uint16)(unsafe.Pointer(dst)), uintptr(capacity), (*uint16)(unsafe.Pointer(format)), args))
}
