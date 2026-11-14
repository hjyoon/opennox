package legacy

/*
#include "GAME1_1.h"
extern uint32_t dword_587000_66116;
*/
import "C"

// The extracted C scalar is independent of the raw blob at the same PE32
// address. Keep its live owner while the table retains packed DWORD stride.
func buffIterationCountScalar424CB0() int32 {
	return int32(C.dword_587000_66116)
}

//export sub_424CB0
func sub_424CB0(unit *C.nox_object_t) C.char {
	return C.char(buffIterationCountNative424CB0(asObjectS((*nox_object_t)(unit))))
}

//export sub_424D00
func sub_424D00() C.int {
	return C.int(buffIterationFirstNative424D00())
}

//export sub_424D20
func sub_424D20(buff C.int) C.int {
	return C.int(buffIterationNextNative424D20(int32(buff)))
}
