package legacy

/*
#include "summoned_creature_limit_500d70.h"
*/
import "C"

import "github.com/opennox/opennox/v1/server"

var summonedCreatureLimitCall500D70 = func(owner *server.Object, guideIndex int32) bool {
	return Nox_xxx_checkSummonedCreaturesLimit_500D70(owner, guideIndex)
}

func summonedCreatureLimitExportCall500D70(owner *server.Object, guideIndex int32) int32 {
	return int32(C.nox_xxx_checkSummonedCreaturesLimit_500D70(
		asObjectC(owner),
		C.int32_t(guideIndex),
	))
}

//export nox_xxx_checkSummonedCreaturesLimit_500D70
func nox_xxx_checkSummonedCreaturesLimit_500D70(owner *C.nox_object_t, guideIndex C.int32_t) C.int32_t {
	if summonedCreatureLimitCall500D70(
		asObjectS((*nox_object_t)(owner)),
		int32(guideIndex),
	) {
		return 1
	}
	return 0
}
