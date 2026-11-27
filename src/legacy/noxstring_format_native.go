package legacy

/*
#include <stdio.h>
#include <stdlib.h>
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

type formatValue struct {
	number     int32
	text       unsafe.Pointer
	doubleBits uint64
}

type nativeFormatArgs = C.nox_format_args

type formatArguments struct {
	native *nativeFormatArgs
	values []formatValue
}

func (a *formatArguments) nextValue() formatValue {
	if len(a.values) == 0 {
		return formatValue{}
	}
	value := a.values[0]
	a.values = a.values[1:]
	return value
}

func (a *formatArguments) integer(unsigned bool) int32 {
	if a.native == nil {
		return a.nextValue().number
	}
	return nativeFormatInteger(a.native, unsigned)
}

func (a *formatArguments) stringPointer(wide bool) unsafe.Pointer {
	if a.native == nil {
		return a.nextValue().text
	}
	return nativeFormatStringPointer(a.native, wide)
}

func (a *formatArguments) floatingPoint() uint64 {
	if a.native == nil {
		return a.nextValue().doubleBits
	}
	return nativeFormatDoubleBits(a.native)
}

func legacyFormat[T formatCodeUnit](dst *T, capacity uintptr, format *T, args *C.nox_format_args) int32 {
	return formatWithArgs(dst, capacity, format, formatArguments{native: args})
}

func formatWithArgs[T formatCodeUnit](dst *T, capacity uintptr, format *T, args formatArguments) int32 {
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
			emit(uint16(args.integer(false)))
		case 's', 'S':
			wideString := (code == 's') == wide
			arg := args.stringPointer(wideString)
			if arg == nil {
				for _, ch := range "(null)" {
					emit(uint16(ch))
				}
				continue
			}
			if wideString {
				ptr := (*wchar2_t)(arg)
				for ; *ptr != 0; ptr = wideStringUnit(ptr, 1) {
					emit(uint16(*ptr))
				}
			} else {
				ptr := (*byte)(arg)
				for ; *ptr != 0; ptr = (*byte)(unsafe.Add(unsafe.Pointer(ptr), 1)) {
					emit(uint16(*ptr))
				}
			}
		case 'd', 'i', 'x', 'X', 'o', 'u':
			// The legacy integer service takes a signed 32-bit word for every base.
			number := args.integer(code != 'd' && code != 'i')
			base := 10
			switch code {
			case 'x', 'X':
				base = 16
			case 'o':
				base = 8
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
				nox_format_abort(C.int(code))
			}
			if precision <= 0 {
				precision = 5
			}
			var number [32]C.char
			// The CRT supplies the native rounding and nonfinite spelling
			// for these 31 visible bytes. Pass the original double bits.
			C.nox_format_fixed_bits(C.uint64_t(args.floatingPoint()), &number[0], C.size_t(len(number)), C.int(precision))
			for _, digit := range number {
				if digit == 0 {
					break
				}
				emit(uint16(byte(digit)))
			}
		case '%', '!':
			emit(code)
		default:
			nox_format_abort(C.int(code))
		}
	}
	emit(0)
	return out - 1
}

//export nox_format_abort
func nox_format_abort(code C.int) {
	message := []byte("Unhandled format character: ' '")
	message[len(message)-2] = byte(code)
	C.fwrite(unsafe.Pointer(&message[0]), 1, C.size_t(len(message)), C.stdout)
	C.abort()
}

//export nox_format_narrow_go
func nox_format_narrow_go(dst *C.char, capacity C.size_t, format *C.nox_format_byte, args *C.nox_format_args) C.int {
	return C.int(legacyFormat((*byte)(unsafe.Pointer(dst)), uintptr(capacity), (*byte)(unsafe.Pointer(format)), args))
}

//export nox_format_wide_go
func nox_format_wide_go(dst *C.wchar2_t, capacity C.size_t, format *C.nox_format_word, args *C.nox_format_args) C.int {
	return C.int(legacyFormat((*uint16)(unsafe.Pointer(dst)), uintptr(capacity), (*uint16)(unsafe.Pointer(format)), args))
}
