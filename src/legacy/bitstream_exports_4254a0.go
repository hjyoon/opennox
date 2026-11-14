package legacy

/*
#include "bitstream_native_4254a0.h"
*/
import "C"

import "unsafe"

//export sub_4254A0
func sub_4254A0(cursor *C.nox_bit_cursor, data *C.uint8_t) C.int {
	return C.int(bitCursorReadInit4254A0((*bitCursorNative4254A0)(unsafe.Pointer(cursor)), (*byte)(unsafe.Pointer(data))))
}

//export sub_4254C0
func sub_4254C0(cursor *C.nox_bit_cursor) C.bool {
	return C.bool(bitCursorReadNext4254C0((*bitCursorNative4254A0)(unsafe.Pointer(cursor))))
}

//export sub_425500
func sub_425500(cursor *C.nox_bit_cursor, data *C.uint8_t, value C.char) *C.uint8_t {
	return (*C.uint8_t)(unsafe.Pointer(bitCursorWriteInit425500((*bitCursorNative4254A0)(unsafe.Pointer(cursor)), (*byte)(unsafe.Pointer(data)), byte(value))))
}

//export sub_425520
func sub_425520(cursor *C.nox_bit_cursor, value C.char) C.char {
	return C.char(bitCursorWriteNext425520((*bitCursorNative4254A0)(unsafe.Pointer(cursor)), byte(value)))
}

//export sub_425550
func sub_425550(source, target *C.uint8_t, bits C.int) C.int {
	return C.int(bitstreamEncodeNative425550((*byte)(unsafe.Pointer(source)), (*byte)(unsafe.Pointer(target)), int32(bits)))
}
