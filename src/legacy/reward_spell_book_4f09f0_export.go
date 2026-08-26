package legacy

/*
#include "reward_spell_book_4f09f0.h"
*/
import "C"

import "github.com/opennox/opennox/v1/server"

func rewardSpellBookCall4F09F0(
	s *server.Server,
	marker *server.Object,
	stage uint32,
) *server.Object {
	return s.RewardSpellBook4F09F0(marker, stage)
}

func rewardRandomSlotsCall4F0B60(s *server.Server, stage uint32) uint32 {
	return s.RewardRandomSlots4F0B60(stage)
}

//export nox_xxx_rewardSpellBook_4F09F0
func nox_xxx_rewardSpellBook_4F09F0(
	marker *C.nox_object_t,
	stage C.uint32_t,
) *C.nox_object_t {
	result := rewardSpellBookCall4F09F0(
		GetServer().S(),
		asObjectS((*nox_object_t)(marker)),
		uint32(stage),
	)
	return (*C.nox_object_t)(asObjectC(result))
}

//export nox_server_rewardGen_pickRandomSlots_4F0B60
func nox_server_rewardGen_pickRandomSlots_4F0B60(stage C.uint32_t) C.uint32_t {
	return C.uint32_t(rewardRandomSlotsCall4F0B60(GetServer().S(), uint32(stage)))
}
