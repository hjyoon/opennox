package legacy

/*
#include "server__ability__ability.h"
*/
import "C"

func abilityResultExportCall4FB960(status uint32) {
	C.nox_xxx_abilGetSuccess_4FB960_ability(C.uint32_t(status))
}

//export nox_xxx_abilGetSuccess_4FB960_ability
func nox_xxx_abilGetSuccess_4FB960_ability(status C.uint32_t) {
	GetClient().AbilityResult4FB960(uint32(status))
}
