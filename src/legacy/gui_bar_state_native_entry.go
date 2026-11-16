package legacy

/*
#include "GAME2_1.h"
*/
import "C"

//export sub_470C40
func sub_470C40(mode C.int) C.int {
	return C.int(guiBarStateNative(0, int32(mode), 0))
}

//export nox_xxx_cliSetTotalHealth_470C80
func nox_xxx_cliSetTotalHealth_470C80(current, maximum C.int) C.int {
	return C.int(guiBarStateNative(1, int32(current), int32(maximum)))
}

//export sub_470CB0
func sub_470CB0(current C.int) C.int {
	return C.int(guiBarStateNative(2, int32(current), 0))
}

//export nox_xxx_cliSetManaAndMax_470CE0
func nox_xxx_cliSetManaAndMax_470CE0(current, maximum C.int) C.int {
	return C.int(guiBarStateNative(3, int32(current), int32(maximum)))
}

//export nox_xxx_cliSetMana_470D10
func nox_xxx_cliSetMana_470D10(current C.int) C.int {
	return C.int(guiBarStateNative(4, int32(current), 0))
}

//export nox_xxx_cliGetMana_470DD0
func nox_xxx_cliGetMana_470DD0() C.int { return C.int(guiBarStateNative(5, 0, 0)) }
