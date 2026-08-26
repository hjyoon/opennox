package legacy

/*
#include "reward_field_guide_4f0d20.h"
*/
import "C"

import "github.com/opennox/opennox/v1/server"

func rewardFieldGuideCall4F0D20(
	s *server.Server,
	marker *server.Object,
	stage uint32,
) *server.Object {
	return s.RewardFieldGuide4F0D20(marker, stage)
}

//export nox_xxx_rewardFieldGuide_4F0D20
func nox_xxx_rewardFieldGuide_4F0D20(
	marker *C.nox_object_t,
	stage C.uint32_t,
) *C.nox_object_t {
	result := rewardFieldGuideCall4F0D20(
		GetServer().S(),
		asObjectS((*nox_object_t)(marker)),
		uint32(stage),
	)
	return (*C.nox_object_t)(asObjectC(result))
}
