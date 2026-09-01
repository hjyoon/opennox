package legacy

/*
#include <stdint.h>

#include "coop_ability_state_set_4fc670.h"
*/
import "C"

func coopAbilityStateSetExportCall4FC670(value int32) int32 {
	return int32(C.sub_4FC670(C.int32_t(value)))
}

//export sub_4FC670
func sub_4FC670(value C.int32_t) C.int32_t {
	return C.int32_t(GetServer().S().SetCoopAbilityState4FC670(int32(value)))
}
