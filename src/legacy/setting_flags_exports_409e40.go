package legacy

/*
#include <stdint.h>
#include "GAME1.h"
*/
import "C"

//export sub_409E40
func sub_409E40(argument C.int) C.int {
	return C.int(sub409E40Native(int32(argument)))
}

//export sub_409E70
func sub_409E70(argument C.int) C.int {
	return C.int(sub409E70Native(int32(argument)))
}

//export sub_409EC0
func sub_409EC0(argument C.int) C.int {
	return C.int(sub409EC0Native(int32(argument)))
}

//export sub_409EF0
func sub_409EF0(argument C.int) C.int {
	return C.int(sub409EF0Native(int32(argument)))
}

//export sub_409F40
func sub_409F40(argument C.int) C.int {
	return C.int(sub409F40Native(int32(argument)))
}
