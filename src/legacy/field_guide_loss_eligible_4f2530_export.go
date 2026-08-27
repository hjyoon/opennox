package legacy

/*
#include "field_guide_loss_eligible_4f2530.h"
*/
import "C"

import "github.com/opennox/opennox/v1/server"

func randomFieldGuideLossEligibilityCall4F2530(guideID int32) int32 {
	return server.RandomFieldGuideLossEligible4F2530(guideID)
}

//export sub_4F2530
func sub_4F2530(guideID C.int32_t) C.int32_t {
	return C.int32_t(randomFieldGuideLossEligibilityCall4F2530(int32(guideID)))
}
