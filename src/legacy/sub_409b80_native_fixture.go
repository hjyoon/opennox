//go:build darwin && arm64

package legacy

/*
#include <stdint.h>
#include <stddef.h>
#include <string.h>
#include "GAME1.h"
extern unsigned char byte_5D4594[2598284];
static void* sub409b80_fixture_blob(void) { return &byte_5D4594[3452]; }
static char* sub409b80_fixture_C_caller(void) { return sub_409B80(); }
static uint32_t sub409b80_fixture_bad_u32(void) {
    return (uint32_t)(uintptr_t)sub_409B80();
}
static size_t sub409b80_fixture_C_copy(void* destination, size_t capacity) {
    const char* actual = sub_409B80();
    size_t length = strlen(actual) + 1;
    if (length <= capacity) memcpy(destination, actual, length);
    return length;
}
*/
import "C"

import (
	"runtime"
	"unsafe"
)

func sub409B80FixtureBlob() unsafe.Pointer { return C.sub409b80_fixture_blob() }
func sub409B80FixtureC() unsafe.Pointer    { return unsafe.Pointer(C.sub409b80_fixture_C_caller()) }
func sub409B80FixtureBadU32() uint32       { return uint32(C.sub409b80_fixture_bad_u32()) }
func sub409B80FixtureSet(raw []byte) uint32 {
	n := C.sub_409B50((*C.char)(unsafe.Pointer(&raw[0])))
	runtime.KeepAlive(raw)
	return uint32(n)
}
func sub409B80FixtureCCopy(out []byte) int {
	n := C.sub409b80_fixture_C_copy(unsafe.Pointer(&out[0]), C.size_t(len(out)))
	runtime.KeepAlive(out)
	return int(n)
}
