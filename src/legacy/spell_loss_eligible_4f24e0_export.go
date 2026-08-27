package legacy

/*
#include "spell_loss_eligible_4f24e0.h"
*/
import "C"

import "github.com/opennox/opennox/v1/server"

func randomSpellLossEligibilityCall4F24E0(spellID int32) int32 {
	return server.RandomSpellLossEligible4F24E0(spellID)
}

//export sub_4F24E0
func sub_4F24E0(spellID C.int32_t) C.int32_t {
	return C.int32_t(randomSpellLossEligibilityCall4F24E0(int32(spellID)))
}
