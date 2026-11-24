package legacy

/*
#include "defs.h"
#include "GAME2_3.h"
#include "GAME3_1.h"
*/
import "C"

func Nox_xxx_cliSummonCreat_4C2E50(netCode, typeID int, silent bool) {
	var silentC C.int
	if silent {
		silentC = 1
	}
	nox_xxx_cliSummonCreat_4C2E50(C.int(netCode), C.int(typeID), silentC)
}

func Nox_xxx_cliSummonOnDieOrBanish_4C3140(netCode int, silent bool) {
	removeSummonCreature(C.int(netCode), silent)
}

func Sub_495060(netCode, x, y int) int {
	return int(C.sub_495060(C.int(netCode), C.short(x), C.short(y)))
}

func Sub_4950C0(netCode int) int {
	return int(C.sub_4950C0(C.int(netCode)))
}
