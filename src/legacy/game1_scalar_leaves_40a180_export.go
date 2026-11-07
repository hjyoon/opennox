package legacy

/*
#include "GAME1.h"
*/
import "C"

//export sub_40A180
func sub_40A180(value C.short) C.uchar {
	return C.uchar(sub40A180Native(int16(value)))
}

//export sub_40A220
func sub_40A220() C.int {
	return C.int(sub40A220Native())
}

//export sub_40A3C0
func sub_40A3C0(value C.uint) C.uint {
	return C.uint(sub40A3C0Native(uint32(value)))
}

//export sub_40A6A0
func sub_40A6A0(value C.int) C.int {
	return C.int(sub40A6A0Native(int32(value)))
}

//export sub_40A6B0
func sub_40A6B0() C.int {
	return C.int(sub40A6B0Native())
}

//export sub_40A710
func sub_40A710(value C.int) C.int {
	return C.int(sub40A710Native(int32(value)))
}
