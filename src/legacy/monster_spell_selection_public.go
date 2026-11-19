package legacy

/*
#include "GAME4_3.h"
*/
import "C"

//export nox_xxx_monsterBuffSelf_540B90
func nox_xxx_monsterBuffSelf_540B90(unit *nox_object_t) C.int {
	return C.int(bool2int(GetServer().S().MonsterFightBuffSelf540B90(asObjectS(unit))))
}

//export nox_xxx_mobCastRelated2_540D90
func nox_xxx_mobCastRelated2_540D90(unit, target *nox_object_t) C.int {
	canSummon := monsterFightSchedulerRuntime531B40().CanSummon
	return C.int(bool2int(GetServer().S().MonsterFightCastRelated540D90(asObjectS(unit), asObjectS(target), canSummon)))
}

//export nox_xxx_monsterCastOffensive_540F20
func nox_xxx_monsterCastOffensive_540F20(unit, target *nox_object_t) C.int {
	return C.int(bool2int(GetServer().S().MonsterFightCastOffensive540F20(asObjectS(unit), asObjectS(target))))
}

//export nox_xxx_mobCastRelated_541050
func nox_xxx_mobCastRelated_541050(unit *nox_object_t) C.int {
	return C.int(bool2int(GetServer().S().MonsterFleeCastRelated541050(asObjectS(unit))))
}
