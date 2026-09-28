package legacy

/*
#include "defs.h"
#include "GAME2_3.h"
#include "GAME3_1.h"

static void nox_xxx_cliSummonOnDieOrBanish_go(int net_code, int silent) {
	nox_xxx_cliSummonOnDieOrBanish_4C3140(net_code, silent ? (void*)(uintptr_t)1 : NULL);
}
*/
import "C"

func Nox_xxx_cliSummonCreat_4C2E50(netCode, typeID int, silent bool) {
	var silentC C.int
	if silent {
		silentC = 1
	}
	C.nox_xxx_cliSummonCreat_4C2E50(C.int(netCode), C.int(typeID), silentC)
}

func Nox_xxx_cliSummonOnDieOrBanish_4C3140(netCode int, silent bool) {
	var silentC C.int
	if silent {
		silentC = 1
	}
	C.nox_xxx_cliSummonOnDieOrBanish_go(C.int(netCode), silentC)
}

func Sub_495060(netCode, x, y int) int {
	return int(C.sub_495060(C.int(netCode), C.short(x), C.short(y)))
}

func Sub_4950C0(netCode int) int {
	return int(C.sub_4950C0(C.int(netCode)))
}
