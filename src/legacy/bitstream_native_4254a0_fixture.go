package legacy

/*
#include "GAME1_1.h"

typedef int (*nox_bitstream_read_init_t)(nox_bit_cursor*, uint8_t*);
typedef bool (*nox_bitstream_read_next_t)(nox_bit_cursor*);
typedef uint8_t* (*nox_bitstream_write_init_t)(nox_bit_cursor*, uint8_t*, char);
typedef char (*nox_bitstream_write_next_t)(nox_bit_cursor*, char);
typedef int (*nox_bitstream_encode_t)(uint8_t*, uint8_t*, int);

_Static_assert(_Generic(&sub_4254A0, nox_bitstream_read_init_t: 1, default: 0), "native read cursor ABI");
_Static_assert(_Generic(&sub_4254C0, nox_bitstream_read_next_t: 1, default: 0), "native read step ABI");
_Static_assert(_Generic(&sub_425500, nox_bitstream_write_init_t: 1, default: 0), "native write cursor ABI");
_Static_assert(_Generic(&sub_425520, nox_bitstream_write_next_t: 1, default: 0), "native write step ABI");
_Static_assert(_Generic(&sub_425550, nox_bitstream_encode_t: 1, default: 0), "seven-bit encoder ABI");

static uintptr_t nox_bitstream_fixture_call(int kind, void* owner, uint8_t* data, uint8_t value) {
	switch (kind) {
	case 0: return sub_4254A0((nox_bit_cursor*)owner, data);
	case 1: return sub_4254C0((nox_bit_cursor*)owner);
	case 2: return (uintptr_t)sub_425500((nox_bit_cursor*)owner, data, (char)value);
	case 3: return (uint8_t)sub_425520((nox_bit_cursor*)owner, (char)value);
	}
	return 0;
}

static int nox_bitstream_fixture_encode(uint8_t* source, uint8_t* target, int bits) {
	return sub_425550(source, target, bits);
}

static size_t nox_bitstream_fixture_bit_offset(void) {
	return offsetof(nox_bit_cursor, bit);
}
*/
import "C"

import "unsafe"

func bitstreamNativeCursorLayoutC() (size, bitOffset uintptr) {
	return uintptr(C.sizeof_nox_bit_cursor), uintptr(C.nox_bitstream_fixture_bit_offset())
}

func bitstreamNativeCursorCallC(kind string, owner, data unsafe.Pointer, value byte) uintptr {
	var code C.int
	switch kind {
	case "read_init":
		code = 0
	case "read_next":
		code = 1
	case "write_init":
		code = 2
	case "write_next":
		code = 3
	default:
		panic("unknown bitstream fixture operation")
	}
	return uintptr(C.nox_bitstream_fixture_call(code, owner, (*C.uint8_t)(data), C.uint8_t(value)))
}

func bitstreamNativeEncodeC(source, target unsafe.Pointer, bits int32) int32 {
	return int32(C.nox_bitstream_fixture_encode((*C.uint8_t)(source), (*C.uint8_t)(target), C.int(bits)))
}
