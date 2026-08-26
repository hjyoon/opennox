package legacy

/*
#include "reward_ability_book_4f0c70.h"
*/
import "C"

import "github.com/opennox/opennox/v1/server"

func rewardAbilityBookCall4F0C70(
	s *server.Server,
	marker *server.Object,
) *server.Object {
	return s.RewardAbilityBook4F0C70(marker)
}

//export nox_xxx_rewardAbilityBook_4F0C70
func nox_xxx_rewardAbilityBook_4F0C70(marker *C.nox_object_t) *C.nox_object_t {
	result := rewardAbilityBookCall4F0C70(
		GetServer().S(),
		asObjectS((*nox_object_t)(marker)),
	)
	return (*C.nox_object_t)(asObjectC(result))
}
