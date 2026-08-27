package legacy

/*
#include "ability_loss_eligible_4f2570.h"
*/
import "C"

import "github.com/opennox/opennox/v1/server"

func randomAbilityLossEligibilityCall4F2570(abilityID int32) int32 {
	return server.RandomAbilityLossEligible4F2570(abilityID)
}

//export sub_4F2570
func sub_4F2570(abilityID C.int32_t) C.int32_t {
	return C.int32_t(randomAbilityLossEligibilityCall4F2570(int32(abilityID)))
}
