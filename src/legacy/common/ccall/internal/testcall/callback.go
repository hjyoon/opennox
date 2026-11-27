package testcall

/*
#include <stdint.h>

uintptr_t nox_test_ccall_words(uintptr_t, uintptr_t, uintptr_t, uintptr_t, uintptr_t, uintptr_t);
int nox_test_ccall_ints(int, int, int, int, int, int);
void* nox_test_ccall_pointers(void*, void*, void*, void*, void*, void*);
*/
import "C"

import (
	"math"
	"runtime"
	"unsafe"
)

var wordArgs = [6]uintptr{7, 11, 29, 37, 101, ^uintptr(123)}
var intArgs = [6]int{-11, math.MinInt32, 5, 37, math.MaxInt32, -1234567}

func Words() (unsafe.Pointer, [6]uintptr) { return C.nox_test_ccall_words, wordArgs }
func Ints() (unsafe.Pointer, [6]int)      { return C.nox_test_ccall_ints, intArgs }
func Pointers() unsafe.Pointer            { return C.nox_test_ccall_pointers }

//export nox_test_ccall_words
func nox_test_ccall_words(a, b, c, d, e, f C.uintptr_t) C.uintptr_t {
	if got := [6]uintptr{uintptr(a), uintptr(b), uintptr(c), uintptr(d), uintptr(e), uintptr(f)}; got != wordArgs {
		return 0
	}
	return f
}

//export nox_test_ccall_ints
func nox_test_ccall_ints(a, b, c, d, e, f C.int) C.int {
	if got := [6]int{int(a), int(b), int(c), int(d), int(e), int(f)}; got != intArgs {
		return 0
	}
	return f
}

//export nox_test_ccall_pointers
func nox_test_ccall_pointers(a, b, c, d, e, f unsafe.Pointer) unsafe.Pointer {
	runtime.GC()
	for i, p := range [6]unsafe.Pointer{a, b, c, d, e, f} {
		*(*byte)(p) = byte(i + 1)
	}
	return f
}
