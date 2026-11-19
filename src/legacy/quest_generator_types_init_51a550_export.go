package legacy

/*
#include "GAME4_1.h"
*/
import "C"
import "unsafe"

//export sub_51A550
func sub_51A550() *C.char {
	return (*C.char)(unsafe.Pointer(questGeneratorTypesInitNative51A550(questGeneratorTypesInitDepsFactory51A550())))
}

func questGeneratorTypesInitCEntry51A550() *byte {
	return (*byte)(unsafe.Pointer(C.sub_51A550()))
}
