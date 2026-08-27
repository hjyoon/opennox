package legacy

/*
#include "reward_marker_activate_4f0720.h"
*/
import "C"

import "github.com/opennox/opennox/v1/server"

func rewardMarkerActivateRuntime4F0720(s *server.Server) server.RewardMarkerActivateRuntime4F0720 {
	return server.RewardMarkerActivateRuntime4F0720{
		SpellBook: func(marker *server.Object, stage uint32) *server.Object {
			return s.RewardSpellBook4F09F0(marker, stage)
		},
		AbilityBook: func(marker *server.Object, _ uint32) *server.Object {
			return s.RewardAbilityBook4F0C70(marker)
		},
		FieldGuide: func(marker *server.Object, stage uint32) *server.Object {
			return s.RewardFieldGuide4F0D20(marker, stage)
		},
		Weapon: func(marker *server.Object, stage uint32) *server.Object {
			return s.RewardWeapon4F14E0(marker, stage)
		},
		Armor: func(marker *server.Object, stage uint32) *server.Object {
			return s.RewardArmor4F0E80(marker, stage)
		},
		Gem: func(marker *server.Object, stage uint32) *server.Object {
			return s.RewardGem4F1D30(marker, stage)
		},
		Potion: func(marker *server.Object, stage uint32) *server.Object {
			return s.RewardPotion4F1C40(marker, stage)
		},
		Gem2: func(marker *server.Object, stage uint32) *server.Object {
			return s.RewardGem2_4F1F00(marker, stage)
		},
	}
}

//export nox_server_rewardgen_activateMarker_4F0720
func nox_server_rewardgen_activateMarker_4F0720(
	marker *C.nox_object_t,
	stage C.uint32_t,
) *C.nox_object_t {
	s := GetServer().S()
	result := rewardMarkerActivateCall4F0720(
		s,
		asObjectS((*nox_object_t)(marker)),
		uint32(stage),
		rewardMarkerActivateRuntime4F0720(s),
	)
	return (*C.nox_object_t)(asObjectC(result))
}
